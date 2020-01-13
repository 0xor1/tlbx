package options

import (
	"net/http"
)

func Mware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// options request for cors pre flight
		if r.Method == http.MethodOptions {
			w.Header().Add("Access-Control-Allow-Methods", "GET, PUT")
			w.WriteHeader(200)
			return
		}
		next(w, r)
	}
}
