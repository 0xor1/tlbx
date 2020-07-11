package app

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/web/server"
)

const (
	KB int64 = 1000
	MB int64 = 1000000
	GB int64 = 1000000000

	ApiPathPrefix = "/api"
	docsPath      = ApiPathPrefix + "/docs"
	mdoPath       = ApiPathPrefix + "/mdo"
)

type Config struct {
	Log                     log.Log
	Version                 string
	StaticDir               string
	ContentSecurityPolicies []string
	// id
	IDGenPoolSize int
	// mdo
	MDoMax          int
	MDoMaxBodyBytes int64
	// tlbx
	TlbxMwares []func(Tlbx)
	// app
	Name        string
	Description string
	Endpoints   []*Endpoint
	Serve       func(http.HandlerFunc)
}

type TlbxMware func(Tlbx)
type TlbxMwares []func(Tlbx)

func Run(configs ...func(*Config)) {
	c := config(configs...)
	// static file server
	staticFileDir, err := filepath.Abs(c.StaticDir)
	PanicOn(err)
	fileServer := http.FileServer(http.Dir(staticFileDir))
	// content-security-policy
	csps := strings.Join(append([]string{"default-src 'self'"}, c.ContentSecurityPolicies...), ";")
	// id pool
	idGenPool := NewIDGenPool(c.IDGenPoolSize)
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
	ApiPathPrefixSegment := ApiPathPrefix + "/"
	// Handle requests!
	var root http.HandlerFunc
	root = func(w http.ResponseWriter, r *http.Request) {
		start := NowUnixMilli()
		// tlbx
		tlbx := &tlbx{
			resp:          &responseWrapper{w: w},
			req:           r,
			idGenPool:     idGenPool,
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
		tlbx.resp.Header().Set("Cache-Control", "no-cache, no-store")
		tlbx.resp.Header().Set("X-Version", c.Version)
		// check method
		method := tlbx.req.Method
		tlbx.BadReqIf(method != http.MethodGet && method != http.MethodPut, "only GET and PUT methods are accepted")
		lPath := strings.ToLower(tlbx.req.URL.Path)
		// tlbx mwares
		for _, mware := range c.TlbxMwares {
			mware(tlbx)
		}
		// serve static file
		if method == http.MethodGet && !strings.HasPrefix(lPath, ApiPathPrefixSegment) {
			// set common headers
			tlbx.resp.Header().Set("X-Frame-Options", "DENY")
			tlbx.resp.Header().Set("X-XSS-Protection", "1; mode=block")
			tlbx.resp.Header().Set("Content-Security-Policy", csps)
			fileServer.ServeHTTP(tlbx.resp, tlbx.req)
			return
		}
		// lower path now we have passed static file server
		tlbx.req.URL.Path = lPath
		// endpoint docs
		if lPath == docsPath {
			writeJsonRaw(tlbx.resp, http.StatusOK, docsBytes)
			return
		}
		// check all requests have a X-Client header
		tlbx.BadReqIf(tlbx.req.Header.Get("X-Client") == "", "X-Client header missing")
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
						for _, val := range subResp.Header().Values("Set-Cookie") {
							tlbx.Resp().Header().Add("Set-Cookie", val)
						}
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
				s.Size = tlbx.req.ContentLength
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
		Log:             l,
		Version:         "dev",
		StaticDir:       ".",
		IDGenPoolSize:   50,
		MDoMax:          20,
		MDoMaxBodyBytes: MB,
		Name:            "Web App",
		Description:     "A web app",
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
				Handler: func(tlbx Tlbx, args interface{}) interface{} {
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

type Tlbx interface {
	Req() *http.Request
	Resp() http.ResponseWriter
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

type tlbx struct {
	resp          *responseWrapper
	req           *http.Request
	idGenPool     IDGenPool
	idGen         IDGen
	isSubMDo      bool
	log           log.Log
	queryStatsMtx *sync.Mutex
	queryStats    []*QueryStats
	storeMtx      *sync.RWMutex
	store         map[interface{}]interface{}
}

func (t *tlbx) Req() *http.Request {
	return t.req
}

func (t *tlbx) Resp() http.ResponseWriter {
	return t.resp
}

func (t *tlbx) Ctx() context.Context {
	return t.req.Context()
}

func (t *tlbx) NewID() ID {
	if t.idGen == nil {
		t.idGen = t.idGenPool.Get()
	}
	return t.idGen.MustNew()
}

func (t *tlbx) Log() log.Log {
	return t.log
}

func (t *tlbx) LogQueryStats(qs *QueryStats) {
	t.queryStatsMtx.Lock()
	defer t.queryStatsMtx.Unlock()
	t.queryStats = append(t.queryStats, qs)
}

func (t *tlbx) Redirect(status int, url string) {
	PanicOn(&redirect{
		status: status,
		url:    url,
	})
}

func (t *tlbx) ExitIf(condition bool, status int, format string, args ...interface{}) {
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

func (t *tlbx) BadReqIf(condition bool, format string, args ...interface{}) {
	t.ExitIf(condition, http.StatusBadRequest, format, args...)
}

func (t *tlbx) Get(key interface{}) interface{} {
	t.storeMtx.RLock()
	defer t.storeMtx.RUnlock()
	return t.store[key]
}

func (t *tlbx) Set(key, value interface{}) {
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

type Endpoint struct {
	Description        string
	Path               string
	Timeout            int64
	MaxBodyBytes       int64
	IsPrivate          bool
	GetDefaultArgs     func() interface{}
	GetExampleArgs     func() interface{}
	GetExampleResponse func() interface{}
	Handler            func(tlbx Tlbx, args interface{}) interface{}
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

func checkErrForMaxBytes(tlbx *tlbx, err error) {
	if err != nil {
		tlbx.ExitIf(
			err.Error() == "http: request body too large",
			http.StatusRequestEntityTooLarge,
			"request body too large")
		PanicOn(err)
	}
}

func getJsonArgs(tlbx *tlbx, args interface{}) {
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
	r.ContentLength = s.Size
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
		if s != nil {
			req, err = s.ToReq(method, url)
		} else {
			req, err = http.NewRequest(method, url, nil)
		}
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
	req.Header.Set("X-Client", "tlbx-go-client")

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
