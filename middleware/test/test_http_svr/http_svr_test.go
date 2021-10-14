package test_http_svr

import (
	"fmt"
	"os"
	"test_proto"
	"testing"
	"time"

	"google.golang.org/grpc/codes"

	"github.com/xutils/lib-common/clients"
	"google.golang.org/grpc/status"

	"github.com/stretchr/testify/assert"

	"github.com/xutils/lib-common/local_context"

	"github.com/xutils/lib-common/http_clients"
	"github.com/xutils/lib-common/xlog"
	"golang.org/x/net/context"
)

const (
	port     = ":8007"
	grpcPort = ":18007"
)

var httpCli *http_clients.HttpClient

var grpcCli *clients.GrpcClientBase

func initEnv() (err error) {
	httpCli, err = http_clients.NewHttpClient(http_clients.HttpClientConfig{
		Addrs: []string{
			fmt.Sprintf("127.0.0.1%s", port),
		},
		Caller: "test",
	})
	if err != nil {
		return
	}
	grpcCli, err = clients.NewGrpcClientBase(clients.GrpcClientConfig{
		Caller: "test",
		Addrs: []string{
			fmt.Sprintf("127.0.0.1%s", grpcPort),
		},
	})
	if err != nil {
		return
	}
	ctx := context.Background()
	go func() {
		err = NewHttpStub(ctx, port, grpcPort)
		xlog.Debug("err=%+v", err)
		time.Sleep(time.Second)
	}()
	time.Sleep(time.Second)
	return
}

func TestMain(m *testing.M) {
	xlog.SetupLogDefault()
	err := initEnv()
	if err != nil {
		panic(err)
	}
	// setup code...
	code := m.Run()
	// teardown code...
	xlog.Close()
	os.Exit(code)

}

func Test_NewHttpStub(t *testing.T) {
	ctx := local_context.NewLocalContext()
	type args struct {
		req *test_proto.ReqWithoutTrace
		rsp interface{}
	}
	tests := []struct {
		name    string
		args    args
		assert  func(t *testing.T, rsp interface{}, err error)
		wantErr bool
	}{{
		args: args{
			req: &test_proto.ReqWithoutTrace{
				Id: caseUnimplemented,
			},
			rsp: &test_proto.Data{},
		},
		assert: func(t *testing.T, rsp interface{}, err error) {
			xlog.Info("rsp=%+v||err=%+v", rsp, err)
			assert.NotNil(t, err)
			statusRsp, ok := status.FromError(err)
			if ok {
				xlog.Fatal("CODE=%+v", statusRsp.Code())
			} else {
				t.Fatal(t)
			}
			assert.Equal(t, int32(codes.Unimplemented), int32(statusRsp.Code()), "err=%+v", err)
		},
	}, {
		args: args{
			req: &test_proto.ReqWithoutTrace{
				Id: caseSuccReturn,
			},
			rsp: &test_proto.Data{},
		},
		assert: func(t *testing.T, rsp interface{}, err error) {
			xlog.Info("rsp=%+v||err=%+v", rsp, err)
			assert.Nil(t, err)
		},
	}, {
		args: args{
			req: &test_proto.ReqWithoutTrace{
				Id: caseCustomizedCode,
			},
			rsp: &test_proto.Data{},
		},
		assert: func(t *testing.T, rsp interface{}, err error) {
			xlog.Info("rsp=%+v||err=%+v", rsp, err)
			assert.NotNil(t, err)
			statusRsp, ok := status.FromError(err)
			if ok {
				xlog.Fatal("CODE=%+v", statusRsp.Code())
			} else {
				t.Fatal(t)
			}
			assert.Equal(t, int32(caseCustomizedCode), int32(statusRsp.Code()), "err=%+v", err)
		},
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rspBytes, err := httpCli.PostJsonBody(ctx, "/test/stub2", tt.args.req, tt.args.rsp, http_clients.StdPbAnyUnmarshaler)
			xlog.Debug("err=%v||rsp=%s", err, rspBytes)
			tt.assert(t, tt.args.rsp, err)

			// do grpc req
			conn, err := grpcCli.Get()
			cctx := grpcCli.GetTimeout(ctx)
			rsp, err := test_proto.NewTestStub2Client(conn).TestApi2(cctx, tt.args.req)
			xlog.Debug("err=%v||rsp=%v", err, rsp)

			tt.assert(t, rsp, err)
			grpcCli.Put(conn)
		})
	}
}
