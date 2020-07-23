package iredis

import (
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/gomodule/redigo/redis"
)

func CreatePool(address string) Pool {
	return NewPool(&redis.Pool{
		MaxIdle:     300,
		IdleTimeout: time.Minute,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", address, redis.DialDatabase(0), redis.DialConnectTimeout(500*time.Millisecond), redis.DialReadTimeout(500*time.Millisecond), redis.DialWriteTimeout(500*time.Millisecond))
		},
		TestOnBorrow: func(c redis.Conn, ti time.Time) error {
			if time.Since(ti) < time.Minute {
				return nil
			}
			return ToError("Redis connection timed out")
		},
	})
}

type Pool interface {
	Get() Conn
}

type Conn interface {
	redis.Conn
}

func NewPool(p *redis.Pool) Pool {
	return &pool{p}
}

type pool struct {
	pool *redis.Pool
}

func (p *pool) Get() Conn {
	return p.pool.Get()
}
