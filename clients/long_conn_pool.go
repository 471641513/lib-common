package clients

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/keepalive"

	"github.com/xutils/lib-common/xlog"

	"github.com/smallnest/weighted"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

/*
####################################################################################
USE LONG CONNECTION POOL
note: single connection in grpc is highly performanced already
*/

//reference: http://xiaorui.cc/2019/08/13/golang-grpc%e7%bd%91%e5%85%b3%e7%94%a8%e8%bf%9e%e6%8e%a5%e6%b1%a0%e6%8f%90%e9%ab%98%e5%90%9e%e5%90%90%e9%87%8f/

type longConnSlice []*longConn

func (s longConnSlice) Len() int           { return len(s) }
func (s longConnSlice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s longConnSlice) Less(i, j int) bool { return s[i].ts < s[j].ts }

type longConn struct {
	conn     *grpc.ClientConn
	ts       int64
	addrStat *addrStats
}

func newLongConn(conn *grpc.ClientConn, stats *addrStats) (c *longConn) {
	c = &longConn{
		conn:     conn,
		ts:       time.Now().Unix(),
		addrStat: stats,
	}
	return
}

var (
	ErrConnShutdown = errors.New("grpc conn shutdown")
	ErrLongConnNil  = errors.New("long conn wrapper nil")
	ErrConnNil      = errors.New("grpc conn wrapper nil")
)

const MaxNext = 1000000000

const MaxReconnectCnt = 10

const UNHEALTH_LOAD_SCORE = 2000000

type addrStats struct {
	loadScore *int64
	idx       int
	addr      string
}

func newAddrStats(idx int, addr string) *addrStats {
	return &addrStats{
		loadScore: new(int64),
		idx:       idx,
		addr:      addr,
	}
}

type GrpcClientPool struct {
	conf     GrpcClientConfig
	mtx      *sync.Mutex
	mapMtx   *sync.RWMutex
	next     int64
	capacity int64
	conns    []*longConn
	ctx      context.Context
	cancel   context.CancelFunc
	w        *weighted.SW

	connStats []*addrStats

	dialOpts []grpc.DialOption
}

func NewGrpcClientPool(conf GrpcClientConfig, opt ...grpc.DialOption) (pool *GrpcClientPool, err error) {

	if conf.PoolSize < len(conf.Addrs) {
		conf.PoolSize = len(conf.Addrs)
	}

	pool = &GrpcClientPool{
		conf:     conf,
		mtx:      &sync.Mutex{},
		mapMtx:   &sync.RWMutex{},
		capacity: int64(conf.PoolSize),
		conns:    make([]*longConn, conf.PoolSize),
		w:        &weighted.SW{},

		dialOpts: opt,
	}

	if len(pool.dialOpts) == 0 {
		pool.dialOpts = []grpc.DialOption{
			grpc.WithInsecure(),
		}
	}
	pool.dialOpts = append(pool.dialOpts,
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:    time.Duration(pool.conf.KeepAliveSec) * time.Second,
			Timeout: time.Duration(pool.conf.KeepAliveTimeOut) * time.Second,
		}))

	pool.connStats = make([]*addrStats, len(conf.Addrs))

	// init conn counts
	for idx, _ := range pool.connStats {
		stat := newAddrStats(idx, conf.Addrs[idx])
		pool.w.Add(stat, 1)
		pool.connStats[idx] = stat
	}

	//xlog.Info("addr_stat=%+v", pool.connStats)

	pool.ctx, pool.cancel = context.WithCancel(context.Background())
	err = pool.init()
	go pool.balanceWorker()
	return
}

func (pool *GrpcClientPool) debugConn(action string, conn *grpc.ClientConn) {
	//xlog.Debug("act=%v||conn=%p", action, conn)
}

