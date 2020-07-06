package http_clients

import (
	"github.com/opay-org/lib-common/utils/json_rsp_unmarshal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RespMarshaler interface {
	Unmarshal(respBytes []byte, dataPtr interface{}) (errStatus *status.Status)
}

var DefaultMarshaler = &defaultMarshaler{}

type defaultMarshaler struct{}

func (opt *defaultMarshaler) Unmarshal(respBytes []byte, dataPtr interface{}) (errStatus *status.Status) {
	err := json.Unmarshal(respBytes, dataPtr)
	if err != nil {
		errStatus = status.New(codes.Code(json_rsp_unmarshal.CodeDecodeError), err.Error())
	}
	return
}

var StdPbAnyUnmarshaler = &stdPbAnyUnmarshaler{}

type stdPbAnyUnmarshaler struct{}

func (opt *stdPbAnyUnmarshaler) Unmarshal(respBytes []byte, dataPtr interface{}) (errStatus *status.Status) {
	return json_rsp_unmarshal.UnmarshalStdPbAny(respBytes, dataPtr)
}

var StdUnmarshaler = &stdUnmarshaler{}

type stdUnmarshaler struct{}

func (opt *stdUnmarshaler) Unmarshal(respBytes []byte, dataPtr interface{}) (errStatus *status.Status) {
	return json_rsp_unmarshal.UnmarshalStdPbAny(respBytes, dataPtr)
}
