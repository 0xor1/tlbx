package static

import (
	"net/http"
	"path/filepath"

	. "github.com/0xor1/wtf/pkg/core"
)

func Mware(dir string, isStaticReq func(*http.Request) bool, next http.HandlerFunc) http.HandlerFunc {
	staticFileDir, err := filepath.Abs(dir)
	PanicOn(err)
	fileServer := http.FileServer(http.Dir(staticFileDir))
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && isStaticReq(r) {
			fileServer.ServeHTTP(w, r)
		} else {
			next(w, r)
		}
	}
}
