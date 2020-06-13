package app

import (
	"bytes"
	"context"
	"encoding/gob"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/json"
	"github.com/0xor1/wtf/pkg/log"
	"github.com/0xor1/wtf/pkg/web/server"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/sessions"
	"github.com/tomasen/realip"
)

const (
	KB int64 = 1000
	MB int64 = 1000000
	GB int64 = 1000000000

	ApiPathPrefix = "/api"
	docsPath      = ApiPathPrefix + "/docs"
	mdoPath       = ApiPathPrefix + "/mdo"
)

type mdoRootTlbxCtxKey struct{}

type Config struct {
	Log                     log.Log
	Version                 string
	StaticDir               string
	ContentSecurityPolicies []string
	// session
	SessionAuthKey64s [][]byte
	SessionEncrKey32s [][]byte
	SessionName       string
	SessionPath       string
	SessionDomain     string
	SessionMaxAge     int
	SessionSecure     bool
	SessionHttpOnly   bool
	SessionSameSite   http.SameSite
	// ratelimit
	RateLimitPerMinute   int
	RateLimitExitOnError bool
	RateLimiterPool      iredis.Pool
	// mdo
	MDoMax          int
	MDoMaxBodyBytes int64
	// tlbx
	ToolboxMware func(Toolbox)
	// app
	Name        string
	Description string
	Endpoints   []*Endpoint
	Serve       func(http.HandlerFunc)
}

