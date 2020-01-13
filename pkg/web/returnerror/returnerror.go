package returnerror

import (
	"net/http"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/json"
	"github.com/0xor1/wtf/pkg/web/log"
)

func Mware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := ToError(recover()); e != nil {
				log.Get(r).ErrorOn(e)
				if err, ok := e.Value().(*err); ok {
					json.WriteHttp(w, err.status, err.body)
				} else {
					json.WriteHttp(w, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				}
			}
		}()
		next(w, r)
	}
}

func If(condition bool, status int, fmt string, args ...interface{}) {
	if condition {
		PanicOn(&err{
			status: status,
			body:   Sprintf(fmt, args...),
		})
	}
}

type err struct {
	status int
	body   string
}

func (e *err) Error() string {
	return Sprintf("returning error status: %d, body: %s", e.status, e.body)
}
