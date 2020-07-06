package middleware

import (
	"google.golang.org/genproto/googleapis/rpc/status"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	jsoniter "github.com/json-iterator/go"
	"github.com/opay-org/lib-common/utils/json_rsp_unmarshal"
	"google.golang.org/protobuf/encoding/protojson"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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

// @override
func (m *StandardResponsMarshaler) Marshal(v interface{}) (data []byte, err error) {
	if rspErr, ok := v.(*status.Status); ok {
		data, err = m.Marshaler.Marshal(rspErr)
		return
	}
	vProtoMsg, ok := v.(proto.Message)
	if !ok {
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
