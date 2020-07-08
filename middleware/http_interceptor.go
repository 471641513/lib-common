package middleware

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/opay-org/lib-common/clients"

	status2 "google.golang.org/grpc/status"

	"google.golang.org/grpc/codes"

	"github.com/opay-org/lib-common/local_context"
	"github.com/opay-org/lib-common/xlog"

	"google.golang.org/genproto/googleapis/rpc/status"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	jsoniter "github.com/json-iterator/go"
	"github.com/opay-org/lib-common/utils/json_rsp_unmarshal"
	"google.golang.org/protobuf/encoding/protojson"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func TracedIncomingHeaderMatcherMuxOption() (serverMuxOpt runtime.ServeMuxOption) {
	return runtime.WithIncomingHeaderMatcher(tracedIncomingHeaderMatcher)
}

func tracedIncomingHeaderMatcher(in string) (out string, match bool) {
	out, match = runtime.DefaultHeaderMatcher(in)
	if match {
		return
	}
	if strings.HasPrefix(in, clients.HEADER_PREFIX) {
		out = in
		match = true
	}
	return
}

type HttpInterceptorOpt interface{}

func HttpMarshalerServerMuxOption(
	opt ...HttpInterceptorOpt) (serverMuxOpt runtime.ServeMuxOption) {
	marshaler := HttpMarshaler(opt...)
	return runtime.WithMarshalerOption(runtime.MIMEWildcard, marshaler)

}

func HttpMarshaler(opts ...HttpInterceptorOpt) (marshaler runtime.Marshaler) {
	marshaler = &runtime.JSONBuiltin{}

	marshaler = &StandardResponsMarshaler{
		Marshaler: marshaler,
	}
	return
}

var ProtoUnmarshalOptions = protojson.UnmarshalOptions{
	AllowPartial: true,
}
var ProtoMarshalOptions = protojson.MarshalOptions{
	UseProtoNames:   true,
	EmitUnpopulated: true,
}

type StandardResponsMarshaler struct {
	runtime.Marshaler
}

type rspErr interface {
	GetMessage() string
	GetCode() int32
}

// @override
func (m *StandardResponsMarshaler) Marshal(v interface{}) (data []byte, err error) {
	defer func() {
		xlog.Debug("data=%s", data)
	}()
	if e, ok := v.(rspErr); ok {
		rspErr := &json_rsp_unmarshal.Response{
			Code:    e.GetCode(),
			Message: e.GetMessage(),
		}
		data, err = m.Marshaler.Marshal(rspErr)
		return
	}
	if rspErr, ok := v.(*status.Status); ok {
		data, err = m.Marshaler.Marshal(rspErr)
		return
	} else {
	}
	vProtoMsg, ok := v.(proto.Message)
	if !ok {
		//xlog.Info("not proto message")
		return m.Marshaler.Marshal(v)
	}
	ret, err := ptypes.MarshalAny(vProtoMsg)
	rspv := json_rsp_unmarshal.NewResponse(ret)
	return ProtoMarshalOptions.Marshal(rspv)
}

// @override
func (m *StandardResponsMarshaler) Unmarshal(data []byte, v interface{}) (err error) {
	errStatus := json_rsp_unmarshal.UnmarshalStdPbAny(data, v)
	if errStatus != nil && int32(errStatus.Code()) == json_rsp_unmarshal.CodeDecodeError {
		errStatus = json_rsp_unmarshal.UnmarshalStd(data, v)
	}
	if errStatus != nil {
		err = errStatus.Err()
	}
	return
}

func DefaultHttpWrapper(h http.Handler) (handler http.Handler) {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lctx := local_context.NewLocalContext()
		// copy request
		body, err := copyReqBody(lctx, r)
		//xlog.Debug("header=%+v||body=%s", r.Header, body)
		if err != nil {
			xlog.Error("logid=%v||failed to copy req||err=%v", lctx.LogId(), err)
		} else {
			xlog.Debug("logid=%v||body=%s", lctx.LogId(), body)
		}
		r = r.WithContext(context.WithValue(r.Context(), _body, body))
		allowCORS(w, r)
		h.ServeHTTP(w, r)
	})
}

func copyReqBody(lctx *local_context.LocalContext, req *http.Request) (body []byte, err error) {
	body, err = ioutil.ReadAll(req.Body)
	if err != nil {
		xlog.Error("logid=%v||failed to read req body||err=%v", lctx.LogId(), err)
		err = status2.New(codes.Internal, err.Error()).Err()
		return
	}
	req.Body = ioutil.NopCloser(bytes.NewReader(body))
	return
}

func allowCORS(w http.ResponseWriter, r *http.Request) {
	if origin := r.Header.Get("Origin"); origin != "" {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
			headers := []string{"Content-Type",
				"Accept",
				"Authorization",
				clients.HEADER_CALLER,
				clients.HEADER_TRACE}
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
			methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE"}
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
		}
	}
}

const (
	_body = "_body"
)