func (pool *GrpcClientPool) balanceWorker() {
	ticker := time.Tick(time.Second * 5)
	connToRecycle := []*longConn{}
	for {
		select {
		case <-ticker:

			connToRecycle = pool.checkLongConns()
			time.Sleep(time.Second)
			xlog.Debug("lconns to recycle=%v", len(connToRecycle))
			//close long conns to do recycle from last round
			for _, lconn := range connToRecycle {
				if lconn != nil && lconn.conn != nil {
					_ = lconn.conn.Close()
					pool.debugConn("close", lconn.conn)
				}
			}
		// do ttl check
		case <-pool.ctx.Done():
			return
		}
	}
}

func (pool *GrpcClientPool) checkLongConns() (connToRecycle []*longConn) {
	nowTs := time.Now().Unix()
	idxToReconnect := map[*longConn]int{}
	idxToNewConnect := []int{}
	connSlice := longConnSlice{}
	healthCountMap := map[int]int64{}

	for idx, lconn := range pool.conns {
		if lconn == nil || lconn.addrStat == nil {
			xlog.Warn("idx=%v||lconn nil or addrStat nil||lconn=%+v", idx, lconn)
			idxToNewConnect = append(idxToNewConnect, idx)
			continue
		}
		if _, ok := healthCountMap[lconn.addrStat.idx]; ok {
			healthCountMap[lconn.addrStat.idx] += 1
		} else {
			healthCountMap[lconn.addrStat.idx] = 1
		}
		if nowTs-lconn.ts > pool.conf.PoolMaxAliveSec {
			//xlog.Debug("idx=%v||addr=%v||timemout||ts=%v", connIdx, lconn.addrStat.addr, lconn.ts)
			idxToReconnect[lconn] = idx
			connSlice = append(connSlice, lconn)
		}
	}

	//xlog.Info("[%v]||healthCountMap=%v", pool.conf.Addrs[0], healthCountMap)
	// check health conn count
	for addIdx, stat := range pool.connStats {
		if *(stat.loadScore) < UNHEALTH_LOAD_SCORE {
			// reset health conn count
			atomic.StoreInt64(stat.loadScore, healthCountMap[addIdx])
		} else {
			// try get long connection
			conn, err := pool.connectAddr(stat.addr)
			if err != nil {
				xlog.Warn("try connect addr and still failed||idx=%v||addr=%v||err=%+v",
					stat.idx, stat.addr, err)
			} else {
				xlog.Info("try connect add||recovered||idx=%v||addr=%v||err=%+v",
					stat.idx, stat.addr, err)
				atomic.StoreInt64(stat.loadScore, 0)
				_ = conn.Close()
			}
		}
		//xlog.Debug("addidx=%v||stat=%v", addIdx, *(stat.loadScore))
	}

	sort.Sort(connSlice)

	// get the oldest ten

	if len(connSlice) > MaxReconnectCnt {
		connSlice = connSlice[:MaxReconnectCnt]
	}

	newConns := map[int]*longConn{}

	for _, idx := range idxToNewConnect {
		lconn, err := pool.connect(0)
		if err != nil {
			xlog.Error("failed to connect||err=%v", err)
			continue
		}
		newConns[idx] = lconn
	}

	for _, oldConn := range connSlice {
		idx := idxToReconnect[oldConn]
		lconn, err := pool.connect(0)
		if err != nil {
			xlog.Error("failed to reConnect||err=%v", err)
			continue
		}
		newConns[idx] = lconn
	}

	pool.mtx.Lock()
	defer pool.mtx.Unlock()
	for idx, newLconn := range newConns {
		connToRecycle = append(connToRecycle, pool.conns[idx])
		pool.conns[idx] = newLconn
	}
	return
}

func (pool *GrpcClientPool) init() (err error) {
	tsOffsetStep := int64(math.Round(float64(pool.conf.PoolMaxAliveSec) / float64(pool.conf.PoolSize+1)))
	if tsOffsetStep == 0 {
		tsOffsetStep = 1
	}
	xlog.Debug("max_alive=%v||pool_size=%v||ts_offset_step=%v",
		pool.conf.PoolMaxAliveSec, pool.conf.PoolSize, tsOffsetStep)

	for idx, _ := range pool.conns {
		lconn, err := pool.connect(0)
		if err != nil {
			return err
		}
		lconn.ts = lconn.ts - pool.conf.PoolMaxAliveSec + tsOffsetStep*int64(idx)
		pool.conns[idx] = lconn
	}
	return nil
}

