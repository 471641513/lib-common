package archive

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"test_proto"

	"github.com/xutils/lib-common/local_context"

	"github.com/xutils/lib-common/metrics"

	"github.com/xutils/lib-common/middleware"

	"github.com/xutils/lib-common/xlog"

	"google.golang.org/grpc"
)

type stub struct {
	test_proto.UnimplementedTestStubServer
}

func (s *stub) AddLocs(ctx context.Context, req *test_proto.Req) (*test_proto.Rsp, error) {
	lctx := ctx.(*local_context.LocalContext)
	xlog.Info("req=%+v||lctx=%+v", req, lctx)
	return &test_proto.Rsp{}, nil
}

var listen = flag.String("p", "13333", "port")

func main() {
	flag.Parse()
	//listen := ":13333"
	xlog.SetupLogDefault()
	xlog.Info("listen to %v", *listen)
	options := middleware.DefaultGrpcOptions()
	options = append(options,
		middleware.GrpcInterceptorServerOption(metrics.MetricsBase{}, nil))
	s := grpc.NewServer(options...)
	handler := &stub{}
	test_proto.RegisterTestStubServer(s, handler)

	// set up server
	lis, err := net.Listen("tcp", ":"+*listen)
	if err != nil {
		xlog.Error("e=failed to set up server||err=%v", err)
		return
	}
	go s.Serve(lis)

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)
	<-exit

}