func Run(configs ...func(*Config)) {
	c := config(configs...)
	// init session store
	sessionAuthEncrKeyPairs := make([][]byte, 0, len(c.SessionAuthKey64s)*2)
	for i := range c.SessionAuthKey64s {
		PanicIf(len(c.SessionAuthKey64s[i]) != 64, "authKey64s length is not 64")
		PanicIf(len(c.SessionEncrKey32s[i]) != 32, "encrKey32s length is not 32")
		sessionAuthEncrKeyPairs = append(sessionAuthEncrKeyPairs, c.SessionAuthKey64s[i], c.SessionEncrKey32s[i])
	}
	sessionStore := sessions.NewCookieStore(sessionAuthEncrKeyPairs...)
	sessionStore.Options.Path = c.SessionPath
	sessionStore.Options.Domain = c.SessionDomain
	sessionStore.Options.MaxAge = c.SessionMaxAge
	sessionStore.Options.Secure = c.SessionSecure
	sessionStore.Options.HttpOnly = c.SessionHttpOnly
	sessionStore.Options.SameSite = c.SessionSameSite
	// register types for sessionCookie
	gob.Register(NewIDGen().MustNew())
	gob.Register(time.Time{})
	// static file server
	staticFileDir, err := filepath.Abs(c.StaticDir)
	PanicOn(err)
	fileServer := http.FileServer(http.Dir(staticFileDir))
	// content-security-policy
	csps := strings.Join(append([]string{"default-src 'self'"}, c.ContentSecurityPolicies...), ";")
	// endpoints
	router := map[string]*Endpoint{}
	router[docsPath] = nil
	router[mdoPath] = nil
	docs := &endpointsDocs{
		Name:        c.Name,
		Description: c.Description,
		Endpoints:   make([]*endpointDoc, 0, len(c.Endpoints)),
	}
	for _, ep := range c.Endpoints {
		PanicIf(ep.Handler == nil,
			"endpoint: %q, missing Handler", ep.Path)
		PanicIf(ep.GetDefaultArgs == nil,
			"endpoint: %q, missing GetDefaultArgs", ep.Path)
		PanicIf(ep.GetExampleArgs == nil,
			"endpoint: %q, missing GetExampleArgs", ep.Path)
		PanicIf(ep.GetExampleResponse == nil,
			"endpoint: %q, missing GetExampleResponse", ep.Path)
		ep.Path = ApiPathPrefix + ep.Path
		path := strings.ToLower(ep.Path)
		_, exists := router[path]
		PanicIf(exists, "duplicate endpoint path: %q", path)
		router[path] = ep
		if !ep.IsPrivate {
			docs.Endpoints = append(docs.Endpoints, &endpointDoc{
				Description:     ep.Description,
				Path:            path,
				Timeout:         ep.Timeout,
				MaxBodyBytes:    ep.MaxBodyBytes,
				DefaultArgs:     ep.GetDefaultArgs(),
				ExampleArgs:     ep.GetExampleArgs(),
				ExampleResponse: ep.GetExampleResponse(),
			})
		}
	}
	delete(router, docsPath)
	delete(router, mdoPath)
	docsBytes := json.MustMarshal(docs)
	// Handle requests!
	var root http.HandlerFunc
	root = func(w http.ResponseWriter, r *http.Request) {
		start := NowUnixMilli()
		// toolbox
		tlbx := &toolbox{
			resp:          &responseWrapper{w: w},
			req:           r,
			isSubMDo:      isSubMDo(r),
			log:           c.Log,
			queryStatsMtx: &sync.Mutex{},
			queryStats:    make([]*QueryStats, 0, 10),
			storeMtx:      &sync.RWMutex{},
			store:         map[interface{}]interface{}{},
		}
		// close body
		if tlbx.req != nil && tlbx.req.Body != nil {
			defer tlbx.req.Body.Close()
		}
		// log stats
		defer func() {
			tlbx.queryStatsMtx.Lock()
			defer tlbx.queryStatsMtx.Unlock()
			tlbx.log.Stats(&reqStats{
				Milli:   NowUnixMilli() - start,
				Status:  tlbx.resp.status,
				Method:  tlbx.req.Method,
				Path:    tlbx.req.URL.Path,
				Queries: tlbx.queryStats,
			})
		}()
		// recover from errors / redirects
		defer func() {
			if e := ToError(recover()); e != nil {
				if err, ok := e.Value().(*ErrMsg); ok {
					writeJson(tlbx.resp, err.Status, err.Msg)
				} else if redirect, ok := e.Value().(*redirect); ok {
					http.Redirect(tlbx.resp, tlbx.req, redirect.url, redirect.status)
				} else {
					tlbx.log.ErrorOn(e)
					writeJson(tlbx.resp, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				}
			}
		}()
		// set common headers
		tlbx.resp.Header().Set("X-Frame-Options", "DENY")
		tlbx.resp.Header().Set("X-XSS-Protection", "1; mode=block")
		tlbx.resp.Header().Set("Content-Security-Policy", csps)
		tlbx.resp.Header().Set("Cache-Control", "no-cache, no-store")
		tlbx.resp.Header().Set("X-Version", c.Version)
		// check method
		method := tlbx.req.Method
		tlbx.BadReqIf(method != http.MethodGet && method != http.MethodPut, "only GET and PUT methods are accepted")
		// session
		gses, err := sessionStore.Get(tlbx.req, c.SessionName)
		PanicOn(err)
		tlbx.session = &session{
			r:       tlbx.req,
			w:       tlbx.resp,
			gorilla: gses,
			mtx:     &sync.RWMutex{},
		}
		if !tlbx.session.gorilla.IsNew {
			i, ok := tlbx.session.gorilla.Values["me"]
			if ok {
				me := i.(ID)
				tlbx.session.me = &me
				tlbx.session.isAuthed = true
			}
			i, ok = tlbx.session.gorilla.Values["authedOn"]
			if ok {
				tlbx.session.authedOn = i.(time.Time)
			}
		}
		// rate limiter
		rateLimit(c, tlbx)
		lPath := strings.ToLower(tlbx.req.URL.Path)
		// serve static file
		if method == http.MethodGet && !strings.HasPrefix(lPath, ApiPathPrefix+"/") {
			fileServer.ServeHTTP(tlbx.resp, tlbx.req)
			return
		}
		// check all requests have a X-Client header
		tlbx.BadReqIf(tlbx.req.Header.Get("X-Client") == "", "X-Client header missing")
		// lower path now we have passed static file server
		tlbx.req.URL.Path = lPath
		// endpoint docs
		if lPath == docsPath {
			writeJsonRaw(tlbx.resp, http.StatusOK, docsBytes)
			return
		}
		// toolbox mware
		if c.ToolboxMware != nil {
			c.ToolboxMware(tlbx)
		}
		// do mdo
		if lPath == mdoPath {
			if c.MDoMaxBodyBytes > 0 {
				tlbx.req.Body = http.MaxBytesReader(tlbx.resp, tlbx.req.Body, c.MDoMaxBodyBytes)
			}
			mDoReqs := map[string]*mDoReq{}
			getJsonArgs(tlbx, &mDoReqs)
			tlbx.BadReqIf(len(mDoReqs) > c.MDoMax, "too many mdo reqs, max reqs allowed: %d", c.MDoMax)
			fullMDoResp := map[string]*mDoResp{}
			fullMDoRespMtx := &sync.Mutex{}
			does := make([]func(), 0, len(mDoReqs))
			for key := range mDoReqs {
				does = append(does, func(key string, mdoReq *mDoReq) func() {
					return func() {
						argsBytes, err := json.Marshal(mdoReq.Args)
						PanicOn(err)
						subReq, err := http.NewRequest(http.MethodPut, strings.ToLower(mdoReq.Path)+"?isSubMDo=true", bytes.NewReader(argsBytes))
						PanicOn(err)
						PanicIf(subReq.URL.Path == mdoPath, "can't have mdo request inside an mdo request")
						// thread tlbx through ctx back on itself so a subMDoReq tlbx can get its
						// root tlbx so calls to tlbx.Session.Login() effect the root requests
						// cookies
						subReq = subReq.WithContext(context.WithValue(subReq.Context(), mdoRootTlbxCtxKey{}, tlbx))
						for _, c := range tlbx.req.Cookies() {
							subReq.AddCookie(c)
						}
						for name := range tlbx.req.Header {
							subReq.Header.Add(name, tlbx.req.Header.Get(name))
						}
						subResp := &mDoResp{returnHeaders: mdoReq.Headers, header: http.Header{}, body: new(bytes.Buffer)}
						root(subResp, subReq)
						fullMDoRespMtx.Lock()
						defer fullMDoRespMtx.Unlock()
						fullMDoResp[key] = subResp
					}
				}(key, mDoReqs[key]))
			}
			err = GoGroup(does...)
			PanicOn(err)
			writeJsonOk(tlbx.resp, fullMDoResp)
			return
		}
		// endpoints
		ep, exists := router[tlbx.req.URL.Path]
		tlbx.ExitIf(!exists, http.StatusNotFound, "")

		if ep.MaxBodyBytes > 0 {
			tlbx.req.Body = http.MaxBytesReader(tlbx.resp, tlbx.req.Body, ep.MaxBodyBytes)
		}

		// timeout
		ctx := tlbx.req.Context()
		cancel := func() {}
		timeout := time.Duration(ep.Timeout) * time.Millisecond
		if ep.Timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
			tlbx.req = tlbx.req.WithContext(ctx)
		}

		do := func() {
			// validation check
			if tlbx.isSubMDo {
				_, ok := ep.GetExampleResponse().(*Stream)
				tlbx.BadReqIf(ok, "can not call stream endpoint in an mdo request")
			}
			// process args
			args := ep.GetDefaultArgs()
			s, isStream := args.(*Stream)
			tlbx.BadReqIf(isStream && tlbx.isSubMDo, "can not call stream endpoint in an mdo request")
			if isStream {
				s.Type = tlbx.req.Header.Get("Content-Type")
				s.Size, err = strconv.ParseInt(tlbx.req.Header.Get("Content-Length"), 10, 64)
				PanicOn(err)
				s.Name = tlbx.req.Header.Get("Content-Name")
				contentID := tlbx.req.Header.Get("Content-Id")
				if contentID != "" {
					s.ID = MustParseID(contentID)
				}
				s.Content = tlbx.req.Body
				args = s
			} else {
				getJsonArgs(tlbx, args)
			}
			// handle request
			res := ep.Handler(tlbx, args)
			// process response
			if s, ok := res.(*Stream); ok {
				defer s.Content.Close()
				tlbx.BadReqIf(tlbx.isSubMDo, "can not call stream endpoint in an mdo request")
				tlbx.resp.Header().Add("Content-Type", s.Type)
				tlbx.resp.Header().Add("Content-Length", Sprintf("%d", s.Size))
				tlbx.resp.Header().Add("Content-Name", Sprintf("%s", s.Name))
				tlbx.resp.Header().Add("Content-Id", Sprintf("%s", s.ID))
				if s.IsDownload {
					tlbx.resp.Header().Add("Content-Disposition", Sprintf(`attachment; filename="%s"`, s.Name))
				}
				tlbx.resp.WriteHeader(http.StatusOK)
				_, err = io.Copy(tlbx.resp, s.Content)
				PanicOn(err)
			} else {
				writeJsonOk(tlbx.resp, res)
			}
			cancel()
		}

		if ep.Timeout > 0 {
			errCh := make(chan interface{})
			Go(do, func(err interface{}) {
				errCh <- err
			})
			select {
			case err := <-errCh:
				PanicOn(err)
			case <-ctx.Done():
				return
			case <-time.After(timeout):
				tlbx.ExitIf(true, http.StatusServiceUnavailable, "processing request has exceeded endpoint timeout: %dms", ep.Timeout)
			}
		} else {
			do()
		}
	}
	c.Serve(root)
}

