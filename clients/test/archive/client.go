package archive

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/xutils/lib-common/clients"
	"github.com/xutils/lib-common/xlog"
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

func run(base *clients.GrpcClientBase, err error) {

}
