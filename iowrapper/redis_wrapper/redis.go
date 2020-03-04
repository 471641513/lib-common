package redis_wrapper

import (
	"errors"
	"fmt"
	"github.com/opay-org/lib-common/xlog"
	"net"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type RedisConfig struct {
	Addrs    []string `toml:"addrs"`
	Pwd      string   `toml:"pwd"`
	PoolSize int      `toml:"pool_size"`
}

func NewRedisClientWithTimeout(c *RedisConfig, timeout time.Duration) (*redis.Client, error) {
	redisNum := len(c.Addrs)
	if redisNum == 0 {
		return nil, errors.New("redis addrs is empty")
	}
	ch := make(chan []string, redisNum)
	for i := 0; i < redisNum; i++ {
		list := make([]string, redisNum)
		for j := 0; j < redisNum; j++ {
			list[j] = c.Addrs[(i+j)%redisNum]
		}
		ch <- list
	}
	options := &redis.Options{
		Password:     c.Pwd,
		PoolSize:     c.PoolSize,
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		PoolTimeout:  timeout,
		IdleTimeout:  60 * time.Second,
		Dialer: func() (net.Conn, error) {
			list := <-ch
			ch <- list
			for _, addr := range list {
				c, err := net.DialTimeout("tcp", addr, 1000*time.Millisecond)
				if err == nil {
					return c, nil
				}
			}
			return nil, errors.New("all redis down")
		},
	}
	return redis.NewClient(options), nil
}

func NewRedisClient(c *RedisConfig) (*redis.Client, error) {
	return NewRedisClientWithTimeout(c, time.Second)
}

func RedisIncrBy(client *redis.Client, key string, delta int64, expiration int64) (sum int64, err error) {
	sum, err = client.IncrBy(key, delta).Result()
	if err != nil {
		xlog.Error("RedisHncrBy failed||err=%v||key=%v||delta=%v", err, key, delta)
		return
	}
	if expiration > 0 {
		_, err = client.Expire(key, time.Duration(expiration)*time.Second).Result()
		if err != nil {
			xlog.Error("err=%v||key=%v", err, key)
		}
	}
	return
}

func RedisDel(client *redis.Client, keys ...string) int {
	i, err := client.Del(keys...).Result()
	if err != nil {
		xlog.Error("err=%v||keys=%v", err, keys)
	}
	xlog.Debug("redis del||keys=%v||ret=%v", keys, i)
	return int(i)
}

func RedisExpire(client *redis.Client, key string, ttlSec int64) (err error) {
	_, err = client.Expire(key, time.Duration(ttlSec)*time.Second).Result()
	if err != nil {
		xlog.Error("failed RedisExpire|| to err=%v||keys=%v", err, key)
	}
	return
}

func RedisExist(client *redis.Client, key string) bool {
	i, err := client.Exists(key).Result()
	if err != nil {
		xlog.Error("err=%v||keys=%v", err, key)
	}
	return i > 0
}

func RedisSimpleWaitLock(client *redis.Client, key string, ttl time.Duration, lockTimeout time.Duration) bool {
	lockObtained, err := client.SetNX(key, key, ttl).Result()
	if err != nil {
		xlog.Error("RedisSimpleLock failed||err=%v||key=%v||ttl=%v", err, key, ttl)
		return false
	}
	if !lockObtained {
		xlog.Warn("concurrent update on %v occurred, start retry obtain lock", key)
		tick := time.Tick(ttl / 2)
		timeout := time.After(lockTimeout)
		for {
			select {
			case <-tick:
				lockObtained, err = client.SetNX(key, key, ttl).Result()
				if lockObtained {
					return lockObtained
				}
				if err != nil {
					xlog.Error("RedisSimpleLock failed||err=%v||key=%v||ttl=%v", err, key, ttl)
					return false
				}
			case <-timeout:
				break
			}
		}
	}
	return lockObtained
}

func RedisSimpleUnLock(client *redis.Client, key string) {
	_, err := client.Del(key).Result()
	if err != nil {
		xlog.Error("RedisSimpleLock failed||err=%v||key=%v||ttl=%v", err, key)
	}
}

func RedisSet(client *redis.Client, key string, value interface{}, expiration int64) {
	_, err := client.Set(key, value, time.Duration(expiration)*time.Second).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
	}
}

func RedisSetNx(client *redis.Client, key string, value interface{}, expiration int64) (setted bool, err error) {
	setted, err = client.SetNX(key, value, time.Duration(expiration)*time.Second).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
	}
	return
}

func RedisGet(client *redis.Client, key string) string {
	str, err := client.Get(key).Result()
	if err != nil {
		if err != redis.Nil {
			xlog.Error("err=%v||key=%v", err, key)
		}
		return ""
	}
	return str
}

func RedisGeoAdd(client *redis.Client, key string, locs ...[3]float64) (int64, error) {
	geos := make([]*redis.GeoLocation, len(locs))
	for idx, loc := range locs {
		geos[idx] = &redis.GeoLocation{
			Name:      fmt.Sprintf("%.0f", loc[0]),
			Latitude:  loc[1],
			Longitude: loc[2],
		}
	}
	return client.GeoAdd(key, geos...).Result()
}
func RedisGeoRem(client *redis.Client, key string, name string) (int64, error) {
	return client.ZRem(key, name).Result()
}

