package isql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"math/rand"
	"reflect"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	_ "github.com/go-sql-driver/mysql"
)

func NewOpener() Opener {
	return &opener{}
}

type Opener interface {
	Open(dataSourceName string) (DB, error)
}

func NewDB(db *sql.DB) DB {
	if db == nil {
		return nil
	}
	return &dbWrapper{
		db: db,
	}
}

type DB interface {
	DBCore
	Begin() (Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error)
	Close() error
	Driver() driver.Driver
	Exec(query string, args ...interface{}) (sql.Result, error)
	Ping() error
	PingContext(ctx context.Context) error
	Prepare(query string) (Stmt, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	Query(query string, args ...interface{}) (Rows, error)
	QueryRow(query string, args ...interface{}) Row
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}

func NewReplicaSet(primaryDataSourceName string, slaveDataSourceNames ...string) (ReplicaSet, error) {
	op := &opener{}
	primary, err := op.Open(primaryDataSourceName)
	if err != nil {
		return nil, err
	}
	rs := &replicaSet{
		primary: primary,
		slaves:  make([]DB, 0, len(slaveDataSourceNames)),
	}
	for _, slaveDataSourceName := range slaveDataSourceNames {
		slave, err := op.Open(slaveDataSourceName)
		if err != nil {
			return nil, err
		}
		rs.slaves = append(rs.slaves, slave)
	}
	return rs, nil
}

func MustNewReplicaSet(driverName, primaryDataSourceName string, slaveDataSourceNames ...string) ReplicaSet {
	rs, err := NewReplicaSet(primaryDataSourceName, slaveDataSourceNames...)
	PanicOn(err)
	return rs
}

type DBCore interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
}

type ReplicaSet interface {
	Primary() DB
	RandSlave() DB
}

func NewRow(row *sql.Row) Row {
	if row == nil {
		return nil
	}
	return &rowWrapper{
		row: row,
	}
}

type Row interface {
	Scan(dest ...interface{}) error
}

func NewRows(rows *sql.Rows) Rows {
	if rows == nil {
		return nil
	}
	return &rowsWrapper{
		rows: rows,
	}
}

type Rows interface {
	Close() error
	ColumnTypes() ([]ColumnType, error)
	Columns() ([]string, error)
	Err() error
	Next() bool
	NextResultSet() bool
	Scan(dest ...interface{}) error
}

func NewStmt(stmt *sql.Stmt) Stmt {
	if stmt == nil {
		return nil
	}
	return &stmtWrapper{
		stmt: stmt,
	}
}

type Stmt interface {
	Close() error
	Exec(args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error)
	Query(args ...interface{}) (Rows, error)
	QueryContext(ctx context.Context, args ...interface{}) (Rows, error)
	QueryRow(args ...interface{}) Row
	QueryRowContext(ctx context.Context, args ...interface{}) Row
}

func NewTx(tx *sql.Tx) Tx {
	if tx == nil {
		return nil
	}
	return &txWrapper{
		tx: tx,
	}
}

type Tx interface {
	DBCore
	Commit() error
	Exec(query string, args ...interface{}) (sql.Result, error)
	Prepare(query string) (Stmt, error)
	PrepareContext(ctx context.Context, query string) (Stmt, error)
	Query(query string, args ...interface{}) (Rows, error)
	QueryRow(query string, args ...interface{}) Row
	Rollback() error
	Stmt(stmt *sql.Stmt) Stmt
	StmtContext(ctx context.Context, stmt *sql.Stmt) Stmt
}

func NewColumnType(columnType *sql.ColumnType) ColumnType {
	if columnType == nil {
		return nil
	}
	return &columnTypeWrapper{
		columnType: columnType,
	}
}

type ColumnType interface {
	DatabaseTypeName() string
	DecimalSize() (precision, scale int64, ok bool)
	Length() (length int64, ok bool)
	Name() string
	Nullable() (nullable, ok bool)
	ScanType() reflect.Type
}

type opener struct {
}

func (o *opener) Open(dataSourceName string) (DB, error) {
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		if db != nil {
			db.Close()
		}
		return nil, err
	}
	return NewDB(db), nil
}

type dbWrapper struct {
	db *sql.DB
}

func (d *dbWrapper) Begin() (Tx, error) {
	tx, err := d.db.Begin()
	return NewTx(tx), err
}

func (d *dbWrapper) BeginTx(ctx context.Context, opts *sql.TxOptions) (Tx, error) {
	tx, err := d.db.BeginTx(ctx, opts)
	return NewTx(tx), err
}

func (d *dbWrapper) Close() error {
	return d.db.Close()
}

func (d *dbWrapper) Driver() driver.Driver {
	return d.db.Driver()
}

func (d *dbWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(query, args...)
}

func (d *dbWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}

func (d *dbWrapper) Ping() error {
	return d.db.Ping()
}

func (d *dbWrapper) PingContext(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *dbWrapper) Prepare(query string) (Stmt, error) {
	stmt, err := d.db.Prepare(query)
	return NewStmt(stmt), err
}

func (d *dbWrapper) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	stmt, err := d.db.PrepareContext(ctx, query)
	return NewStmt(stmt), err
}

func (d *dbWrapper) Query(query string, args ...interface{}) (Rows, error) {
	rows, err := d.db.Query(query, args...)
	return NewRows(rows), err
}

