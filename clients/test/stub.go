package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"
	"syscall"
	"test_proto"

	"github.com/opay-org/lib-common/xlog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type stub struct {
	test_proto.UnimplementedTestStubServer
}

func (s *stub) AddLocs(ctx context.Context, req *test_proto.Req) (*test_proto.Rsp, error) {
	return &test_proto.Rsp{}, nil
}

var listen = flag.String("p", "13333", "port")

func main() {
	flag.Parse()
	//listen := ":13333"
	xlog.SetupLogDefault()
	xlog.Info("listen to %v", *listen)
	s := grpc.NewServer(
		grpc.KeepaliveParams(keepalive.ServerParameters{}),
		grpc.MaxConcurrentStreams(10000),
	)
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