func config(configs ...func(*Config)) *Config {
	l := log.New()
	c := &Config{
		Log:                  l,
		Version:              "dev",
		StaticDir:            ".",
		SessionAuthKey64s:    [][]byte{},
		SessionEncrKey32s:    [][]byte{},
		SessionName:          "s",
		SessionPath:          "/",
		SessionDomain:        "",
		SessionMaxAge:        0,
		SessionSecure:        false,
		SessionHttpOnly:      true,
		SessionSameSite:      http.SameSiteDefaultMode,
		RateLimitPerMinute:   120,
		RateLimitExitOnError: false,
		RateLimiterPool:      nil,
		MDoMax:               20,
		MDoMaxBodyBytes:      MB,
		Name:                 "Web App",
		Description:          "A web app",
		Endpoints: []*Endpoint{
			{
				Description:  "A test endpoint to echo back the args",
				Path:         "/test/echo",
				Timeout:      100,
				MaxBodyBytes: MB,
				IsPrivate:    false,
				GetDefaultArgs: func() interface{} {
					return &map[string]interface{}{}
				},
				GetExampleArgs: func() interface{} {
					return &map[string]interface{}{
						"a": "ali",
						"b": "bob",
						"c": "cat",
					}
				},
				GetExampleResponse: func() interface{} {
					return &map[string]interface{}{
						"a": "ali",
						"b": "bob",
						"c": "cat",
					}
				},
				Handler: func(tlbx Toolbox, args interface{}) interface{} {
					return args
				},
			},
		},
		Serve: func(h http.HandlerFunc) {
			server.Run(func(c *server.Config) {
				c.Log = l
				c.Handler = h
			})
		},
	}
	for _, config := range configs {
		config(c)
	}
	return c
}

