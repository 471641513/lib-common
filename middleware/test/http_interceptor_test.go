package test

import (
	"os"
	"test_proto"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/opay-org/lib-common/xlog"

	"github.com/golang/protobuf/proto"

	"github.com/opay-org/lib-common/middleware"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
)

func TestMain(m *testing.M) {
	xlog.SetupLogDefault()
	// setup code...
	code := m.Run()
	// teardown code...
	xlog.Close()
	os.Exit(code)
}

func TestStandardResponsMarshaler(t *testing.T) {
	m := middleware.StandardResponsMarshaler{
		Marshaler: &runtime.JSONBuiltin{},
	}

	var data proto.Message

	data = &test_proto.Data{
		UserList: []*test_proto.Data_User{
			{
				Id:   1,
				Name: "test",
			}, {
				Id:       2,
				UserName: "test2",
				Name:     "test22",
			},
		},
	}
	dataBytes, err := m.Marshal(data)
	xlog.Info("data=%s||err=%v", dataBytes, err)
	assert.Nil(t, err)
	data2 := &test_proto.Data{}
	err = m.Unmarshal(dataBytes, data2)
	assert.Nil(t, err)
	xlog.Info("data2=%+v", data2)
	dataBytes = []byte(`{"code":"200","message":"success","data":{"@type":"type.googleapis.com/test_proto.Data", "user_list":[{"id":1,"name":"test","user_name":""},{"id":2,"name":"test22","user_name":"test2"}]}}`)
	data33 := &test_proto.Data{}
	err = m.Unmarshal(dataBytes, data33)
	assert.Nil(t, err)
	xlog.Info("data3=%+v", data33)

	dataBytes = []byte(`{"code":"200","message":"success","data":{"user_list":[{"id":1,"name":"test","user_name":""},{"id":2,"name":"test22","user_name":"test2"}]}}`)
	data3 := &test_proto.Data{}
	err = m.Unmarshal(dataBytes, data3)
	assert.Nil(t, err)
	xlog.Info("data3=%+v", data3)

	//data4 := &test_proto.Data{}
	//dataBytes3 := []byte(`{"user_list":[{"id":"1","name":"test","user_name":""},{"id":"2","name":"test22","user_name":"test2"}]}`)
	//err = middleware.ProtoUnmarshalOptions.Unmarshal(dataBytes3, data4)
	//xlog.Info("data4=%+v,err=%v", data4, err)
}
