package model

import (
	"encoding/json"
	"time"

	"github.com/go-redis/redis"
)

type BaseCacheModel struct {
	cache  *redis.Client
	expire time.Duration
}

func NewBaseCacheModel(cache *redis.Client, expire time.Duration) (m *BaseCacheModel) {
	m = &BaseCacheModel{
		cache:  cache,
		expire: expire,
	}
	return
}

func (m *BaseCacheModel) Get(key string, data interface{}) (bool, error) {
	cache := m.cache

	result, err := cache.Get(key).Result()

	if err == redis.Nil {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	err = json.Unmarshal([]byte(result), data)
	if err != nil {
		return false, err
	}

	return true, nil
}
func (m *BaseCacheModel) Cache() (client *redis.Client) {
	return m.cache
}

func (m *BaseCacheModel) Set(key string, data interface{}) error {
	return m.SetEx(key, data, m.expire)
}

func (m *BaseCacheModel) SetEx(key string, data interface{}, expire time.Duration) error {
	cache := m.cache

	result, err := json.Marshal(data)

	if err != nil {
		return err
	}

	return cache.Set(key, result, expire).Err()
}

func (m *BaseCacheModel) Del(key string) error {
	cache := m.cache
	return cache.Del(key).Err()
}