type QueryStats struct {
	Milli int64  `json:"ms"`
	Query string `json:"query"`
}

type reqStats struct {
	Milli   int64         `json:"ms"`
	Status  int           `json:"status"`
	Method  string        `json:"method"`
	Path    string        `json:"path"`
	Queries []*QueryStats `json:"queries"`
}

func (r *reqStats) String() string {
	basic := Sprintf("%dms\t%d\t%s\t%s", r.Milli, r.Status, r.Method, r.Path)
	if len(r.Queries) == 0 {
		return basic
	}
	queries := make([]string, 0, len(r.Queries))
	for _, q := range r.Queries {
		queries = append(queries, Sprintf("%dms\t%s", q.Milli, q.Query))
	}
	return Sprintf("%s\n%s", basic, strings.Join(queries, "\n"))
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

type Toolbox interface {
	Me() ID
	Session() Session
	Ctx() context.Context
	NewID() ID
	Log() log.Log
	LogQueryStats(*QueryStats)
	Redirect(status int, url string)
	ExitIf(condition bool, status int, format string, args ...interface{})
	BadReqIf(condition bool, format string, args ...interface{})
	// add any extra arbitrary stuff with these
	Get(key interface{}) interface{}
	Set(key, value interface{})
}

type toolbox struct {
	resp          *responseWrapper
	req           *http.Request
	idGen         IDGen
	session       *session
	isSubMDo      bool
	log           log.Log
	queryStatsMtx *sync.Mutex
	queryStats    []*QueryStats
	storeMtx      *sync.RWMutex
	store         map[interface{}]interface{}
}

func (t *toolbox) Me() ID {
	return t.Session().Me()
}

func (t *toolbox) Session() Session {
	return t.session
}

func (t *toolbox) Ctx() context.Context {
	return t.req.Context()
}

func (t *toolbox) NewID() ID {
	if t.idGen == nil {
		t.idGen = NewIDGen()
	}
	return t.idGen.MustNew()
}

func (t *toolbox) Log() log.Log {
	return t.log
}

func (t *toolbox) LogQueryStats(qs *QueryStats) {
	t.queryStatsMtx.Lock()
	defer t.queryStatsMtx.Unlock()
	t.queryStats = append(t.queryStats, qs)
}

func (t *toolbox) Redirect(status int, url string) {
	PanicOn(&redirect{
		status: status,
		url:    url,
	})
}

func (t *toolbox) ExitIf(condition bool, status int, format string, args ...interface{}) {
	if format == "" {
		format = http.StatusText(status)
	}
	if condition {
		PanicOn(&ErrMsg{
			Status: status,
			Msg:    Sprintf(format, args...),
		})
	}
}

func (t *toolbox) BadReqIf(condition bool, format string, args ...interface{}) {
	t.ExitIf(condition, http.StatusBadRequest, format, args...)
}

func (t *toolbox) Get(key interface{}) interface{} {
	t.storeMtx.RLock()
	defer t.storeMtx.RUnlock()
	return t.store[key]
}

func (t *toolbox) Set(key, value interface{}) {
	t.storeMtx.Lock()
	defer t.storeMtx.Unlock()
	t.store[key] = value
}

type redirect struct {
	status int
	url    string
}

type ErrMsg struct {
	Status int    `json:"status"`
	Msg    string `json:"message"`
}

func (e *ErrMsg) Error() string {
	return Sprintf("status: %d, message: %s", e.Status, e.Msg)
}

func writeJsonOk(w http.ResponseWriter, body interface{}) {
	writeJson(w, http.StatusOK, body)
}

func writeJson(w http.ResponseWriter, status int, body interface{}) {
	bodyBytes, err := json.Marshal(body)
	PanicOn(err)
	writeJsonRaw(w, status, bodyBytes)
}

func writeJsonRaw(w http.ResponseWriter, status int, body []byte) {
	w.Header().Set("Content-Type", json.ContentType)
	w.WriteHeader(status)
	_, err := w.Write(body)
	PanicOn(err)
}

func isSubMDo(r *http.Request) bool {
	return r.URL.Query().Get("isSubMDo") == "true"
}

type mDoReq struct {
	Headers bool                   `json:"headers"`
	Path    string                 `json:"path"`
	Args    map[string]interface{} `json:"args"`
}

type mDoResp struct {
	returnHeaders bool
	status        int
	header        http.Header
	body          *bytes.Buffer
}

func (r *mDoResp) Header() http.Header {
	return r.header
}

func (r *mDoResp) Write(data []byte) (int, error) {
	return r.body.Write(data)
}

func (r *mDoResp) WriteHeader(status int) {
	r.status = status
}

func (r *mDoResp) MarshalJSON() ([]byte, error) {
	if r.returnHeaders {
		h, err := json.Marshal(r.header)
		PanicOn(err)
		return []byte(Sprintf(`{"status":%d,"header":%s,"body":%s}`, r.status, h, r.body)), nil
	} else {
		return []byte(Sprintf(`{"status":%d,"body":%s}`, r.status, r.body)), nil
	}
}

type Session interface {
	IsAuthed() bool
	Me() ID
	AuthedOn() time.Time
	Login(ID)
	Logout()
}

type session struct {
	w        http.ResponseWriter
	r        *http.Request
	isAuthed bool
	me       *ID
	authedOn time.Time
	gorilla  *sessions.Session
	mtx      *sync.RWMutex
}

func (s *session) IsAuthed() bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.isAuthed
}

