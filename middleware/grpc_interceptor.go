package middleware

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/opay-org/lib-common/utils"

	"github.com/opay-org/lib-common/clients"
	"google.golang.org/grpc/metadata"

	"github.com/opay-org/lib-common/xlog"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/opay-org/lib-common/local_context"
	"github.com/opay-org/lib-common/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func DefaultGrpcOptions() (opts []grpc.ServerOption) {
	opts = append(opts,
		grpc.KeepaliveParams(keepalive.ServerParameters{}),
		grpc.MaxConcurrentStreams(10000))
	return
}

func InitRpcMetrics(metrics *metrics.MetricsBase, prefix string) {
	metrics.CreateMetrics(fmt.Sprintf("%v_rpc", prefix), nil, []string{"method"})
	metrics.CreateMetricsCountVec(prefix, "rpc", "cnt", []string{"method", "err", "caller"})
	prometheus.MustRegister(metrics.GetMetricsVectors()...)
}

type TraceIface interface {
	GetTraceId() string
	GetCaller() string
}

type Trace struct {
	TraceId string
	Caller  string
}

type Error struct {
	Code    int64
	Message string
}

type ErrorIface interface {
	GetCode() int64
	GetMessage() string
}

var GrpcMethodReg = regexp.MustCompile(`\/([^\/]*)$`)

func init() {}

// parse trace from md
func ParseTraceAndCaller(ctx context.Context, tracedContext local_context.TraceContext) (traceId string, caller string) {
	if tracedContext != nil {
		md, ok := metadata.FromIncomingContext(ctx)
		xlog.Debug("md=%+v", md)
		if ok && md != nil {
			l := md.Get(clients.HEADER_TRACE)
			if len(l) > 0 {
				traceId = l[0]
			}
			l = md.Get(clients.HEADER_CALLER)
			if len(l) > 0 {
				caller = l[0]
			}
		}
		if traceId != "" {
			tracedContext.SetLogId(traceId)
		}
	}
	return
}

func GrpcInterceptorServerOption(
	metrics metrics.MetricsBase,
	opt ...GrpcInterceptorOpt) (serverOpt grpc.ServerOption) {
	return grpc.UnaryInterceptor(GrpcInterceptor(metrics, opt...))
}

func setCaller(trace interface{}, caller string) {
	val := reflect.ValueOf(trace).Elem().FieldByName("Caller")
	if val.Type().Kind() == reflect.String {
		val.Set(reflect.ValueOf(caller))
	}
}
func setTraceId(trace interface{}, traceId string) {
	val := reflect.ValueOf(trace).Elem().FieldByName("TraceId")
	if val.Type().Kind() == reflect.String {
		val.Set(reflect.ValueOf(traceId))
	}

}

type GrpcInterceptorOpt interface{}

type optEnsureTrace func(req reflect.Type) TraceIface

func OptEnsureTrace(defaultTrace func(req reflect.Type) TraceIface) GrpcInterceptorOpt {
	return GrpcInterceptorOpt(optEnsureTrace(defaultTrace))
}

type optEnsureError func(req reflect.Type) ErrorIface

func OptEnsureError(defaultError func(req reflect.Type) ErrorIface) GrpcInterceptorOpt {
	return GrpcInterceptorOpt(optEnsureError(defaultError))
}

func OptInnerInterceptor(innerInterceptor grpc.UnaryServerInterceptor) GrpcInterceptorOpt {
	return GrpcInterceptorOpt(innerInterceptor)
}

const (
	fieldTrace = "Trace"
	fieldError = "Error"
)

