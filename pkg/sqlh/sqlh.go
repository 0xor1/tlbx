package sqlh

import (
	"database/sql"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	_ "github.com/go-sql-driver/mysql"
)

var (
	ErrNoRows   = sql.ErrNoRows
	ErrConnDone = sql.ErrConnDone
	ErrTxDone   = sql.ErrTxDone
)

func NewArgs(size int) *Args {
	return &Args{
		args: make([]interface{}, 0, size),
	}
}

type Args struct {
	args []interface{}
}

func (a *Args) Append(arg ...interface{}) {
	a.args = append(a.args, arg...)
}

func (a *Args) Is() []interface{} {
	return a.args
}

type Row interface {
	Scan(dst ...interface{}) error
}

type ReplicaSet interface {
	Primary() *sqlx.DB
	Slaves() []*sqlx.DB
	RandSlave() *sqlx.DB
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
}

func NewReplicaSet(primaryDataSourceName string, slaveDataSourceNames ...string) (ReplicaSet, error) {
	primary, err := open(primaryDataSourceName)
	if err != nil {
		return nil, err
	}
	rs := &replicaSet{
		primary: primary,
		slaves:  make([]*sqlx.DB, 0, len(slaveDataSourceNames)),
	}
	for _, slaveDataSourceName := range slaveDataSourceNames {
		slave, err := open(slaveDataSourceName)
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

func open(dataSourceName string) (*sqlx.DB, error) {

	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		if db != nil {
			db.Close()
		}
		return nil, err
	}
	return db, nil
}

type replicaSet struct {
	primary *sqlx.DB
	slaves  []*sqlx.DB
}

func (r *replicaSet) Primary() *sqlx.DB {
	return r.primary
}

func (r *replicaSet) Slaves() []*sqlx.DB {
	return r.slaves
}

func (r *replicaSet) RandSlave() *sqlx.DB {
	if len(r.slaves) > 0 {
		return r.slaves[rand.Intn(len(r.slaves))]
	} else {
		return r.primary
	}
}
func (r *replicaSet) SetConnMaxLifetime(d time.Duration) {
	r.primary.SetConnMaxLifetime(d)
	for _, slave := range r.slaves {
		slave.SetConnMaxLifetime(d)
	}
}

func (r *replicaSet) SetMaxIdleConns(n int) {
	r.primary.SetMaxIdleConns(n)
	for _, slave := range r.slaves {
		slave.SetMaxIdleConns(n)
	}
}

func (r *replicaSet) SetMaxOpenConns(n int) {
	r.primary.SetMaxOpenConns(n)
	for _, slave := range r.slaves {
		slave.SetMaxOpenConns(n)
	}
}

func ReturnNotFoundIfIsNoRows(err error) {
	app.ReturnIf(IsNoRows(err), http.StatusNotFound, "")
	PanicOn(err)
}

func PanicIfIsntNoRows(err error) {
	if !IsNoRows(err) {
		PanicOn(err)
	}
}

func IsNoRows(err error) bool {
	return err != nil && err == sql.ErrNoRows
}

func Named(query string, arg interface{}) (string, []interface{}) {
	qry, args, err := sqlx.Named(query, arg)
	PanicOn(err)
	return qry, args
}

func In(query string, args ...interface{}) (string, []interface{}) {
	qry, args, err := sqlx.In(query, args...)
	PanicOn(err)
	return qry, args
}

func Asc(asc bool) string {
	if asc {
		return ` ASC`
	}
	return ` DESC`
}

func GtLtSymbol(asc bool) string {
	if asc {
		return ">"
	} else {
		return "<"
	}
}

func Limit(l, max uint16) uint16 {
	switch {
	case l >= max:
		return max + 1
	case l < 1:
		return 2 // 1 + 1 for "more": true/false detection
	default:
		return l + 1
	}
}

func Limit100(l uint16) uint16 {
	return Limit(l, 100)
}

func OrderLimit(field string, asc bool, l, max uint16) string {
	return Strf(` ORDER BY %s %s LIMIT %d`, field, Asc(asc), Limit(l, max))
}

func OrderLimit100(field string, asc bool, l uint16) string {
	return OrderLimit(field, asc, l, 100)
}

func InCondition(and bool, field string, setLen int) string {
	PanicIf(setLen <= 0, "setLen must be > 0")
	op := `AND`
	if !and {
		op = `OR`
	}
	return Strf(` %s %s IN (%s)`, op, field, PList(setLen))
}

func OrderByField(field string, setLen int) string {
	if setLen <= 0 {
		return ``
	}
	return Strf(` ORDER BY FIELD (%s,%s)`, field, PList(setLen))
}

func PList(count int) string {
	PanicIf(count < 1, `count must be >= 1`)
	return `?` + strings.Repeat(`,?`, count-1)
}
