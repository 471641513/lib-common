package clients

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"google.golang.org/grpc/metadata"

	"github.com/opay-org/lib-common/local_context"
	"github.com/opay-org/lib-common/metrics"
	"github.com/opay-org/lib-common/xlog"

	"google.golang.org/grpc"

	"github.com/prometheus/client_golang/prometheus"
)

/**
version: 1.0.3
*/

type TraceContext interface {
	LogId() string
	SetLogId(logid string)

	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}

const ERR_SUCC = "succ"
const ERR_ERR = "err"

const (
	CODE_SUCC = 0
)

type GrpcClientConfig struct {
	Caller         string   `toml:"caller"`
	Addrs          []string `toml:"addrs"`
	SvrName        string   `toml:"svr_name"`
	DialTimeoutMs  int      `toml:"dial_timeout_ms"`
	IdleTimeoutSec int      `toml:"idle_timeout_sec"`
	ReadTimeoutMs  int      `toml:"read_timeout_ms"`
	LongConnection bool     `toml:"long_connection"`

	WriteTimeoutMs   int   `toml:"write_timeout_ms"`
	PoolSize         int   `toml:"pool_size"`
	PoolMaxAliveSec  int64 `toml:"pool_max_alive_sec"`
	KeepAliveSec     int   `toml:"keep_alive_sec"`
	KeepAliveTimeOut int   `toml:"keep_alive_timeout_sec"`
}

func NewGrpcClientBase(conf GrpcClientConfig, dialOpts ...grpc.DialOption) (base *GrpcClientBase, err error) {

	if conf.DialTimeoutMs <= 0 {
		conf.DialTimeoutMs = 200
	}
	if conf.WriteTimeoutMs <= 0 {
		conf.WriteTimeoutMs = 200
	}
	if conf.ReadTimeoutMs <= 0 {
		conf.ReadTimeoutMs = 200
	}
	if conf.IdleTimeoutSec <= 0 {
		conf.IdleTimeoutSec = 60
	}

	if conf.KeepAliveSec <= 0 {
		conf.KeepAliveSec = 30
	}

	if conf.KeepAliveTimeOut <= 0 {
		conf.KeepAliveTimeOut = 10
	}

	if conf.PoolMaxAliveSec <= 0 {
		conf.PoolMaxAliveSec = 60
	}

	// try get addr from svr-addr mgr

	if addrs := getAddrFromSvrMgr(conf.SvrName); addrs != nil {
		if len(addrs.Addrs) > 0 {
			conf.Addrs = addrs.Addrs
			xlog.Info("_GrpcClientBase_init||use addrs from mgr||svrname=[%v]||addrs=%+v",
				conf.SvrName,
				conf.Addrs)
		}
	}

	if len(conf.Addrs) == 0 {
		err = fmt.Errorf("addr is empty||conf=%+v", conf)
	}

	base = &GrpcClientBase{
		conf: conf,
	}
	if !conf.LongConnection {
		base.pool, _ = newShortGrpcClientPool(conf, dialOpts...)
	} else {
		xlog.Info(" _GrpcClientBase_init||long_pool=true||conf=%v", conf)
		base.pool, err = NewGrpcClientPool(conf, dialOpts...)
		if err != nil {
			return nil, err
		}
	}

	/*
		base.pool, err = pool.NewGRPCPool(
			&pool.Options{
				InitTargets:  conf.Addrs,
				InitCap:      conf.InitPoolSize,
				MaxCap:       conf.MaxPoolSize,
				DialTimeout:  time.Millisecond * time.Duration(conf.DialTimeoutMs),
				IdleTimeout:  time.Second * time.Duration(conf.IdleTimeoutSec),
				ReadTimeout:  time.Millisecond * time.Duration(conf.ReadTimeoutMs),
				WriteTimeout: time.Millisecond * time.Duration(conf.WriteTimeoutMs),
			},
			grpc.WithInsecure())
	*/
	return
}

type GrpcClientBase struct {
	conf GrpcClientConfig
	//pool *pool.GRPCPool
	pool GrpcPool
	metrics.MetricsBase
}

func (cli *GrpcClientBase) CreateMetrics(
	prefix string,
	buckets []float64,
	labels []string) *prometheus.HistogramVec {
	return nil
}

func (cli *GrpcClientBase) CreateMetricsV2(
	prefix string,
	buckets []float64,
	timecostLables []string,
	countLables []string) {
	if nil == buckets {
		buckets = []float64{5, 10, 60, 200, 500}
	}

	cli.MetricsBase.CreateMetrics(prefix, buckets, timecostLables)
	cli.MetricsBase.CreateMetricsCountVec(prefix, "grpc", "cnt", countLables)
}

const (
	HEADER_PREFIX = "Grpc-"
	HEADER_TRACE  = "grpc-trace-id"
	HEADER_CALLER = "grpc-caller"
)

func (cli *GrpcClientBase) GetTimeout(parentCtx local_context.TraceContext) (cctx context.Context) {
	mdMap := map[string]string{
		HEADER_TRACE: parentCtx.LogId(),
	}
	if cli.conf.Caller != "" {
		mdMap[HEADER_CALLER] = cli.conf.Caller
	}
	md := metadata.New(mdMap)
	cctx = metadata.NewOutgoingContext(parentCtx, md)
	cctx, _ = context.WithTimeout(cctx, time.Duration(cli.conf.ReadTimeoutMs)*time.Millisecond)
	return
}

func (cli *GrpcClientBase) GetTimeoutFromCtx(parentCtx TraceContext) (cctx context.Context) {
	return cli.GetTimeout(parentCtx)
}

func (cli *GrpcClientBase) Get() (conn *grpc.ClientConn, err error) {
	return cli.pool.Get()
}

func (cli *GrpcClientBase) Put(conn *grpc.ClientConn) {
	_ = cli.pool.Put(conn)
}

func (cli *GrpcClientBase) Conf() GrpcClientConfig {
	return cli.conf
}
func (cli *GrpcClientBase) GetConf() GrpcClientConfig {
	return cli.conf
}
func (cli *GrpcClientBase) GetAddr() string {
	return cli.conf.Addrs[rand.Intn(len(cli.conf.Addrs))]
}

func (cli *GrpcClientBase) Close() {
	cli.pool.Close()
}
