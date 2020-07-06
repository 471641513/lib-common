package test_http_svr

import (
	"fmt"
	"os"
	"test_proto"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/opay-org/lib-common/local_context"

	"github.com/opay-org/lib-common/http_clients"
	"github.com/opay-org/lib-common/xlog"
	"golang.org/x/net/context"
)

const (
	port = ":8007"
)

var httpCli *http_clients.HttpClient

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
	ctx := context.Background()
	go func() {
		err = NewHttpStub(ctx, port)
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
	},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rspBytes, err := httpCli.PostJsonBody(ctx, "/test/stub2", tt.args.req, tt.args.rsp, http_clients.StdPbAnyUnmarshaler)
			xlog.Debug("err=%v||rsp=%s", err, rspBytes)
			tt.assert(t, tt.args.rsp, err)
		})
	}
}
