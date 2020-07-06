package test_http_svr

import (
	"net/http"
	"test_proto"

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
	caseUnimplemented = 1
	caseSuccReturn    = 2
)

func (*handler) TestApi2(ctx context.Context, req *test_proto.ReqWithoutTrace) (*test_proto.Data, error) {
	if req.Id == caseUnimplemented {
		return nil, status.Errorf(codes.Unimplemented, "method TestApi2 not implemented")
	}
	return &test_proto.Data{UserList: []*test_proto.Data_User{{
		Id:   req.Id,
		Name: "testname",
	},
	}}, nil
}

func NewHttpStub(ctx context.Context, listen string) (err error) {
	interceptor := middleware.GrpcInterceptor(metrics.MetricsBase{})
	httpOpts := []runtime.ServeMuxOption{
		middleware.HttpMarshalerServerMuxOption(),
	}
	mux := runtime.NewServeMux(httpOpts...)
	h := &handler{}
	err = test_proto.RegisterTestStub2HandlerServer(ctx, mux, h, interceptor)
	if err != nil {
		xlog.Fatal("failed to register test stub||err=%v", err)
		return
	}
	err = http.ListenAndServe(listen, mux)
	if err != nil {
		xlog.Fatal("failed to listen and serve http||err=%v", err)
	}
	return
}