func (s *session) Me() ID {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	if !s.isAuthed {
		PanicOn(&ErrMsg{
			Status: http.StatusUnauthorized,
			Msg:    http.StatusText(http.StatusUnauthorized),
		})
	}
	return *s.me
}

func (s *session) AuthedOn() time.Time {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.authedOn
}

func (s *session) Login(me ID) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if isSubMDo(s.r) {
		// get root tlbx from ctx to login on root
		rootTlbxI := s.r.Context().Value(mdoRootTlbxCtxKey{})

		PanicIf(rootTlbxI == nil, "isSubMDo but rootTlbx is nil in req ctx")
		rootTlbx := rootTlbxI.(*toolbox)
		rootTlbx.session.Login(me)
	}
	s.isAuthed = true
	s.me = &me
	s.authedOn = Now()
	s.gorilla.Values = map[interface{}]interface{}{
		"me":       me,
		"authedOn": s.authedOn,
	}
	PanicOn(s.gorilla.Save(s.r, s.w))
}

func (s *session) Logout() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	if isSubMDo(s.r) {
		// get root tlbx from ctx to login on root
		rootTlbxI := s.r.Context().Value(mdoRootTlbxCtxKey{})
		PanicIf(rootTlbxI == nil, "isSubMDo but rootTlbx is nil in req ctx")
		rootTlbx := rootTlbxI.(*toolbox)
		rootTlbx.session.Logout()
	}
	s.isAuthed = false
	s.me = nil
	s.authedOn = time.Time{}
	s.gorilla.Options.MaxAge = -1
	s.gorilla.Values = map[interface{}]interface{}{}
	PanicOn(s.gorilla.Save(s.r, s.w))
}

