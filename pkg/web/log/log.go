package log

import (
	"net/http"

	"github.com/0xor1/wtf/pkg/log"
	"github.com/0xor1/wtf/pkg/web/toolbox"
)

func Mware(log log.Log, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		toolbox.Get(r).Set(tlbxKey{}, log)
		next(w, r)
	}
}

func Get(r *http.Request) log.Log {
	return toolbox.Get(r).Get(tlbxKey{}).(log.Log)
}

type tlbxKey struct{}
