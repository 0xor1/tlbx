package service

import (
	"database/sql"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/email"
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/store"
	"github.com/0xor1/wtf/pkg/web/app"
)

type tlbxKey struct{}

type Layer interface {
	Cache() RedisPoolLogClient
	User() SqlLogClient
	Pwd() SqlLogClient
	Data() SqlLogClient
	Email() email.Client
	Store() store.Client
}

func Mware(cache iredis.Pool, user, pwd, data isql.ReplicaSet, email email.Client, store store.Client) func(app.Toolbox) {
	return func(tlbx app.Toolbox) {
		tlbx.Set(tlbxKey{}, &service{
			cache: &redisPoolLogWrapper{tlbx: tlbx, pool: cache},
			user:  &sqlLogWrapper{tlbx: tlbx, sql: user},
			pwd:   &sqlLogWrapper{tlbx: tlbx, sql: pwd},
			data:  &sqlLogWrapper{tlbx: tlbx, sql: data},
			email: email,
			store: store,
		})
	}
}

func Get(tlbx app.Toolbox) Layer {
	return tlbx.Get(tlbxKey{}).(Layer)
}

type service struct {
	cache *redisPoolLogWrapper
	user  *sqlLogWrapper
	pwd   *sqlLogWrapper
	data  *sqlLogWrapper
	email email.Client
	store store.Client
}

func (d *service) Cache() RedisPoolLogClient {
	return d.cache
}

func (d *service) User() SqlLogClient {
	return d.user
}

func (d *service) Pwd() SqlLogClient {
	return d.pwd
}

func (d *service) Data() SqlLogClient {
	return d.data
}

func (d *service) Email() email.Client {
	return d.email
}

func (d *service) Store() store.Client {
	return d.store
}

type SqlLogClient interface {
	Base() isql.ReplicaSet
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (isql.Rows, error)
	QueryRow(query string, args ...interface{}) isql.Row
}

type sqlLogWrapper struct {
	tlbx app.Toolbox
	sql  isql.ReplicaSet
}

func (w *sqlLogWrapper) Base() isql.ReplicaSet {
	return w.sql
}

func (w *sqlLogWrapper) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	w.do(func(q string) { res, err = w.sql.Primary().ExecContext(w.tlbx.Ctx(), q, args...) }, query)
	return
}

func (w *sqlLogWrapper) Query(query string, args ...interface{}) (rows isql.Rows, err error) {
	w.do(func(q string) { rows, err = w.sql.RandSlave().QueryContext(w.tlbx.Ctx(), q, args...) }, query)
	return
}

func (w *sqlLogWrapper) QueryRow(query string, args ...interface{}) (row isql.Row) {
	w.do(func(q string) { row = w.sql.RandSlave().QueryRowContext(w.tlbx.Ctx(), q, args...) }, query)
	return
}

func (w *sqlLogWrapper) do(do func(string), query string) {
	// no query should ever even come close to 1 second in execution time
	start := NowUnixMilli()
	do(`SET STATEMENT max_statement_time=1 FOR ` + query)
	w.tlbx.LogQueryStats(&app.QueryStats{
		Milli: NowUnixMilli() - start,
		Query: query,
	})
}

type RedisPoolLogClient interface {
	Base() iredis.Pool
	Get() RedisConnLogClient
}

type redisPoolLogWrapper struct {
	tlbx app.Toolbox
	pool iredis.Pool
}

func (w *redisPoolLogWrapper) Base() iredis.Pool {
	return w.pool
}

func (w *redisPoolLogWrapper) Get() RedisConnLogClient {
	return &redisConnLogWrapper{
		tlbx: w.tlbx,
		conn: w.pool.Get(),
	}
}

type RedisConnLogClient interface {
	Base() iredis.Conn
	iredis.Conn
}

type redisConnLogWrapper struct {
	tlbx app.Toolbox
	conn iredis.Conn
}

func (w *redisConnLogWrapper) Base() iredis.Conn {
	return w.conn
}

func (w *redisConnLogWrapper) Close() error {
	return w.conn.Close()
}

func (w *redisConnLogWrapper) Err() error {
	return w.conn.Err()
}

func (w *redisConnLogWrapper) Do(cmd string, args ...interface{}) (reply interface{}, err error) {
	w.do(func(q string, a ...interface{}) { reply, err = w.conn.Do(cmd, args...) }, cmd, args...)
	return
}

func (w *redisConnLogWrapper) Send(cmd string, args ...interface{}) (err error) {
	w.do(func(q string, a ...interface{}) { err = w.conn.Send(cmd, args...) }, cmd, args...)
	return
}

func (w *redisConnLogWrapper) Flush() error {
	return w.conn.Flush()
}

func (w *redisConnLogWrapper) Receive() (reply interface{}, err error) {
	return w.conn.Receive()
}

func (w *redisConnLogWrapper) do(do func(string, ...interface{}), cmd string, args ...interface{}) {
	start := NowUnixMilli()
	do(cmd, args...)
	w.tlbx.LogQueryStats(&app.QueryStats{
		Milli: NowUnixMilli() - start,
		Query: Sprint(append([]interface{}{cmd, " "}, args...)...),
	})
}