type Endpoint struct {
	Description        string
	Path               string
	Timeout            int64
	MaxBodyBytes       int64
	IsPrivate          bool
	GetDefaultArgs     func() interface{}
	GetExampleArgs     func() interface{}
	GetExampleResponse func() interface{}
	Handler            func(tlbx Toolbox, args interface{}) interface{}
}

func ExampleID() ID {
	id := ID{}
	id.UnmarshalText([]byte("01DWWXG07ZKYXGWJFP1XMBM45C"))
	return id
}

func ExampleTime() time.Time {
	t, err := time.Parse("2006-01-02 15:04:05 +0000 UTC", "2020-03-03 17:00:00 +0000 UTC")
	PanicOn(err)
	return t
}

type endpointsDocs struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Endpoints   []*endpointDoc `json:"endpoints"`
}

type endpointDoc struct {
	Description     string      `json:"description"`
	Path            string      `json:"path"`
	Timeout         int64       `json:"timeoutMilli"`
	MaxBodyBytes    int64       `json:"maxBodyBytes"`
	ArgsTypes       interface{} `json:"argsTypes"`
	DefaultArgs     interface{} `json:"defaultArgs"`
	ExampleArgs     interface{} `json:"exampleArgs"`
	ExampleResponse interface{} `json:"exampleResponse"`
}

func checkErrForMaxBytes(tlbx *toolbox, err error) {
	if err != nil {
		tlbx.ExitIf(
			err.Error() == "http: request body too large",
			http.StatusRequestEntityTooLarge,
			"request body too large")
		PanicOn(err)
	}
}

func rateLimit(c *Config, tlbx *toolbox) {
	if c.RateLimiterPool == nil || c.RateLimitPerMinute < 1 {
		return
	}

	shouldReturn := func(err error) bool {
		if err != nil {
			if c.RateLimitExitOnError {
				PanicOn(err)
			}
			c.Log.ErrorOn(err)
			return true
		}
		return false
	}

	remaining := c.RateLimitPerMinute

	defer func() {
		tlbx.resp.Header().Add("X-Rate-Limit-Limit", strconv.Itoa(c.RateLimitPerMinute))
		tlbx.resp.Header().Add("X-Rate-Limit-Remaining", strconv.Itoa(remaining))
		tlbx.resp.Header().Add("X-Rate-Limit-Reset", "60")

		tlbx.ExitIf(remaining < 1, http.StatusTooManyRequests, "")
	}()

	// get key
	var key string
	if tlbx.session.me != nil {
		key = tlbx.session.me.String()
	}
	key = Sprintf("rate-limiter-%s-%s", realip.RealIP(tlbx.req), key)

	now := NowUnixNano()
	cnn := c.RateLimiterPool.Get()
	defer cnn.Close()

	err := cnn.Send("MULTI")
	if shouldReturn(err) {
		return
	}

	err = cnn.Send("ZREMRANGEBYSCORE", key, 0, now-time.Minute.Nanoseconds())
	if shouldReturn(err) {
		return
	}

	err = cnn.Send("ZRANGE", key, 0, -1)
	if shouldReturn(err) {
		return
	}

	results, err := redis.Values(cnn.Do("EXEC"))
	if shouldReturn(err) {
		return
	}

	keys, err := redis.Strings(results[len(results)-1], err)
	if shouldReturn(err) {
		return
	}

	remaining = remaining - len(keys)

	if remaining > 0 {
		remaining--

		err := cnn.Send("MULTI")
		if shouldReturn(err) {
			return
		}

		err = cnn.Send("ZADD", key, now, now)
		if shouldReturn(err) {
			return
		}

		err = cnn.Send("EXPIRE", key, 60)
		if shouldReturn(err) {
			return
		}

		_, err = cnn.Do("EXEC")
		if shouldReturn(err) {
			return
		}
	}
}

