package sql

import (
	"database/sql"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type tlbxKey struct {
	name string
}

func Mware(name string, sql isql.ReplicaSet) func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Set(tlbxKey{name}, &client{tlbx: tlbx, name: name, sql: sql})
	}
}

func Get(tlbx app.Tlbx, name string) Client {
	return tlbx.Get(tlbxKey{name}).(Client)
}

type Client interface {
	Base() isql.ReplicaSet
	BeginRead() Tx
	BeginWrite() Tx
	ClientCore
}

type ClientCore interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(rowsFn func(isql.Rows), query string, args ...interface{}) error
	QueryRow(query string, args ...interface{}) isql.Row
}

type Tx interface {
	ClientCore
	Rollback()
	Commit()
}

type tx struct {
	tx        isql.Tx
	tlbx      app.Tlbx
	sqlClient *client
	readOnly  bool
	done      bool
}

func (t *tx) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	PanicIf(t.readOnly, "can't perform exec write query on a read only transaction")
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

type client struct {
	tlbx app.Tlbx
	name string
	sql  isql.ReplicaSet
}

func (c *client) Base() isql.ReplicaSet {
	return c.sql
}

func (c *client) BeginRead() Tx {
	var t isql.Tx
	var err error
	c.do(func(s string) {
		t, err = c.sql.Primary().BeginTx(c.tlbx.Ctx(), &sql.TxOptions{
			ReadOnly: true,
		})
	}, "START TRANSACTION (READONLY)")
	PanicOn(err)
	return &tx{
		tx:        t,
		tlbx:      c.tlbx,
		readOnly:  true,
		sqlClient: c,
	}
}

func (c *client) BeginWrite() Tx {
	var t isql.Tx
	var err error
	c.do(func(s string) {
		t, err = c.sql.Primary().BeginTx(c.tlbx.Ctx(), &sql.TxOptions{
			ReadOnly: false,
		})
	}, "START TRANSACTION (WRITE)")
	PanicOn(err)
	return &tx{
		tx:        t,
		tlbx:      c.tlbx,
		readOnly:  false,
		sqlClient: c,
	}
}

func (c *client) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	c.do(func(q string) { res, err = c.sql.Primary().ExecContext(c.tlbx.Ctx(), q, args...) }, query)
	return
}

func (c *client) Query(rowsFn func(isql.Rows), query string, args ...interface{}) (err error) {
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

func (c *client) QueryRow(query string, args ...interface{}) (row isql.Row) {
	c.do(func(q string) { row = c.sql.RandSlave().QueryRowContext(c.tlbx.Ctx(), q, args...) }, query)
	return
}

func (c *client) do(do func(string), query string) {
	// no query should ever even come close to 1 second in execution time
	start := NowUnixMilli()
	do(`SET STATEMENT max_statement_time=1 FOR ` + query)
	c.tlbx.LogActionStats(&app.ActionStats{
		Milli:  NowUnixMilli() - start,
		Type:   "SQL",
		Name:   c.name,
		Action: query,
	})
}
