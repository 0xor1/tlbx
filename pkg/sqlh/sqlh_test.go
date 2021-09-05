package sqlh_test

import (
	"testing"
	"time"

	"github.com/0xor1/tlbx/pkg/sqlh"
	"github.com/stretchr/testify/assert"
)

func TestReplicaSet(t *testing.T) {
	a := assert.New(t)
	// using pwds schema for purposes of testing
	cnnStr := "pwds:C0-Mm-0n-Pwd5@tcp(localhost:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true"
	rs := sqlh.MustNewReplicaSet(cnnStr, cnnStr, cnnStr)
	a.Len(rs.Slaves(), 2)
	a.NotNil(rs.RandSlave())
	rs.SetConnMaxLifetime(5 * time.Second)
	rs.SetMaxIdleConns(5)
	rs.SetMaxOpenConns(5)
}

func TestNamed(t *testing.T) {
	a := assert.New(t)
	qry, args := sqlh.Named(":yolo, :nolo, :solo", map[string]interface{}{
		"yolo": 1,
		"nolo": 1,
		"solo": 1,
	})
	a.Equal("?, ?, ?", qry)
	a.Len(args, 3)
}

func TestIn(t *testing.T) {
	a := assert.New(t)
	qry, args := sqlh.Named("SELECT * FROM foo WHERE y=:yolo, n IN (:nolo), s IN (:solo)",
		map[string]interface{}{
			"yolo": 1,
			"nolo": []int{1, 2},
			"solo": []int{1, 2, 3},
		})
	a.Equal("SELECT * FROM foo WHERE y=?, n IN (?), s IN (?)", qry)
	a.Equal(args, []interface{}{1, []int{1, 2}, []int{1, 2, 3}})
	qry, args = sqlh.In(qry, args...)
	a.Equal("SELECT * FROM foo WHERE y=?, n IN (?, ?), s IN (?, ?, ?)", qry)
	a.Equal(args, []interface{}{1, 1, 2, 1, 2, 3})
}
