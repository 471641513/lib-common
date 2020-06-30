package local_context

import (
	"context"
	"time"

	"github.com/opay-org/lib-common/utils"
)

type LocalContext struct {
	context.Context
	data   map[string]interface{}
	logid  string
	method string
}

func (ctx *LocalContext) Method() string {
	return ctx.method
}
func (ctx *LocalContext) SetMethod(method string) {
	ctx.method = method
}
func (ctx *LocalContext) LogId() string {
	return ctx.logid
}
func (ctx *LocalContext) SetLogId(logid string) {
	ctx.logid = logid
}
func (ctx *LocalContext) Put(key string, data interface{}) {
	ctx.data[key] = data
}
func (ctx *LocalContext) Get(key string) (data interface{}) {
	return ctx.data[key]
}

func NewLocalContext() *LocalContext {
	return &LocalContext{
		Context: context.Background(),
		data:    map[string]interface{}{},
		logid:   utils.GenerateUid(),
	}
}
func NewLocalContextWithCtx(ctx context.Context) *LocalContext {
	return &LocalContext{
		Context: ctx,
		data:    map[string]interface{}{},
		logid:   utils.GenerateUid(),
	}
}
func NewLocalContextWithTrace(logid string) *LocalContext {
	return &LocalContext{
		Context: context.Background(),
		data:    map[string]interface{}{},
		logid:   logid,
	}
}

type TraceContext interface {
	LogId() string
	SetLogId(logid string)

	Deadline() (deadline time.Time, ok bool)
	Done() <-chan struct{}
	Err() error
	Value(key interface{}) interface{}
}
