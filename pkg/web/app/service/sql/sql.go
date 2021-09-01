package sql

import (
	"database/sql"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/sqlh"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/jmoiron/sqlx"
)

type tlbxKey struct {
	name string
}

func Mware(name string, sql sqlh.ReplicaSet) func(app.Tlbx) {
	return func(tlbx app.Tlbx) {
		tlbx.Set(tlbxKey{name}, &client{tlbx: tlbx, name: name, sql: sql})
	}
}

func Get(tlbx app.Tlbx, name string) Client {
	return tlbx.Get(tlbxKey{name}).(Client)
}

type Client interface {
	Base() sqlh.ReplicaSet
	BeginRead() Tx
	BeginWrite() Tx
	ClientCore
}

type ClientCore interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	MustExec(query string, args ...interface{}) sql.Result
	NamedExec(query string, arg interface{}) (sql.Result, error)
	MustNamedExec(query string, arg interface{}) sql.Result
	Query(rowsFn func(*sqlx.Rows), query string, args ...interface{}) error
	MustQuery(rowsFn func(*sqlx.Rows), query string, args ...interface{})
	NamedQuery(rowsFn func(*sqlx.Rows), query string, arg interface{}) error
	MustNamedQuery(rowsFn func(*sqlx.Rows), query string, arg interface{})
	Query_b(rowsFn func(*sql.Rows), query string, args ...interface{}) error
	MustQuery_b(rowsFn func(*sql.Rows), query string, args ...interface{})
	QueryRow(query string, args ...interface{}) *sqlx.Row
	QueryRow_b(query string, args ...interface{}) *sql.Row
	Get1(dst interface{}, query string, args ...interface{}) error
	MustGet1(dst interface{}, query string, args ...interface{})
	GetN(dst interface{}, query string, args ...interface{}) error
	MustGetN(dst interface{}, query string, args ...interface{})
}

type Tx interface {
	ClientCore
	Rollback()
	Commit()
}

type tx struct {
	tx        *sqlx.Tx
	tlbx      app.Tlbx
	sqlClient *client
	readOnly  bool
	done      bool
}

func (t *tx) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	PanicIf(t.readOnly, "can't perform exec query on a read only transaction")
	t.sqlClient.do(func(q string) { res, err = t.tx.ExecContext(t.tlbx.Ctx(), q, args...) }, query)
	return
}

func (t *tx) MustExec(query string, args ...interface{}) sql.Result {
	res, err := t.Exec(query, args...)
	PanicOn(err)
	return res
}

func (t *tx) NamedExec(query string, arg interface{}) (res sql.Result, err error) {
	PanicIf(t.readOnly, "can't perform exec query on a read only transaction")
	t.sqlClient.do(func(q string) { res, err = t.tx.NamedExecContext(t.tlbx.Ctx(), q, arg) }, query)
	return
}

func (t *tx) MustNamedExec(query string, arg interface{}) sql.Result {
	res, err := t.NamedExec(query, arg)
	PanicOn(err)
	return res
}

