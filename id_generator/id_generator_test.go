package id_generator

import (
	"github.com/xutils/lib-common/xlog"
	"testing"
)

func TestNewIdGenerator(t *testing.T) {
	xlog.SetupLogDefault()
	defer xlog.Close()
	/*
		redisCli, err := redis_wrapper.NewRedisClient(&redis_wrapper.RedisConfig{
			Addrs: []string{"127.0.0.1:6379"},
		})*/
	//assert.NilError(t, err)
	gotSf, err := NewIdGenerator()
	xlog.Info("sf=%+v, err=%+v", gotSf, err)
	id, err := gotSf.NextID()
	xlog.Info("id=%v||err=%v", id, err)
	xlog.Info("id=156619121358790310||err=%v", err)
}
