package autocertcache

import (
	"context"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/gomodule/redigo/redis"
	"golang.org/x/crypto/acme/autocert"
)

const (
	keyPrefix = "acme_autocert_"
	daySecs   = 86400
)

func Dir(dir string) autocert.Cache {
	return autocert.DirCache(dir)
}

func Redis(p iredis.Pool) autocert.Cache {
	return &redisCache{pool: p}
}

type redisCache struct {
	pool iredis.Pool
}

func (c *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	cnn := c.pool.Get()
	defer cnn.Close()

	if err := cnn.Send("MULTI"); err != nil {
		return nil, err
	}

	if err := cnn.Send("GET", keyPrefix+key); err != nil {
		return nil, err
	}

	// set to expire after 1 day after last read
	if err := cnn.Send("EXPIRE", keyPrefix+key, daySecs); err != nil {
		return nil, err
	}

	results, err := redis.Values(cnn.Do("EXEC"))
	if err != nil {
		return nil, err
	}

	return redis.Bytes(results[len(results)-1], err)
}

func (c *redisCache) Put(ctx context.Context, key string, data []byte) error {
	cnn := c.pool.Get()
	defer cnn.Close()

	_, err := cnn.Do("SETEX", keyPrefix+key, daySecs, data)
	return ToError(err)
}

func (c *redisCache) Delete(ctx context.Context, key string) error {
	cnn := c.pool.Get()
	defer cnn.Close()

	_, err := cnn.Do("DEL", keyPrefix+key)
	return ToError(err)
}
