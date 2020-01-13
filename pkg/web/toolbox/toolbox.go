package toolbox

import (
	"context"
	"net/http"
	"sync"
)

func Mware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(
			context.WithValue(
				r.Context(),
				ctxKey{},
				&toolBox{
					mtx:   &sync.RWMutex{},
					store: map[interface{}]interface{}{}}))
		next(w, r)
	}
}

func Get(r *http.Request) ToolBox {
	return r.Context().Value(ctxKey{}).(ToolBox)
}

type ToolBox interface {
	Get(key interface{}) interface{}
	Set(key, value interface{})
}

type toolBox struct {
	mtx   *sync.RWMutex
	store map[interface{}]interface{}
}

func (t *toolBox) Get(key interface{}) interface{} {
	t.mtx.RLock()
	defer t.mtx.RUnlock()
	return t.store[key]
}

func (t *toolBox) Set(key, value interface{}) {
	t.mtx.Lock()
	defer t.mtx.Unlock()
	t.store[key] = value
}

type ctxKey struct{}