func getJsonArgs(tlbx *toolbox, args interface{}) {
	if args == nil {
		return
	}
	argsBytes := []byte(tlbx.req.URL.Query().Get("args"))
	if len(argsBytes) == 0 {
		var err error
		argsBytes, err = ioutil.ReadAll(tlbx.req.Body)
		checkErrForMaxBytes(tlbx, err)
	}
	if len(argsBytes) > 0 {
		err := json.Unmarshal(argsBytes, args)
		tlbx.BadReqIf(err != nil, "error unmarshalling json: %s", err)
	}
}

// client stuff

type Stream struct {
	Type       string
	Name       string
	Size       int64
	ID         ID
	IsDownload bool
	Content    io.ReadCloser
}

func (s *Stream) ToReq(method, url string) (*http.Request, error) {
	r, err := http.NewRequest(method, url, s.Content)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", s.Type)
	r.Header.Add("Content-Length", strconv.FormatInt(s.Size, 10))
	r.Header.Add("Content-Name", s.Name)
	r.Header.Add("Content-Id", s.ID.String())
	if s.IsDownload {
		r.Header.Add("Content-Disposition", Sprintf(`attachment; filename="%s"`, s.Name))
	}
	return r, nil
}

func (bs *Stream) MustToReq(method, url string) *http.Request {
	r, err := bs.ToReq(method, url)
	PanicOn(err)
	return r
}

func (s *Stream) FromResp(r *http.Response) error {
	size, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return err
	}
	var id ID
	contentID := r.Header.Get("Content-Id")
	if contentID != "" {
		id = MustParseID(contentID)
	}
	*s = Stream{
		Type:    r.Header.Get("Content-Type"),
		Size:    size,
		Name:    r.Header.Get("Content-Name"),
		ID:      id,
		Content: r.Body,
	}
	return nil
}

func (s *Stream) MustFromResp(r *http.Response) {
	PanicOn(s.FromResp(r))
}

func (_ *Stream) MarshalJSON() ([]byte, error) {
	return json.Marshal(`body contains content bytes plus headers:
"Content-Type": "mime_type",
"Content-Length": bytes_count,
"Content-Name": "name"
"Content-Id": "id_string"`)
}

type Client struct {
	// protocol and host
	baseHref string
	http     *http.Client
	cookies  map[string]string
}

func NewClient(baseHref string) *Client {
	return &Client{
		baseHref: baseHref,
		http:     &http.Client{},
		cookies:  map[string]string{},
	}
}

func Call(c *Client, path string, args interface{}, res interface{}) error {
	url := c.baseHref + ApiPathPrefix + path
	method := http.MethodPut
	var req *http.Request
	var err error
	if s, ok := args.(*Stream); ok {
		req, err = s.ToReq(method, url)
	} else if args != nil {
		argsBytes, err := json.Marshal(args)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(argsBytes))
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return err
	}
	for name, value := range c.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}
	req.Header.Set("X-Client", "wtf-go-client")

	httpRes, err := c.http.Do(req)
	if err != nil {
		return err
	}
	for _, cookie := range httpRes.Cookies() {
		c.cookies[cookie.Name] = cookie.Value
	}
	if s, ok := res.(*Stream); ok {
		err = s.FromResp(httpRes)
		if err != nil {
			return err
		}
		*res.(*Stream) = *s
		return nil
	}
	defer httpRes.Body.Close()
	bs, err := ioutil.ReadAll(httpRes.Body)
	if err != nil {
		return err
	}
	if httpRes.StatusCode >= 400 {
		if res != nil {
			v := reflect.ValueOf(res)
			v.Elem().Set(reflect.Zero(v.Elem().Type()))
		}
		msg := &ErrMsg{
			Status: httpRes.StatusCode,
		}
		err = json.Unmarshal(bs, &msg.Msg)
		if err != nil {
			return err
		}
		return msg
	}
	if res == nil {
		return nil
	}
	if len(bs) == 0 || string(bs) == "null" {
		v := reflect.ValueOf(res)
		v.Elem().Set(reflect.Zero(v.Elem().Type()))
		return nil
	}
	return json.Unmarshal(bs, res)
}

func MustCall(c *Client, path string, args interface{}, res interface{}) {
	PanicOn(Call(c, path, args, res))
}
