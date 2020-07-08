package test_http_svr

import (
	"net"
	"net/http"
	"reflect"
	"test_proto"

	"google.golang.org/grpc"

	"github.com/opay-org/lib-common/local_context"

	"github.com/opay-org/lib-common/metrics"

	"github.com/opay-org/lib-common/middleware"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/opay-org/lib-common/xlog"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type handler struct {
	test_proto.UnimplementedTestStub2Server
}

const (
	caseUnimplemented  = 1
	caseSuccReturn     = 2
	caseCustomizedCode = 10001
)

func (*handler) TestApi2(ctx context.Context, req *test_proto.ReqWithoutTrace) (*test_proto.Data, error) {
	xlog.Info("ctx_type=%v", reflect.TypeOf(ctx))
	lctx, _ := ctx.(*local_context.LocalContext)
	xlog.Info("logid=%v", lctx.LogId())
	if req.Id == caseUnimplemented {
		return nil, status.Errorf(codes.Unimplemented, "method TestApi2 not implemented")
	}
	if req.Id == caseCustomizedCode {
		return nil, status.Error(codes.Code(caseCustomizedCode), "customized error")
	}
	return &test_proto.Data{UserList: []*test_proto.Data_User{{
		Id:   req.Id,
		Name: "testname",
	},
	}}, nil
}

func NewHttpStub(ctx context.Context, listen string, grpcListen string) (err error) {
	interceptor := middleware.GrpcInterceptor(metrics.MetricsBase{})
	h := &handler{}

	grpcOpts := middleware.DefaultGrpcOptions()
	grpcOpts = append(grpcOpts,
		grpc.UnaryInterceptor(interceptor))
	s := grpc.NewServer(grpcOpts...)
	test_proto.RegisterTestStub2Server(s, h)
	lis, err := net.Listen("tcp", grpcListen)
	if err != nil {
		xlog.Fatal("[tcp] listen to %v||err=%v", grpcListen, err)
		return
	}
	go s.Serve(lis)

	httpOpts := []runtime.ServeMuxOption{
		middleware.HttpMarshalerServerMuxOption(),
		middleware.TracedIncomingHeaderMatcherMuxOption(),
	}
	mux := runtime.NewServeMux(httpOpts...)
	err = test_proto.RegisterTestStub2HandlerServer(ctx, mux, h, interceptor)
	if err != nil {
		xlog.Fatal("failed to register test stub||err=%v", err)
		return
	}
	err = http.ListenAndServe(listen, middleware.DefaultHttpWrapper(mux))
	if err != nil {
		xlog.Fatal("failed to listen and serve http||err=%v", err)
	}
	return
}