func (d *dbWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := d.db.QueryContext(ctx, query, args...)
	return NewRows(rows), err
}

func (d *dbWrapper) QueryRow(query string, args ...interface{}) Row {
	return NewRow(d.db.QueryRow(query, args...))
}

func (d *dbWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return NewRow(d.db.QueryRowContext(ctx, query, args...))
}

func (d *dbWrapper) SetConnMaxLifetime(dur time.Duration) {
	d.db.SetConnMaxLifetime(dur)
}

func (d *dbWrapper) SetMaxIdleConns(n int) {
	d.db.SetMaxIdleConns(n)
}

func (d *dbWrapper) SetMaxOpenConns(n int) {
	d.db.SetMaxOpenConns(n)
}

func (d *dbWrapper) Stats() sql.DBStats {
	return d.db.Stats()
}

type replicaSet struct {
	primary DB
	slaves  []DB
}

func (r *replicaSet) Primary() DB {
	return r.primary
}

func (r *replicaSet) RandSlave() DB {
	if len(r.slaves) > 0 {
		return r.slaves[rand.Intn(len(r.slaves))]
	} else {
		return r.primary
	}
}

type rowWrapper struct {
	row *sql.Row
}

func (r *rowWrapper) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

type rowsWrapper struct {
	rows *sql.Rows
}

func (r *rowsWrapper) Close() error {
	return r.rows.Close()
}

func (r *rowsWrapper) ColumnTypes() ([]ColumnType, error) {
	columnTypes, err := r.rows.ColumnTypes()
	res := make([]ColumnType, 0, len(columnTypes))
	for _, ct := range columnTypes {
		res = append(res, NewColumnType(ct))
	}
	return res, err
}

func (r *rowsWrapper) Columns() ([]string, error) {
	return r.rows.Columns()
}

func (r *rowsWrapper) Err() error {
	return r.rows.Err()
}

func (r *rowsWrapper) Next() bool {
	return r.rows.Next()
}

func (r *rowsWrapper) NextResultSet() bool {
	return r.rows.NextResultSet()
}

func (r *rowsWrapper) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

type stmtWrapper struct {
	stmt *sql.Stmt
}

func (s *stmtWrapper) Close() error {
	return s.stmt.Close()
}

func (s *stmtWrapper) Exec(args ...interface{}) (sql.Result, error) {
	return s.stmt.Exec(args...)
}

func (s *stmtWrapper) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	return s.stmt.ExecContext(ctx, args...)
}

func (s *stmtWrapper) Query(args ...interface{}) (Rows, error) {
	rows, err := s.stmt.Query(args...)
	return NewRows(rows), err
}

func (s *stmtWrapper) QueryContext(ctx context.Context, args ...interface{}) (Rows, error) {
	rows, err := s.stmt.QueryContext(ctx, args...)
	return NewRows(rows), err
}

func (s *stmtWrapper) QueryRow(args ...interface{}) Row {
	return NewRow(s.stmt.QueryRow(args...))
}

func (s *stmtWrapper) QueryRowContext(ctx context.Context, args ...interface{}) Row {
	return NewRow(s.stmt.QueryRowContext(ctx, args...))
}

type txWrapper struct {
	tx *sql.Tx
}

func (t *txWrapper) Commit() error {
	return t.tx.Commit()
}

func (t *txWrapper) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.tx.Exec(query, args...)
}

func (t *txWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *txWrapper) Prepare(query string) (Stmt, error) {
	stmt, err := t.tx.Prepare(query)
	return NewStmt(stmt), err
}

func (t *txWrapper) PrepareContext(ctx context.Context, query string) (Stmt, error) {
	stmt, err := t.tx.PrepareContext(ctx, query)
	return NewStmt(stmt), err
}

func (t *txWrapper) Query(query string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.Query(query, args...)
	return NewRows(rows), err
}

func (t *txWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.QueryContext(ctx, query, args...)
	return NewRows(rows), err
}

func (t *txWrapper) QueryRow(query string, args ...interface{}) Row {
	return NewRow(t.tx.QueryRow(query, args...))
}

func (t *txWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return NewRow(t.tx.QueryRowContext(ctx, query, args...))
}

func (t *txWrapper) Rollback() error {
	return t.tx.Rollback()
}

func (t *txWrapper) Stmt(stmt *sql.Stmt) Stmt {
	return NewStmt(t.tx.Stmt(stmt))
}

func (t *txWrapper) StmtContext(ctx context.Context, stmt *sql.Stmt) Stmt {
	return NewStmt(t.tx.StmtContext(ctx, stmt))
}

type columnTypeWrapper struct {
	columnType *sql.ColumnType
}

func (c *columnTypeWrapper) DatabaseTypeName() string {
	return c.columnType.DatabaseTypeName()
}

func (c *columnTypeWrapper) DecimalSize() (precision, scale int64, ok bool) {
	return c.columnType.DecimalSize()
}

func (c *columnTypeWrapper) Length() (length int64, ok bool) {
	return c.columnType.Length()
}

func (c *columnTypeWrapper) Name() string {
	return c.columnType.Name()
}

func (c *columnTypeWrapper) Nullable() (nullable, ok bool) {
	return c.columnType.Nullable()
}

func (c *columnTypeWrapper) ScanType() reflect.Type {
	return c.columnType.ScanType()
}