func RedisGetInt64(client *redis.Client, key string, defaultVal int64) int64 {
	str, err := client.Get(key).Result()
	if err != nil {
		if err != redis.Nil {
			xlog.Error("err=%v||key=%v", err, key)
		}
		return defaultVal
	}
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return defaultVal
	}
	return n
}

func RedisLPushInt64(client *redis.Client, key string, val int64, trimSize int64, expiration int64) {
	_, err := client.LPush(key, val).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
	}

	if trimSize > 0 {
		_, err := client.LTrim(key, 0, trimSize).Result()
		if err != nil {
			xlog.Error("err=%v||key=%v", err, key)
		}
	}
	if expiration > 0 {
		_, err := client.Expire(key, time.Duration(expiration)*time.Second).Result()
		if err != nil {
			xlog.Error("err=%v||key=%v", err, key)
		}
	}
}

func RedisLLen(client *redis.Client, key string) int64 {
	ret, err := client.LLen(key).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
		return 0
	}
	return ret
}

func RedisLIndexInt64(client *redis.Client, key string, index int64) int64 {
	str, err := client.LIndex(key, index).Result()
	if err != nil {
		if err != redis.Nil {
			xlog.Error("err=%v||key=%v", err, key)
		}
		return 0
	}
	val, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
	}
	return val
}

func RedisSAdd(client *redis.Client, key string, members interface{}, expiration int64) {
	_, err := client.SAdd(key, members).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
	}
	if expiration > 0 {
		_, err = client.Expire(key, time.Duration(expiration)*time.Second).Result()
		if err != nil {
			xlog.Error("err=%v||key=%v", err, key)
		}
	}
}

func RedisSRem(client *redis.Client, key string, members ...interface{}) int {
	i, err := client.SRem(key, members...).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
	}
	xlog.Debug("remove||key=%v||members=%v", key, members)
	return int(i)
}
func RedisSMembers(client *redis.Client, key string) []string {
	ret, err := client.SMembers(key).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
		return nil
	}
	return ret
}

func RedisSIsMember(client *redis.Client, key string, member interface{}) bool {
	ret, err := client.SIsMember(key, member).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
		return false
	}
	return ret
}

func RedisZAdd(client *redis.Client, key string, member interface{}, score float64, expiration int64) int64 {
	ret, err := client.ZAdd(key, redis.Z{Score: score, Member: member}).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
		return 0
	}
	if expiration > 0 {
		_, err = client.Expire(key, time.Duration(expiration)*time.Second).Result()
		if err != nil {
			xlog.Error("err=%v||key=%v", err, key)
		}
	}
	return ret
}

func RedisZRem(client *redis.Client, key string, member ...interface{}) (int64, error) {
	ret, err := client.ZRem(key, member...).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
		return 0, err
	}
	return ret, nil
}
func RedisZCard(client *redis.Client, key string) int64 {
	ret, err := client.ZCard(key).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
		return 0
	}
	return ret
}

func RedisZRangeWithScores(client *redis.Client, key string, start, stop int64) []redis.Z {
	ret, err := client.ZRangeWithScores(key, start, stop).Result()
	xlog.Debug("zrange %v %v %v withscores", key, start, stop)
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
		return nil
	}
	return ret
}

func RedisZRangeByScoreWithScores(client *redis.Client, key string, min, max interface{}) []redis.Z {
	ret, err := client.ZRangeByScoreWithScores(key, redis.ZRangeBy{
		Min: fmt.Sprintf("%v", min),
		Max: fmt.Sprintf("%v", max)}).Result()
	if err != nil {
		xlog.Error("err=%v||key=%v", err, key)
		return nil
	}
	return ret
}

func RedisHMSet(client *redis.Client, key string, kvMap map[string]interface{}) (err error) {
	_, err = client.HMSet(key, kvMap).Result()
	if err != nil {
		xlog.Error("RedisHMSet failed||err=%v||key=%v||kvMap=%v", err, key, kvMap)
	}
	return
}

func RedisHMGet(client *redis.Client, key string, fields []string) (rslt map[string]interface{}, err error) {
	rslt = map[string]interface{}{}
	result, err := client.HMGet(key, fields...).Result()
	if len(fields) != len(result) {
		err = fmt.Errorf("result fields length mismatch||len(fields)=%v||len(result)=%v",
			len(fields), len(result))
	}
	if err != nil {
		xlog.Error("RedisHMSet failed||err=%v||key=%v||fields=%v", err, key, fields)
		return
	}
	for idx, field := range fields {
		rslt[field] = result[idx]
	}
	return
}

func RedisScan(client *redis.Client, cursor uint64, match string, cnt int64) (page []string) {
	page, cursor, err := client.Scan(cursor, match, cnt).Result()
	if err != nil {
		xlog.Error("Scan failed||err=%v", err)
		return
	}
	return page
}
