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
	Cache() RedisPoolClient
	User() SqlClient
	Pwd() SqlClient
	Data() SqlClient
	Email() email.Client
	Store() store.Client
}

func Mware(cache iredis.Pool, user, pwd, data isql.ReplicaSet, email email.Client, store store.Client) func(app.Toolbox) {
	return func(tlbx app.Toolbox) {
		tlbx.Set(tlbxKey{}, &service{
			cache: &redisPoolWrapper{tlbx: tlbx, pool: cache},
			user:  &sqlClient{tlbx: tlbx, sql: user},
			pwd:   &sqlClient{tlbx: tlbx, sql: pwd},
			data:  &sqlClient{tlbx: tlbx, sql: data},
			email: email,
			store: store,
		})
	}
}

func Get(tlbx app.Toolbox) Layer {
	return tlbx.Get(tlbxKey{}).(Layer)
}

type service struct {
	cache *redisPoolWrapper
	user  *sqlClient
	pwd   *sqlClient
	data  *sqlClient
	email email.Client
	store store.Client
}

func (d *service) Cache() RedisPoolClient {
	return d.cache
}

func (d *service) User() SqlClient {
	return d.user
}

func (d *service) Pwd() SqlClient {
	return d.pwd
}

func (d *service) Data() SqlClient {
	return d.data
}

func (d *service) Email() email.Client {
	return d.email
}

func (d *service) Store() store.Client {
	return d.store
}

type SqlClient interface {
	Base() isql.ReplicaSet
	Begin() Tx
	SqlClientCore
}

type SqlClientCore interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(rowsFn func(isql.Rows), query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) isql.Row
}

type Tx interface {
	SqlClientCore
	Rollback()
	Commit()
}

type tx struct {
	tx        isql.Tx
	tlbx      app.Toolbox
	sqlClient *sqlClient
	done      bool
}

func (t *tx) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	t.sqlClient.do(func(q string) { res, err = t.tx.ExecContext(t.tlbx.Ctx(), q, args...) }, query)
	return
}

func (t *tx) Query(rowsFn func(isql.Rows), query string, args ...interface{}) (err error) {
	t.sqlClient.do(func(q string) {
		var rows isql.Rows
		rows, err = t.tx.QueryContext(t.tlbx.Ctx(), q, args...)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (t *tx) QueryRow(query string, args ...interface{}) (row isql.Row) {
	t.sqlClient.do(func(q string) { row = t.tx.QueryRowContext(t.tlbx.Ctx(), q, args...) }, query)
	return
}

func (t *tx) Rollback() {
	if !t.done {
		t.sqlClient.do(func(q string) {
			err := t.tx.Rollback()
			if err != nil && err != sql.ErrTxDone {
				PanicOn(err)
			}
			t.done = true
		}, "ROLLBACK")
	}
}

func (t *tx) Commit() {
	t.sqlClient.do(func(q string) { PanicOn(t.tx.Commit()); t.done = true }, "COMMIT")
}

type sqlClient struct {
	tlbx app.Toolbox
	sql  isql.ReplicaSet
}

func (c *sqlClient) Base() isql.ReplicaSet {
	return c.sql
}

func (c *sqlClient) Begin() Tx {
	t, err := c.sql.Primary().Begin()
	PanicOn(err)
	c.do(func(s string) {}, "START TRANSACTION")
	return &tx{
		tx:        t,
		tlbx:      c.tlbx,
		sqlClient: c,
	}
}

func (c *sqlClient) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	c.do(func(q string) { res, err = c.sql.Primary().ExecContext(c.tlbx.Ctx(), q, args...) }, query)
	return
}

func (c *sqlClient) Query(rowsFn func(isql.Rows), query string, args ...interface{}) (err error) {
	c.do(func(q string) {
		var rows isql.Rows
		rows, err = c.sql.RandSlave().QueryContext(c.tlbx.Ctx(), q, args...)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (c *sqlClient) QueryRow(query string, args ...interface{}) (row isql.Row) {
	c.do(func(q string) { row = c.sql.RandSlave().QueryRowContext(c.tlbx.Ctx(), q, args...) }, query)
	return
}

func (c *sqlClient) do(do func(string), query string) {
	// no query should ever even come close to 1 second in execution time
	start := NowUnixMilli()
	do(`SET STATEMENT max_statement_time=1 FOR ` + query)
	c.tlbx.LogQueryStats(&app.QueryStats{
		Milli: NowUnixMilli() - start,
		Query: query,
	})
}

type RedisPoolClient interface {
	Base() iredis.Pool
	Get() RedisConnClient
}

type redisPoolWrapper struct {
	tlbx app.Toolbox
	pool iredis.Pool
}

func (w *redisPoolWrapper) Base() iredis.Pool {
	return w.pool
}

func (w *redisPoolWrapper) Get() RedisConnClient {
	return &redisConnWrapper{
		tlbx: w.tlbx,
		conn: w.pool.Get(),
	}
}

type RedisConnClient interface {
	Base() iredis.Conn
	iredis.Conn
}

type redisConnWrapper struct {
	tlbx app.Toolbox
	conn iredis.Conn
}

func (w *redisConnWrapper) Base() iredis.Conn {
	return w.conn
}

func (w *redisConnWrapper) Close() error {
	return w.conn.Close()
}

func (w *redisConnWrapper) Err() error {
	return w.conn.Err()
}

func (w *redisConnWrapper) Do(cmd string, args ...interface{}) (reply interface{}, err error) {
	w.do(func(q string, a ...interface{}) { reply, err = w.conn.Do(cmd, args...) }, cmd, args...)
	return
}

func (w *redisConnWrapper) Send(cmd string, args ...interface{}) (err error) {
	w.do(func(q string, a ...interface{}) { err = w.conn.Send(cmd, args...) }, cmd, args...)
	return
}

func (w *redisConnWrapper) Flush() error {
	return w.conn.Flush()
}

func (w *redisConnWrapper) Receive() (reply interface{}, err error) {
	return w.conn.Receive()
}

func (w *redisConnWrapper) do(do func(string, ...interface{}), cmd string, args ...interface{}) {
	start := NowUnixMilli()
	do(cmd, args...)
	w.tlbx.LogQueryStats(&app.QueryStats{
		Milli: NowUnixMilli() - start,
		Query: Sprint(append([]interface{}{cmd, " "}, args...)...),
	})
}
