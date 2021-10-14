package test

import (
	"context"
	"errors"
	"reflect"
	"test_proto"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/xutils/lib-common/clients"
	"google.golang.org/grpc/metadata"

	"github.com/xutils/lib-common/local_context"
	"github.com/xutils/lib-common/xlog"

	"github.com/xutils/lib-common/middleware"

	"github.com/xutils/lib-common/metrics"
	"google.golang.org/grpc"
)

func TestGrpcInterceptor(t *testing.T) {
	m := &metrics.MetricsBase{}
	middleware.InitRpcMetrics(m, "unitTest")

	var ifunc func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (rsp interface{}, err error)
	ifunc = middleware.GrpcInterceptor(*m,
		middleware.OptEnsureTrace(func(req reflect.Type) middleware.TraceIface {
			trace := &test_proto.Trace{}
			return trace
		}),
		middleware.OptEnsureError(func(req reflect.Type) middleware.ErrorIface {
			err := &test_proto.Error{}
			return err
		}))

	info := &grpc.UnaryServerInfo{
		FullMethod: "test.Method",
	}
	defaultAssert := func(t *testing.T, rsp interface{}, err error) {
		xlog.Info("rsp=%+v||err=%v", rsp, err)
	}
	type args struct {
		ctx context.Context
		req interface{}
	}
	tests := []struct {
		name    string
		args    args
		handler func(t *testing.T, ctx context.Context, req interface{}) (interface{}, error)
		assert  func(t *testing.T, rsp interface{}, err error)
	}{
		{
			name: "parse trace from header",
			args: args{
				req: &test_proto.Req{
					Id:    10000,
					Trace: &test_proto.Trace{},
				},
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					clients.HEADER_CALLER: {"test-header-caller"},
					clients.HEADER_TRACE:  {"test-header-trace"},
				}),
			},
			handler: func(t *testing.T, ctx context.Context, req interface{}) (rsp interface{}, e error) {
				lctx, ok := ctx.(*local_context.LocalContext)
				xlog.Info("lctx=%+v||ok=%v", lctx, ok)
				xlog.Info("req=%+v", req)
				r := req.(*test_proto.Req)
				assert.Equal(t, lctx.LogId(), "test-header-trace")
				assert.Equal(t, r.Trace.Caller, "test-header-caller")
				assert.Equal(t, r.Trace.TraceId, "test-header-trace")
				rsp = &test_proto.Rsp{}
				return
			},
			assert: defaultAssert,
		}, {
			name: "parse trace from header with trace nil",
			args: args{
				req: &test_proto.Req{
					Id: 10000,
				},
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					clients.HEADER_CALLER: {"test-header-caller"},
					clients.HEADER_TRACE:  {"test-header-trace"},
				}),
			},
			handler: func(t *testing.T, ctx context.Context, req interface{}) (rsp interface{}, e error) {
				lctx, ok := ctx.(*local_context.LocalContext)
				xlog.Info("lctx=%+v||ok=%v", lctx, ok)
				xlog.Info("req=%+v", req)
				r := req.(*test_proto.Req)
				assert.Equal(t, lctx.LogId(), "test-header-trace")
				assert.NotNil(t, r.Trace)
				assert.Equal(t, r.Trace.Caller, "test-header-caller")
				assert.Equal(t, r.Trace.TraceId, "test-header-trace")
				rsp = &test_proto.Rsp{}
				return
			},
			assert: func(t *testing.T, rsp interface{}, err error) {
				xlog.Info("rsp=%+v||err=%v", rsp, err)
				assert.Nil(t, err)
			},
		}, {
			name: "parse trace from trace",
			args: args{
				req: &test_proto.Req{
					Id: 10000,
					Trace: &test_proto.Trace{
						TraceId: "test-trace",
						Caller:  "test-caller",
					},
				},
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{
					clients.HEADER_CALLER: {"test-header-caller"},
					clients.HEADER_TRACE:  {"test-header-trace"},
				}),
			},
			handler: func(t *testing.T, ctx context.Context, req interface{}) (rsp interface{}, e error) {
				lctx, ok := ctx.(*local_context.LocalContext)
				xlog.Info("lctx=%+v||ok=%v", lctx, ok)
				xlog.Info("req=%+v", req)
				r := req.(*test_proto.Req)
				assert.Equal(t, lctx.LogId(), "test-trace")
				assert.Equal(t, r.Trace.Caller, "test-caller")
				assert.Equal(t, r.Trace.TraceId, "test-trace")
				rsp = &test_proto.Rsp{
					Error: &test_proto.Error{
						Code: 100,
					},
				}
				return
			},
			assert: defaultAssert,
		}, {
			name: "parse trace from trace",
			args: args{
				req: &test_proto.Req{
					Id: 10000,
					Trace: &test_proto.Trace{
						TraceId: "test-trace",
						Caller:  "test-caller",
					},
				},
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{}),
			},
			handler: func(t *testing.T, ctx context.Context, req interface{}) (rsp interface{}, e error) {
				lctx, ok := ctx.(*local_context.LocalContext)
				xlog.Info("lctx=%+v||ok=%v", lctx, ok)
				xlog.Info("req=%+v", req)
				r := req.(*test_proto.Req)
				assert.Equal(t, lctx.LogId(), "test-trace")
				assert.Equal(t, r.Trace.Caller, "test-caller")
				assert.Equal(t, r.Trace.TraceId, "test-trace")
				rsp = &test_proto.Rsp{
					Error: &test_proto.Error{
						Code: 100,
					},
				}
				e = errors.New("test")
				return
			},
			assert: defaultAssert,
		}, {
			name: "parse req with no trace",
			args: args{
				req: &struct {
					Id int64
				}{
					Id: 10000,
				},
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{}),
			},
			handler: func(t *testing.T, ctx context.Context, req interface{}) (rsp interface{}, e error) {
				lctx, ok := ctx.(*local_context.LocalContext)
				xlog.Info("lctx=%+v||ok=%v", lctx, ok)
				xlog.Info("req=%+v", req)
				rsp = &struct {
					Code int64
				}{}
				return
			},
			assert: func(t *testing.T, rsp interface{}, err error) {
				xlog.Info("rsp=%+v||err=%v", rsp, err)
				assert.Nil(t, err)
			},
		}, {
			name: "parse req with no trace and default error",
			args: args{
				req: &struct {
					Id int64
				}{
					Id: 10000,
				},
				ctx: metadata.NewIncomingContext(context.Background(), map[string][]string{}),
			},
			handler: func(t *testing.T, ctx context.Context, req interface{}) (rsp interface{}, e error) {
				lctx, ok := ctx.(*local_context.LocalContext)
				xlog.Info("lctx=%+v||ok=%v", lctx, ok)
				xlog.Info("req=%+v", req)
				rsp = &test_proto.Rsp{}
				return
			},
			assert: func(t *testing.T, r interface{}, err error) {
				xlog.Info("rsp=%+v||err=%v", r, err)
				rsp, ok := r.(*test_proto.Rsp)
				assert.True(t, ok)
				assert.Nil(t, err)
				assert.NotNil(t, rsp.Error)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rsp, err := ifunc(tt.args.ctx,
				tt.args.req,
				info,
				func(ctx context.Context, req interface{}) (interface{}, error) {
					return tt.handler(t, ctx, req)
				})
			tt.assert(t, rsp, err)
		})
	}
}
