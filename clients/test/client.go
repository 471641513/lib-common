package main

import (
	"biz-common/clients"
	"biz-common/clients/test/test_proto"
	"context"
	"flag"
	"github.com/opay-org/lib-common/xlog"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

var clientCnt = flag.Int("c", 10, "routine cnt")

func main() {
	flag.Parse()
	xlog.SetupLogDefault()
	defer xlog.Close()
	conf := clients.GrpcClientConfig{
		Addrs: []string{
			"127.0.0.1:13333",
			"127.0.0.1:13331",
			"127.0.0.1:13333",
			"127.0.0.1:13333",
			"127.0.0.1:13333",
		},
		LongConnection:  true,
		PoolMaxAliveSec: 30,
		PoolSize:        100,
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool, err := clients.NewGrpcClientBase(conf)
	if err != nil {
		xlog.Error("failed to init client base=%+v", err)
		return
	}
	wg := &sync.WaitGroup{}
	for i := 0; i < *clientCnt; i++ {
		go func() {
			wg.Add(1)
			defer wg.Done()
			run(pool, ctx)
		}()
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-exit
	xlog.Info("sig:%v is received, start to quit safely", sig)
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
	defer func() {
		if err != nil {
			xlog.Error("err=%+v", err)
		}
	}()
	conn, err = pool.Get()
	if err != nil {
		xlog.Error("failed to get conn=%v", err)
		return
	}
	defer pool.Put(conn)
	ctx, _ := context.WithTimeout(context.Background(), time.Millisecond*300)
	_, err = test_proto.NewTestStubClient(conn).AddLocs(ctx, &test_proto.Req{
		Id: int64(rand.Int()),
	})
}
