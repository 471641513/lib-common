package test

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"test_proto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/opay-org/lib-common/local_context"

	"github.com/opay-org/lib-common/metrics"
	"github.com/opay-org/lib-common/middleware"

	"github.com/opay-org/lib-common/utils"

	"github.com/opay-org/lib-common/clients"

	"github.com/opay-org/lib-common/xlog"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {
	xlog.SetupLogDefault()
	ctx, cancel := context.WithCancel(context.Background())
	go startStub(ctx)
	// setup code...
	code := m.Run()
	// teardown code...
	xlog.Close()
	cancel()
	time.Sleep(time.Second)
	os.Exit(code)
}

const testPort = 11113
const clientN = 1

type stub struct {
	test_proto.UnimplementedTestStubServer
}

func (s *stub) AddLocs(ctx context.Context, req *test_proto.Req) (*test_proto.Rsp, error) {
	xlog.Info("req=%+v", req)
	return &test_proto.Rsp{}, nil
}

func startStub(ctx context.Context) {
	options := middleware.DefaultGrpcOptions()
	options = append(options,
		middleware.GrpcInterceptorServerOption(metrics.MetricsBase{}, nil))
	s := grpc.NewServer(options...)
	handler := &stub{}
	test_proto.RegisterTestStubServer(s, handler)
	// set up server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%v", testPort))
	if err != nil {
		xlog.Error("e=failed to set up server||err=%v", err)
		return
	}
	xlog.Info("stub start listen :%v", testPort)
	go s.Serve(lis)
	select {
	case <-ctx.Done():
	}
}

func Test_Trace(t *testing.T) {

	time.Sleep(time.Second)
	conf := clients.GrpcClientConfig{
		Addrs: []string{
			fmt.Sprintf("127.0.0.1:%v", testPort),
		},
		PoolMaxAliveSec: 5,
		PoolSize:        10,
		ReadTimeoutMs:   300,
	}
	cli, err := clients.NewGrpcClientBase(conf)
	assert.Nil(t, err)
	doReq(cli)

	time.Sleep(time.Second)

}

func tTest_Pool(t *testing.T) {

	time.Sleep(time.Second)
	// test short conn pool
	testPool(t, false)

	// test long conn
	testPool(t, true)

	time.Sleep(time.Second)
}

func testPool(t *testing.T, LongConnection bool) {
	ctx, cancel := context.WithCancel(context.Background())
	conf := clients.GrpcClientConfig{
		Addrs: []string{
			fmt.Sprintf("127.0.0.1:%v", testPort),
			fmt.Sprintf("127.0.0.1:%v", testPort),
			fmt.Sprintf("127.0.0.1:%v", testPort+1),
		},
		LongConnection:  LongConnection,
		PoolMaxAliveSec: 5,
		PoolSize:        10,
		ReadTimeoutMs:   300,
	}
	pool, err := clients.NewGrpcClientBase(conf)
	if err != nil {
		xlog.Error("failed to init client base=%+v", err)
		return
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < clientN; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()
			run(pool, ctx)
		}()
	}
	time.Sleep(time.Second * 10)
	cancel()
	wg.Wait()
}

func run(pool *clients.GrpcClientBase, ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			doReq(pool)
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func doReq(pool *clients.GrpcClientBase) {
	var err error
	var conn *grpc.ClientConn
	t0 := time.Now()
	defer func() {
		tCost := utils.CalTimecost(t0)
		if err != nil {
			xlog.Error("err=%+v", err)
		} else {
			xlog.Info("t_cost=%v", tCost)
		}
	}()
	conn, err = pool.Get()
	if err != nil {
		xlog.Error("failed to get conn=%v", err)
		return
	}
	defer pool.Put(conn)
	ctx := local_context.NewLocalContext()
	cctx := pool.GetTimeout(ctx)
	_, err = test_proto.NewTestStubClient(conn).AddLocs(cctx, &test_proto.Req{
		Id:    int64(rand.Int()),
		Trace: &test_proto.Trace{},
	})
}
