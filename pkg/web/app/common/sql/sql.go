package sql

import (
	"database/sql"
	"net/http"
	"strings"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
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

func GtLtSymbol(asc bool) string {
	if asc {
		return ">"
	} else {
		return "<"
	}
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

func InCondition(and bool, field string, setLen int) string {
	if setLen <= 0 {
		return ""
	}
	op := "AND"
	if !and {
		op = "OR"
	}
	return Sprintf(` %s %s IN (?%s)`, op, field, strings.Repeat(`,?`, setLen-1))
}
