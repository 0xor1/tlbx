package redis

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type tlbxKey struct {
	name string
}

func Mware(name string, pool iredis.Pool) func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Set(tlbxKey{name}, &poolWrapper{tlbx: tlbx, name: name, pool: pool})
	}
}

func Get(tlbx app.Tlbx, name string) Pool {
	return tlbx.Get(tlbxKey{name}).(Pool)
}

type Pool interface {
	Base() iredis.Pool
	Get() Conn
}

type poolWrapper struct {
	tlbx app.Tlbx
	name string
	pool iredis.Pool
}

func (w *poolWrapper) Base() iredis.Pool {
	return w.pool
}

func (w *poolWrapper) Get() Conn {
	return &connWrapper{
		tlbx: w.tlbx,
		name: w.name,
		conn: w.pool.Get(),
	}
}

type Conn interface {
	Base() iredis.Conn
	iredis.Conn
}

type connWrapper struct {
	tlbx app.Tlbx
	name string
	conn iredis.Conn
}

func (w *connWrapper) Base() iredis.Conn {
	return w.conn
}

func (w *connWrapper) Close() error {
	return w.conn.Close()
}

func (w *connWrapper) Err() error {
	return w.conn.Err()
}

func (w *connWrapper) Do(cmd string, args ...interface{}) (reply interface{}, err error) {
	w.do(func(q string, a ...interface{}) { reply, err = w.conn.Do(cmd, args...) }, cmd, args...)
	return
}

func (w *connWrapper) Send(cmd string, args ...interface{}) (err error) {
	w.do(func(q string, a ...interface{}) { err = w.conn.Send(cmd, args...) }, cmd, args...)
	return
}

func (w *connWrapper) Flush() error {
	return w.conn.Flush()
}

func (w *connWrapper) Receive() (reply interface{}, err error) {
	return w.conn.Receive()
}

func (w *connWrapper) do(do func(string, ...interface{}), cmd string, args ...interface{}) {
	start := NowUnixMilli()
	do(cmd, args...)
	w.tlbx.LogActionStats(&app.ActionStats{
		Milli:  NowUnixMilli() - start,
		Type:   "REDIS",
		Name:   w.name,
		Action: Str(append([]interface{}{cmd, " ", args[0], " ..."})...),
	})
}
