package json_rsp_unmarshal

import (
	"fmt"

	"github.com/golang/protobuf/ptypes/any"

	"google.golang.org/protobuf/reflect/protoreflect"

	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/protobuf/encoding/protojson"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var ProtoUnmarshalOptions = protojson.UnmarshalOptions{
	AllowPartial: true,
}
var ProtoMarshalOptions = protojson.MarshalOptions{
	UseProtoNames:   true,
	EmitUnpopulated: true,
}

const (
	CodeSucc        = int32(200)
	CodeDecodeError = int32(100000)
)

func NewResponse(data *any.Any) (rsp *Response) {
	rsp = &Response{
		Code:    CodeSucc,
		Message: "success",
		Data:    data,
	}
	return
}

/**
* this is faster
e.g.:
{"code":"200", "message":"success", "data":{"@type":"type.googleapis.com/test_proto.Data", "user_list":[{"id":"1", "name":"test", "user_name":""}, {"id":"2", "name":"test22", "user_name":"test2"}]}}
*/
func UnmarshalStdPbAny(data []byte, v interface{}) (errStatus *status.Status) {
	var err error
	defer func() {
		if err != nil && errStatus == nil {
			errStatus = status.New(codes.Code(CodeDecodeError), err.Error())
		}
	}()
	vProtoMsg, ok := v.(proto.Message)
	if !ok {
		err = fmt.Errorf("illegal v type")
		return
	}
	rspv := &Response{}
	// 1. marshalRsp
	err = ProtoUnmarshalOptions.Unmarshal(data, rspv)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal rsp||err=%v", err.Error())
		return
	}
	// 2. build status error
	if rspv.Code != int32(codes.OK) && rspv.Code != CodeSucc {
		errStatus = status.New(codes.Code(rspv.Code), rspv.Message)
		return
	}
	// 3. read any
	err = ptypes.UnmarshalAny(rspv.Data, vProtoMsg)
	if err != nil {
		err = fmt.Errorf("failed to UnmarshalAny||err=%v", err.Error())
		return
	}
	return
}

/**
e.g.:
{"code":"200", "message":"success", "data":{"user_list":[{"id":"1", "name":"test", "user_name":""}, {"id":"2", "name":"test22", "user_name":"test2"}]}}
@type is not given
*/

func UnmarshalStd(data []byte, v interface{}) (errStatus *status.Status) {
	var err error
	defer func() {
		if err != nil && errStatus == nil {
			errStatus = status.New(codes.Code(CodeDecodeError), err.Error())
		}
	}()
	jsonObj := json.Get(data)
	retCode := jsonObj.Get("code").ToInt32()
	if retCode != int32(codes.OK) && retCode != CodeSucc {
		errStatus = status.New(codes.Code(retCode), jsonObj.Get("message").ToString())
		return
	}
	if vProtoWithReflect, ok := v.(protoreflect.ProtoMessage); ok {
		err = ProtoUnmarshalOptions.Unmarshal([]byte(jsonObj.Get("data").ToString()), vProtoWithReflect)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal vProtoWithReflect||err=%v", err.Error())
			return
		}
	} else {
		err = json.UnmarshalFromString(jsonObj.Get("data").ToString(), v)
		if err != nil {
			err = fmt.Errorf("failed to unmarshal vProtoWithReflect||err=%v", err.Error())
			return
		}
	}

	return
}