func (t *tx) Query(rowsFn func(*sqlx.Rows), query string, args ...interface{}) (err error) {
	t.sqlClient.do(func(q string) {
		var rows *sqlx.Rows
		rows, err = t.tx.QueryxContext(t.tlbx.Ctx(), q, args...)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (t *tx) MustQuery(rowsFn func(*sqlx.Rows), query string, args ...interface{}) {
	PanicOn(t.Query(rowsFn, query, args...))
}

func (t *tx) NamedQuery(rowsFn func(*sqlx.Rows), query string, arg interface{}) (err error) {
	t.sqlClient.do(func(q string) {
		var rows *sqlx.Rows
		rows, err = t.tx.NamedQuery(q, arg)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (t *tx) MustNamedQuery(rowsFn func(*sqlx.Rows), query string, arg interface{}) {
	PanicOn(t.NamedQuery(rowsFn, query, arg))
}

func (t *tx) Query_b(rowsFn func(*sql.Rows), query string, args ...interface{}) (err error) {
	t.sqlClient.do(func(q string) {
		var rows *sql.Rows
		rows, err = t.tx.QueryContext(t.tlbx.Ctx(), q, args...)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (t *tx) MustQuery_b(rowsFn func(*sql.Rows), query string, args ...interface{}) {
	PanicOn(t.Query_b(rowsFn, query, args...))
}

func (t *tx) QueryRow(query string, args ...interface{}) (row *sqlx.Row) {
	t.sqlClient.do(func(q string) { row = t.tx.QueryRowxContext(t.tlbx.Ctx(), q, args...) }, query)
	return
}

func (t *tx) QueryRow_b(query string, args ...interface{}) (row *sql.Row) {
	t.sqlClient.do(func(q string) { row = t.tx.QueryRowContext(t.tlbx.Ctx(), q, args...) }, query)
	return
}

func (t *tx) Get1(dst interface{}, query string, args ...interface{}) (err error) {
	t.sqlClient.do(func(q string) { err = t.tx.GetContext(t.tlbx.Ctx(), dst, q, args...) }, query)
	return
}

func (t *tx) MustGet1(dst interface{}, query string, args ...interface{}) {
	PanicOn(t.Get1(dst, query, args...))
}

func (t *tx) GetN(dst interface{}, query string, args ...interface{}) (err error) {
	t.sqlClient.do(func(q string) { err = t.tx.SelectContext(t.tlbx.Ctx(), dst, q, args...) }, query)
	return
}

func (t *tx) MustGetN(dst interface{}, query string, args ...interface{}) {
	PanicOn(t.GetN(dst, query, args...))
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
	sql  sqlh.ReplicaSet
}

func (c *client) Base() sqlh.ReplicaSet {
	return c.sql
}

func (c *client) BeginRead() Tx {
	return c.begin(true)
}

func (c *client) BeginWrite() Tx {
	return c.begin(false)
}

func (c *client) begin(readOnly bool) Tx {
	var t *sqlx.Tx
	var err error
	var msg string
	var db *sqlx.DB
	if readOnly {
		msg = "BEGIN (READONLY)"
		db = c.sql.RandSlave()
	} else {
		msg = "BEGIN (WRITE)"
		db = c.sql.Primary()
	}
	c.do(func(s string) {
		t, err = db.BeginTxx(c.tlbx.Ctx(), &sql.TxOptions{
			ReadOnly: readOnly,
		})
	}, msg)
	PanicOn(err)
	return &tx{
		tx:        t,
		tlbx:      c.tlbx,
		readOnly:  readOnly,
		sqlClient: c,
	}
}

func (c *client) Exec(query string, args ...interface{}) (res sql.Result, err error) {
	c.do(func(q string) { res, err = c.sql.Primary().ExecContext(c.tlbx.Ctx(), q, args...) }, query)
	return
}

func (c *client) MustExec(query string, args ...interface{}) sql.Result {
	res, err := c.Exec(query, args...)
	PanicOn(err)
	return res
}

func (c *client) NamedExec(query string, arg interface{}) (res sql.Result, err error) {
	c.do(func(q string) { res, err = c.sql.Primary().NamedExecContext(c.tlbx.Ctx(), q, arg) }, query)
	return
}

func (c *client) MustNamedExec(query string, arg interface{}) sql.Result {
	res, err := c.NamedExec(query, arg)
	PanicOn(err)
	return res
}

func (c *client) Query(rowsFn func(*sqlx.Rows), query string, args ...interface{}) (err error) {
	c.do(func(q string) {
		var rows *sqlx.Rows
		rows, err = c.sql.RandSlave().QueryxContext(c.tlbx.Ctx(), q, args...)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (c *client) MustQuery(rowsFn func(*sqlx.Rows), query string, args ...interface{}) {
	PanicOn(c.Query(rowsFn, query, args...))
}

func (c *client) NamedQuery(rowsFn func(*sqlx.Rows), query string, arg interface{}) (err error) {
	c.do(func(q string) {
		var rows *sqlx.Rows
		rows, err = c.sql.RandSlave().NamedQuery(q, arg)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (c *client) MustNamedQuery(rowsFn func(*sqlx.Rows), query string, arg interface{}) {
	PanicOn(c.NamedQuery(rowsFn, query, arg))
}

func (c *client) Query_b(rowsFn func(*sql.Rows), query string, args ...interface{}) (err error) {
	c.do(func(q string) {
		var rows *sql.Rows
		rows, err = c.sql.RandSlave().QueryContext(c.tlbx.Ctx(), q, args...)
		if rows != nil {
			defer rows.Close()
			rowsFn(rows)
		}
	}, query)
	return
}

func (c *client) MustQuery_b(rowsFn func(*sql.Rows), query string, args ...interface{}) {
	PanicOn(c.Query_b(rowsFn, query, args...))
}

func (c *client) QueryRow(query string, args ...interface{}) (row *sqlx.Row) {
	c.do(func(q string) { row = c.sql.RandSlave().QueryRowxContext(c.tlbx.Ctx(), q, args...) }, query)
	return
}

func (c *client) QueryRow_b(query string, args ...interface{}) (row *sql.Row) {
	c.do(func(q string) { row = c.sql.RandSlave().QueryRowContext(c.tlbx.Ctx(), q, args...) }, query)
	return
}

func (c *client) Get1(dst interface{}, query string, args ...interface{}) (err error) {
	c.do(func(q string) { err = c.sql.RandSlave().GetContext(c.tlbx.Ctx(), dst, q, args...) }, query)
	return
}

func (c *client) MustGet1(dst interface{}, query string, args ...interface{}) {
	PanicOn(c.Get1(dst, query, args...))
}

func (c *client) GetN(dst interface{}, query string, args ...interface{}) (err error) {
	c.do(func(q string) { err = c.sql.RandSlave().SelectContext(c.tlbx.Ctx(), dst, q, args...) }, query)
	return
}

func (c *client) MustGetN(dst interface{}, query string, args ...interface{}) {
	PanicOn(c.GetN(dst, query, args...))
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
