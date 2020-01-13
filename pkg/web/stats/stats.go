package stats

import (
	"net/http"
	"sync"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/log"
	"github.com/0xor1/wtf/pkg/web/toolbox"
)

func Mware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWrapper{w: w}
		l := &logger{
			mtx:   &sync.Mutex{},
			stats: make([]*Query, 0, 10),
		}
		toolbox.Get(r).Set(tlbxKey{}, l)
		start := NowUnixMilli()

		defer func() {
			l.mtx.Lock()
			defer l.mtx.Unlock()
			log.Get(r).Stats(&Req{
				Milli:   NowUnixMilli() - start,
				Status:  rw.status,
				Method:  r.Method,
				Path:    r.URL.Path,
				Queries: l.stats,
			})
		}()

		next(rw, r)

	}
}

func Get(r *http.Request) Logger {
	return toolbox.Get(r).Get(tlbxKey{}).(Logger)
}

type Logger interface {
	Add(*Query)
}

type logger struct {
	mtx   *sync.Mutex
	stats []*Query
}

func (l *logger) Add(qs *Query) {
	l.mtx.Lock()
	defer l.mtx.Unlock()
	l.stats = append(l.stats, qs)
}

type Query struct {
	Milli int64       `json:"ms"`
	Query string      `json:"query"`
	Args  interface{} `json:"args"`
}

type Req struct {
	Milli   int64    `json:"ms"`
	Status  int      `json:"status"`
	Method  string   `json:"method"`
	Path    string   `json:"path"`
	Queries []*Query `json:"queries"`
}

func (r *Req) String() string {
	return Sprintf("%d\t%dms\t%s", r.Status, r.Milli, r.Path)
}

type responseWrapper struct {
	status int
	w      http.ResponseWriter
}

func (r *responseWrapper) Header() http.Header {
	return r.w.Header()
}

func (r *responseWrapper) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.w.Write(data)
}

func (r *responseWrapper) WriteHeader(status int) {
	r.status = status
	r.w.WriteHeader(status)
}

type tlbxKey struct{}
