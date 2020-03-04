package local_context

import (
	"context"
	"github.com/opay-org/lib-common/utils"
)

type LocalContext struct {
	context.Context
	data  map[string]interface{}
	logid string
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