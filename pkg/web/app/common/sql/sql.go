package sql

import (
	"database/sql"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
	"net/http"
)

func ReturnNotFoundOrPanic(err error) {
	if err != nil && err == sql.ErrNoRows {
		PanicOn(&app.ErrMsg{
			Status: http.StatusNotFound,
			Msg:    http.StatusText(http.StatusNotFound),
		})
	}
	PanicOn(err)
}

func Asc(asc bool) string {
	if asc {
		return ` ASC`
	}
	return ` DESC`
}

func Limit(l, max int) int {
	switch {
	case l < 1:
		return 2 // 1 + 1 for "more": true/false detection
	case l > max:
		return max + 1
	default:
		return 1 + 1
	}
}

func LimitMax100(l int) int {
	return Limit(l, 100)
}

func OrderLimit(field string, asc bool, l, max int) string {
	return Sprintf(` ORDER BY %s %s LIMIT %d`, field, Asc(asc), Limit(l, max))
}

func OrderLimitMax100(field string, asc bool, l int) string {
	return Sprintf(` ORDER BY %s %s LIMIT %d`, field, Asc(asc), LimitMax100(l))
}
