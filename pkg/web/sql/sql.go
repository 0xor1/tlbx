package sql

import (
	"net/http"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/web/toolbox"
	_ "github.com/go-sql-driver/mysql"
)

func Mware(usrDSN, pwdDSN, dataDSN string, next http.HandlerFunc) http.HandlerFunc {
	usr, err := isql.NewOpener().Open("mysql", usrDSN)
	PanicOn(err)
	pwd, err := isql.NewOpener().Open("mysql", pwdDSN)
	PanicOn(err)
	data, err := isql.NewOpener().Open("mysql", dataDSN)
	PanicOn(err)
	return func(w http.ResponseWriter, r *http.Request) {
		toolbox.Get(r).Set(tlbxKey{}, &dbSet{
			usr:  usr,
			pwd:  pwd,
			data: data,
		})
		next(w, r)
	}
}

type dbSet struct {
	usr  isql.DB
	pwd  isql.DB
	data isql.DB
}

func User(r *http.Request) isql.DBCore {
	return get(r).usr
}

func Pwd(r *http.Request) isql.DBCore {
	return get(r).pwd
}

func Data(r *http.Request) isql.DBCore {
	return get(r).data
}

func get(r *http.Request) *dbSet {
	return toolbox.Get(r).Get(tlbxKey{}).(*dbSet)
}

type tlbxKey struct{}