func (pool *GrpcClientPool) checkState(lconn *longConn) error {
	if lconn == nil {
		return ErrLongConnNil
	}
	if lconn.conn == nil {
		return ErrConnNil
	}

	state := lconn.conn.GetState()
	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		_ = lconn.conn.Close()
		return ErrConnShutdown
	}
	return nil
}

func (pool *GrpcClientPool) randAddr() (stat *addrStats) {
	for i := 0; i < len(pool.conf.Addrs); i++ {
		stat, _ = pool.w.Next().(*addrStats)
		if *(stat.loadScore) < UNHEALTH_LOAD_SCORE {
			return stat
		}
	}
	return
}

func (pool *GrpcClientPool) connect(retry int) (*longConn, error) {
	addrStat := pool.randAddr()
	if addrStat == nil {
		return nil, fmt.Errorf("failed to get addr")
	}
	addr, _ := addrStat.addr, addrStat.idx
	atomic.AddInt64(addrStat.loadScore, 1)
	//xlog.Debug("create connect||addrIdx=%v||stat=%v", addrIdx, *(addrStat.loadScore))
	conn, err := pool.connectAddr(addr)
	if err != nil {
		atomic.AddInt64(addrStat.loadScore, UNHEALTH_LOAD_SCORE)
		if retry < 1 {
			time.Sleep(time.Millisecond * 100)
			return pool.connect(retry + 1)
		}
		xlog.Error("failed to connect to %v||retry=%v||err=%v", addr, retry, err)
		return nil, err
	}
	lconn := newLongConn(conn, addrStat)
	return lconn, nil
}

func (pool *GrpcClientPool) connectAddr(addr string) (conn *grpc.ClientConn, err error) {
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*500)
	conn, err = grpc.DialContext(
		ctx,
		addr,
		pool.dialOpts...,
	)
	if conn != nil {
		pool.debugConn("connect", conn)
	}
	return
}

func (pool *GrpcClientPool) getConn() (conn *grpc.ClientConn, err error) {
	var (
		idx  int64
		next int64
	)
	var lconn *longConn

	next = atomic.AddInt64(&pool.next, 1)
	if next > MaxNext {
		atomic.SwapInt64(&pool.next, 0)
	}
	idx = next % pool.capacity
	lconn = pool.conns[idx]

	if err = pool.checkState(lconn); err == nil && lconn != nil {
		return lconn.conn, nil
	} else {
		xlog.Warn("conn stat err=%+v", err)
	}

	// gc old conn
	if lconn != nil && lconn.conn != nil {
		_ = lconn.conn.Close()
		pool.debugConn("close", lconn.conn)
	}

	pool.mtx.Lock()
	defer pool.mtx.Unlock()
	// check if is inited already
	lconn = pool.conns[idx]
	if err = pool.checkState(lconn); err == nil && lconn != nil {
		return lconn.conn, nil
	}

	lconn, err = pool.connect(0)
	if err != nil {
		return nil, err
	}
	pool.conns[idx] = lconn
	return lconn.conn, nil
}

func (pool *GrpcClientPool) Get() (conn *grpc.ClientConn, err error) {

	for i := 0; i < get_conn_retry; i++ {
		conn, err = pool.getConn()
		if conn != nil {
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
	if err != nil {
		xlog.Error("failed to get connection after retry")
	}
	return
}

func (pool *GrpcClientPool) Put(conn *grpc.ClientConn) error {

	return nil
}

func (pool *GrpcClientPool) Close() {
	pool.cancel()
	pool.mtx.Lock()
	defer pool.mtx.Unlock()
	for _, lconn := range pool.conns {
		if lconn == nil || lconn.conn == nil {
			continue
		}
		pool.debugConn("close", lconn.conn)
		_ = lconn.conn.Close()
	}
}