func GrpcInterceptor(
	metrics metrics.MetricsBase,
	opts ...GrpcInterceptorOpt) (interceptor grpc.UnaryServerInterceptor) {

	var ensureTraceFunc optEnsureTrace = nil
	var innerInterceptor grpc.UnaryServerInterceptor = nil
	var ensureErrorFunc optEnsureError = nil

	for _, opt := range opts {
		if newTraceFunc, ok := opt.(optEnsureTrace); ok {
			ensureTraceFunc = newTraceFunc
			xlog.Info("ensureTraceFunc registered")
		}
		if newErrorFunc, ok := opt.(optEnsureError); ok {
			ensureErrorFunc = newErrorFunc
			xlog.Info("ensureErrorFunc registered")
		}
		if inner, ok := opt.(grpc.UnaryServerInterceptor); ok {
			innerInterceptor = inner
			xlog.Info("inner interceptor registered")
		}
	}
	//var interceptor grpc.UnaryServerInterceptor
	interceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (rsp interface{}, err error) {
		lctx := local_context.NewLocalContextWithCtx(ctx)
		method := info.FullMethod
		subStrs := GrpcMethodReg.FindStringSubmatch(method)
		if len(subStrs) > 1 {
			method = subStrs[1]
		}
		lctx.SetMethod(method)
		t0 := time.Now()
		defer func() {
			if e := recover(); e != nil {
				xlog.Fatal("_grpc_recover||logid=%v||method=%v||catch panic||%s\n%s", lctx.LogId(), method, e, debug.Stack())
				err = fmt.Errorf("panic err=%v", e)
				return
			}
		}()
		// 0.parse trace and caller from header
		traceId, caller := ParseTraceAndCaller(ctx, lctx)
		xlog.Debug("trace_id=%v||caller=%v", traceId, caller)
		if ensureTraceFunc != nil {
			// 0.1 compatible to request with trace object
			traceRef := reflect.ValueOf(req).Elem().FieldByName(fieldTrace)
			xlog.Debug("traceRef=%v||req=%+v||method=%v||ctx=%+v", traceRef, req, method, ctx)

			if traceRef.IsValid() && traceRef.Type().Kind() == reflect.Ptr {
				if traceRef.IsNil() && ensureTraceFunc != nil {
					newTrace := ensureTraceFunc(traceRef.Type())
					// 0.1 compare filed type and set default value
					traceRefType, _ := reflect.TypeOf(req).Elem().FieldByName(fieldTrace)
					if reflect.TypeOf(newTrace) == traceRefType.Type {
						traceRef.Set(reflect.ValueOf(newTrace))
					}
				}

				if !traceRef.IsNil() {
					reqTrace, ok := traceRef.Interface().(TraceIface)
					xlog.Info("traceRef=%+v", ok)
					if ok && reqTrace != nil {
						if reqTrace.GetCaller() == "" && caller != "" {
							setCaller(reqTrace, caller)
						}
						if reqTrace.GetTraceId() != "" {
							lctx.SetLogId(reqTrace.GetTraceId())
						} else {
							if reqTrace.GetTraceId() == "" && traceId != "" {
								setTraceId(reqTrace, traceId)
							}
						}
					}
				}
			}
		}
		// 1. common metrics
		defer func() {
			timecost := utils.CalTimecost(t0)
			metrics.Observe(timecost, method)
			xlog.Fatal("TIMECOST=%v", timecost)
			errType := ""
			if err != nil {
				errType = clients.ERR_ERR
			} else {
				// 0.1 compatible to request with error object
				if rsp != nil && ensureErrorFunc != nil {
					refErr := reflect.ValueOf(rsp).Elem().FieldByName(fieldError)
					if refErr.IsValid() && refErr.Type().Kind() == reflect.Ptr {
						if !refErr.IsNil() {
							Error, ok := refErr.Interface().(ErrorIface)
							if ok {
								if Error.GetCode() != clients.CODE_SUCC {
									errType = strconv.Itoa(int(Error.GetCode()))
								}
							}
						} else if ensureErrorFunc != nil {
							// 1.1 assign default error Msg
							newError := ensureErrorFunc(refErr.Type())
							// 0.1 compare filed type and set default value
							errorRefType, _ := reflect.TypeOf(rsp).Elem().FieldByName(fieldError)
							if reflect.TypeOf(newError) == errorRefType.Type {
								refErr.Set(reflect.ValueOf(newError))
							}
						}
					}

				}
			}
			if errType != "" {
				metrics.ObserveCounter(1, method, errType, caller)
			} else {
				metrics.ObserveCounter(1, method, clients.ERR_SUCC, caller)
			}
		}()

		if innerInterceptor != nil {
			return innerInterceptor(lctx, req, info, handler)
		}
		return handler(lctx, req)
	}
	return interceptor
	//return grpc.UnaryInterceptor(interceptor)
}
