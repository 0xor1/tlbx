package redis

import (
	"net/http"

	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/web/toolbox"
)

func Mware(addr string, next http.HandlerFunc) http.HandlerFunc {
	pool := iredis.CreatePool(addr)
	return func(w http.ResponseWriter, r *http.Request) {
		toolbox.Get(r).Set(tlbxKey{}, pool)
		next(w, r)
	}
}

func Get(r *http.Request) iredis.Pool {
	return toolbox.Get(r).Get(tlbxKey{}).(iredis.Pool)
}

type tlbxKey struct{}
