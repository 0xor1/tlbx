package app

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/server"
)

const (
	KB int64 = 1000
	MB int64 = 1000000
	GB int64 = 1000000000

	ApiPathPrefix        = "/api"
	ApiPathPrefixSegment = ApiPathPrefix + "/"
)

type SelfValidator interface {
	MustBeValid(tlbx Tlbx)
}

type Config struct {
	Log                     log.Log
	Version                 string
	StaticDir               string
	ProvideApiDocs          bool
	ContentSecurityPolicies []string
	// id
	IDGenPoolSize int
	// mdo
	MDoMax          int
	MDoMaxBodyBytes int64
	// tlbx
	TlbxSetup   TlbxMwares
	TlbxCleanup TlbxMwares
	// app
	Name        string
	Description string
	Endpoints   []*Endpoint
	Serve       func(http.HandlerFunc)
}

func JoinEps(epss ...[]*Endpoint) []*Endpoint {
	count := 0
	for _, eps := range epss {
		count += len(eps)
	}
	joined := make([]*Endpoint, 0, count)
	for _, eps := range epss {
		joined = append(joined, eps...)
	}
	return joined
}

type TlbxMware func(Tlbx)
type TlbxMwares []func(Tlbx)

func Run(configs ...func(*Config)) {
	c := config(configs...)
	mDoEp.MaxBodyBytes = c.MDoMaxBodyBytes
	// static file server
	staticFileDir, err := filepath.Abs(c.StaticDir)
	PanicOn(err)
	fileServer := http.FileServer(http.Dir(staticFileDir))
	// content-security-policy
	csps := strings.Join(append([]string{"default-src 'self'"}, c.ContentSecurityPolicies...), ";")
	// id pool
	idGenPool := NewIDGenPool(c.IDGenPoolSize)
	// endpoints
	c.Endpoints = JoinEps(defaultEps, c.Endpoints)
	router := make(map[string]*Endpoint, len(c.Endpoints))
	lDocsPath := strings.ToLower(ApiPathPrefix + docsEp.Path)
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
		path := ApiPathPrefix + ep.Path
		lPath := StrLower(path)
		_, exists := router[lPath]
		PanicIf(exists, "duplicate endpoint path: %q", path)
		router[lPath] = ep
		if !ep.IsPrivate {
			epDocs := &endpointDoc{
				Description:  ep.Description,
				Path:         path,
				Timeout:      ep.Timeout,
				MaxBodyBytes: ep.MaxBodyBytes,
				DefaultArgs:  ep.GetDefaultArgs(),
				ExampleArgs:  ep.GetExampleArgs(),
				ExampleRes:   ep.GetExampleResponse(),
			}
			if epDocs.DefaultArgs != nil {
				if _, ok := epDocs.DefaultArgs.(*UpStream); !ok {
					ti := &typeInfo{}
					getTypeInfo(reflect.TypeOf(epDocs.DefaultArgs), ti)
					ti.Ptr = false
					epDocs.ArgsTypes = ti
				} else {
					// in the case that its an stream just use the default json from the stream
					// itself explain the request structure
					epDocs.ArgsTypes = epDocs.DefaultArgs
					epDocs.DefaultArgs = nil
				}
			}
			if epDocs.ExampleRes != nil {
				if _, ok := epDocs.DefaultArgs.(*DownStream); !ok {
					ti := &typeInfo{}
					getTypeInfo(reflect.TypeOf(epDocs.ExampleRes), ti)
					ti.Ptr = false
					epDocs.ResTypes = ti
				} else {
					// in the case that its an stream just use the default json from the stream
					// itself explain the request structure
					epDocs.ResTypes = epDocs.ExampleRes
					epDocs.ExampleRes = nil
				}
			}
			docs.Endpoints = append(docs.Endpoints, epDocs)
		}
	}
	if c.ProvideApiDocs {
		// write docs to StaticDir/api/docs.json
		apiDocsDir := filepath.Join(c.StaticDir, `api`)
		PanicOn(os.MkdirAll(apiDocsDir, os.ModePerm))
		PanicOn(ioutil.WriteFile(filepath.Join(apiDocsDir, `docs.json`), json.MustMarshal(docs), os.ModePerm))
	}
	docs = nil
	// Handle requests!
	var root http.HandlerFunc
	root = func(w http.ResponseWriter, r *http.Request) {
		// tlbx
		tlbx := &tlbx{
			mDoMax:         c.MDoMax,
			root:           root,
			resp:           &responseWrapper{w: w},
			req:            r,
			start:          NowMilli(),
			idGenPool:      idGenPool,
			isSubMDo:       isSubMDo(r),
			log:            c.Log,
			actionStatsMtx: &sync.Mutex{},
			actionStats:    make([]*ActionStats, 0, 10),
			storeMtx:       &sync.RWMutex{},
			store:          map[interface{}]interface{}{},
		}
		tlbx.startMilli = tlbx.start.UnixNano() / 1000000
		// close body
		if tlbx.req != nil && tlbx.req.Body != nil {
			defer tlbx.req.Body.Close()
		}
		// log stats
		defer func() {
			tlbx.actionStatsMtx.Lock()
			defer tlbx.actionStatsMtx.Unlock()
			tlbx.log.Stats(&reqStats{
				Milli:   NowUnixMilli() - tlbx.startMilli,
				Status:  tlbx.resp.status,
				Method:  tlbx.req.Method,
				Path:    tlbx.req.URL.Path,
				Queries: tlbx.actionStats,
			})
		}()
		// recover from errors / redirects
		defer func() {
			if e := ToError(recover()); e != nil {
				if err, ok := e.Value().(*ErrMsg); ok {
					writeJson(tlbx, err.Status, err.Msg)
				} else if redirect, ok := e.Value().(*redirect); ok {
					http.Redirect(tlbx.resp, tlbx.req, redirect.url, redirect.status)
				} else {
					tlbx.log.ErrorOn(e)
					writeJson(tlbx, http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
				}
			}
		}()
		// set common headers
		tlbx.resp.Header().Set("X-Version", c.Version)
		// check method
		method := tlbx.req.Method
		BadReqIf(!(method == http.MethodPut || method == http.MethodGet || method == http.MethodPost), "only GET, PUT and POST methods are accepted")
		lPath := StrLower(tlbx.req.URL.Path)
		// tlbx mwares
		for _, setup := range c.TlbxSetup {
			setup(tlbx)
		}
		defer func() {
			for _, cleanup := range c.TlbxCleanup {
				cleanup(tlbx)
			}
		}()
		// serve static file
		if (method == http.MethodGet && !strings.HasPrefix(lPath, ApiPathPrefixSegment)) || lPath == lDocsPath {
			if lPath == lDocsPath {
				tlbx.req.Method = http.MethodGet
				tlbx.req.URL.Path += `.json`
			}
			// set common headers
			tlbx.resp.Header().Set("Cache-Control", "public, max-age=3600, immutable")
			tlbx.resp.Header().Set("X-Frame-Options", "DENY")
			tlbx.resp.Header().Set("X-XSS-Protection", "1; mode=block")
			tlbx.resp.Header().Set("Content-Security-Policy", csps)
			fileServer.ServeHTTP(tlbx.resp, tlbx.req)
			return
		}
		tlbx.resp.Header().Set("Cache-Control", "no-cache, no-store")
		// lower path now we have passed static file server
		tlbx.req.URL.Path = lPath

		// endpoints
		ep, exists := router[tlbx.req.URL.Path]
		ReturnIf(!exists, http.StatusNotFound, "")
		// check all requests have a X-Client header
		BadReqIf(!ep.SkipXClientCheck && tlbx.req.Header.Get("X-Client") == "", "X-Client header missing")

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
				_, ok := ep.GetExampleResponse().(*DownStream)
				BadReqIf(ok, "can not call stream endpoint in an mdo request")
			}
			// process args
			args := ep.GetDefaultArgs()
			s, isStream := args.(*UpStream)
			BadReqIf(isStream && tlbx.isSubMDo, "can not call stream endpoint in an mdo request")
			if isStream {
				s.Type = tlbx.req.Header.Get("Content-Type")
				s.Size = tlbx.req.ContentLength
				PanicOn(err)
				s.Name = tlbx.req.Header.Get("Content-Name")
				s.Content = tlbx.req.Body
				args = s
				argsStr := tlbx.req.Header.Get("Content-Args")
				if s.Args != nil && argsStr != "" {
					d := json.NewDecoder(bytes.NewBufferString(argsStr))
					d.DisallowUnknownFields()
					err = d.Decode(&s.Args)
					BadReqIf(err != nil, "error unmarshalling json: %s", err)
				}
			} else {
				getJsonArgs(tlbx, args)
			}
			// if args is a self validator, validate
			if sv, ok := args.(SelfValidator); ok {
				sv.MustBeValid(tlbx)
			}
			// handle request
			res := ep.Handler(tlbx, args)
			// process response
			if s, ok := res.(*DownStream); ok {
				defer s.Content.Close()
				BadReqIf(tlbx.isSubMDo, "can not call stream endpoint in an mdo request")
				tlbx.resp.Header().Add("Content-Type", s.Type)
				tlbx.resp.Header().Add("Content-Length", Strf("%d", s.Size))
				tlbx.resp.Header().Add("Content-Name", Strf("%s", s.Name))
				tlbx.resp.Header().Add("Content-Id", Strf("%s", s.ID))
				if s.IsDownload {
					tlbx.resp.Header().Add("Content-Disposition", Strf(`attachment; filename="%s"`, s.Name))
				}
				tlbx.resp.WriteHeader(http.StatusOK)
				_, err = io.Copy(tlbx.resp, s.Content)
				PanicOn(err)
			} else if resBs, ok := res.([]byte); ok {
				writeJsonRaw(tlbx, http.StatusOK, resBs)
			} else {
				writeJsonOk(tlbx, res)
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
				ReturnIf(true, http.StatusServiceUnavailable, "processing request has exceeded endpoint timeout: %dms", ep.Timeout)
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
		ProvideApiDocs:  true,
		IDGenPoolSize:   50,
		MDoMax:          20,
		MDoMaxBodyBytes: MB,
		Name:            "Web App",
		Description:     "A web app",
		Endpoints:       nil,
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

type ActionStats struct {
	Milli  int64  `json:"ms"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Action string `json:"action"`
}

type reqStats struct {
	Milli   int64          `json:"ms"`
	Status  int            `json:"status"`
	Method  string         `json:"method"`
	Path    string         `json:"path"`
	Queries []*ActionStats `json:"queries"`
}

func (r *reqStats) String() string {
	basic := Strf("%dms\t%d\t%s\t%s", r.Milli, r.Status, r.Method, r.Path)
	if len(r.Queries) == 0 {
		return basic
	}
	queries := make([]string, 0, len(r.Queries))
	for _, q := range r.Queries {
		queries = append(queries, Strf("%s\t%s\t%dms\t%s", q.Type, q.Name, q.Milli, q.Action))
	}
	return Strf("%s\n%s", basic, strings.Join(queries, "\n"))
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
	Start() time.Time
	StartMilli() int64
	Ctx() context.Context
	NewID() ID
	Log() log.Log
	LogActionStats(*ActionStats)
	// add any extra arbitrary stuff with these
	Get(key interface{}) interface{}
	Set(key, value interface{})
}

type tlbx struct {
	mDoMax         int
	root           http.HandlerFunc
	resp           *responseWrapper
	req            *http.Request
	start          time.Time
	startMilli     int64
	idGenPool      IDGenPool
	idGen          IDGen
	isSubMDo       bool
	log            log.Log
	actionStatsMtx *sync.Mutex
	actionStats    []*ActionStats
	storeMtx       *sync.RWMutex
	store          map[interface{}]interface{}
}

func (t *tlbx) Req() *http.Request {
	return t.req
}

func (t *tlbx) Resp() http.ResponseWriter {
	return t.resp
}

func (t *tlbx) Start() time.Time {
	return t.start
}

func (t *tlbx) StartMilli() int64 {
	return t.startMilli
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

func (t *tlbx) LogActionStats(as *ActionStats) {
	t.actionStatsMtx.Lock()
	defer t.actionStatsMtx.Unlock()
	t.actionStats = append(t.actionStats, as)
}

func Redirect(status int, url string) {
	PanicOn(&redirect{
		status: status,
		url:    url,
	})
}

func ReturnIf(condition bool, status int, format string, args ...interface{}) {
	if format == "" {
		format = http.StatusText(status)
	}
	if condition {
		PanicOn(&ErrMsg{
			Status: status,
			Msg:    Strf(format, args...),
		})
	}
}

func BadReqIf(condition bool, format string, args ...interface{}) {
	ReturnIf(condition, http.StatusBadRequest, format, args...)
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
	return Strf("status: %d, message: %s", e.Status, e.Msg)
}

func writeJsonOk(tlbx *tlbx, body interface{}) {
	writeJson(tlbx, http.StatusOK, body)
}

func writeJson(tlbx *tlbx, status int, body interface{}) {
	writeJsonRaw(tlbx, status, json.MustMarshal(body))
}

func writeJsonRaw(tlbx *tlbx, status int, body []byte) {
	tlbx.resp.Header().Set("Content-Type", json.ContentType)
	if !tlbx.isSubMDo && StrContains(tlbx.req.Header.Get("Accept-Encoding"), "gzip") {
		tlbx.resp.Header().Set("Content-Encoding", "gzip")
		tlbx.resp.WriteHeader(status)
		gz := gzip.NewWriter(tlbx.resp)
		_, err := gz.Write(body)
		PanicOn(err)
		PanicOn(gz.Close())
	} else {
		tlbx.resp.WriteHeader(status)
		_, err := tlbx.resp.Write(body)
		PanicOn(err)
	}
}

func isSubMDo(r *http.Request) bool {
	return r.URL.Query().Get("isSubMDo") == "true"
}

type MDoReq struct {
	Header bool       `json:"header,omitempty"`
	Path   string     `json:"path,omitempty"`
	Args   *json.Json `json:"args,omitempty"`
}

type MDoResp struct {
	Status int         `json:"status"`
	Header http.Header `json:"header,omitempty"`
	Body   *json.Json  `json:"body"`
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
	b := r.body.String()
	if b == "" || b == "<nil>" {
		b = "null"
	}
	if r.returnHeaders {
		h := json.MustMarshal(r.header)
		return []byte(Strf(`{"status":%d,"header":%s,"body":%s}`, r.status, string(h), b)), nil
	} else {
		return []byte(Strf(`{"status":%d,"body":%s}`, r.status, b)), nil
	}
}

type Endpoint struct {
	Description        string
	Path               string
	Timeout            int64
	SkipXClientCheck   bool
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
	Description  string      `json:"description"`
	Path         string      `json:"path"`
	Timeout      int64       `json:"timeout"`
	MaxBodyBytes int64       `json:"maxBodyBytes"`
	ArgsTypes    interface{} `json:"argsTypes"`
	ResTypes     interface{} `json:"resTypes"`
	DefaultArgs  interface{} `json:"defaultArgs"`
	ExampleArgs  interface{} `json:"exampleArgs"`
	ExampleRes   interface{} `json:"exampleRes"`
}

func checkErrForMaxBytes(tlbx *tlbx, err error) {
	if err != nil {
		ReturnIf(
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
		d := json.NewDecoder(bytes.NewBuffer(argsBytes))
		d.DisallowUnknownFields()
		err := d.Decode(args)
		if err != nil {
			if e, ok := err.(Error); ok {
				BadReqIf(err != nil, "error unmarshalling json: %s", e.Message())
			}
			BadReqIf(err != nil, "error unmarshalling json: %s", err)
		}
	}
}

// client stuff

type stream struct {
	Size    int64
	Type    string
	Name    string
	Content io.ReadCloser
}

type UpStream struct {
	stream
	Args interface{}
}

var streamDocs = map[string]interface{}{
	"body": "content bytes",
	"headers": map[string]string{
		"Content-Type":   "mime type",
		"Content-Length": "content bytes count",
		"Content-Name":   "name",
		"Content-Args":   "optional args json string",
	},
}

func (_ *UpStream) MarshalJSON() ([]byte, error) {
	return json.Marshal(streamDocs)
}

type DownStream struct {
	stream
	ID         ID
	IsDownload bool
}

func (s *UpStream) ToReq(method, url string) (*http.Request, error) {
	r, err := http.NewRequest(method, url, s.Content)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Type", s.Type)
	r.Header.Add("Content-Length", strconv.FormatInt(s.Size, 10))
	r.ContentLength = s.Size
	r.Header.Add("Content-Name", s.Name)
	as, err := json.Marshal(s.Args)
	if err != nil {
		return nil, err
	}
	r.Header.Add("Content-Args", string(as))
	return r, nil
}

func (s *UpStream) MustToReq(method, url string) *http.Request {
	r, err := s.ToReq(method, url)
	PanicOn(err)
	return r
}

func (s *DownStream) FromResp(r *http.Response) error {
	size, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
	if err != nil {
		return ToError(err)
	}
	var id ID
	contentID := r.Header.Get("Content-Id")
	if contentID != "" {
		id = MustParseID(contentID)
	}
	s.Type = r.Header.Get("Content-Type")
	s.Size = size
	s.Name = r.Header.Get("Content-Name")
	s.ID = id
	s.Content = r.Body
	return nil
}

func (s *DownStream) MustFromResp(r *http.Response) {
	PanicOn(s.FromResp(r))
}

func (_ *DownStream) MarshalJSON() ([]byte, error) {
	return json.Marshal(streamDocs)
}

type httpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	// protocol and host
	baseHref string
	http     httpClient
	cookies  map[string]string
}

func (c *Client) Cookies() map[string]string {
	res := map[string]string{}
	for k, v := range c.cookies {
		res[k] = v
	}
	return res
}

func NewClient(baseHref string, optClient ...httpClient) *Client {
	if len(optClient) == 0 {
		optClient = []httpClient{&http.Client{}}
	}
	return &Client{
		baseHref: baseHref,
		http:     optClient[0],
		cookies:  map[string]string{},
	}
}

func Call(c *Client, path string, args interface{}, res interface{}) error {
	url := c.baseHref + ApiPathPrefix + path
	method := http.MethodPut
	var req *http.Request
	var err error
	switch a := args.(type) {
	case *UpStream:
		if a != nil {
			req, err = a.ToReq(method, url)
		}
	default:
		argsBytes, err := json.Marshal(args)
		if err != nil {
			return ToError(err)
		}
		req, err = http.NewRequest(method, url, bytes.NewBuffer(argsBytes))
	}
	if req == nil {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(nil))
	}
	if err != nil {
		return ToError(err)
	}
	for name, value := range c.cookies {
		req.AddCookie(&http.Cookie{
			Name:  name,
			Value: value,
		})
	}
	req.Header.Set("X-Client", "tlbx-go-client")
	req.Header.Set("Accept-Encoding", "gzip")

	httpRes, err := c.http.Do(req)
	if err != nil {
		return ToError(err)
	}
	for _, cookie := range httpRes.Cookies() {
		c.cookies[cookie.Name] = cookie.Value
	}
	if s, ok := res.(**DownStream); ok {
		err = (*s).FromResp(httpRes)
		if err != nil {
			if s != nil && *s != nil && (*s).Content != nil {
				(*s).Content.Close()
				v := reflect.ValueOf(res)
				v.Elem().Set(reflect.Zero(v.Elem().Type()))
			}
			return ToError(err)
		}
		return nil
	}
	defer httpRes.Body.Close()
	var reader io.Reader
	reader = httpRes.Body
	if httpRes.Header.Get("Content-Encoding") == "gzip" {
		reader, err = gzip.NewReader(reader)
		if err != nil {
			return ToError(err)
		}
	}
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return ToError(err)
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
			return ToError(err)
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

type Ping struct{}

func (_ *Ping) Path() string {
	return "/ping"
}

func (a *Ping) Do(c *Client) (string, error) {
	res := ptr.String("")
	err := Call(c, a.Path(), a, &res)
	if res != nil {
		return *res, err
	}
	return "", err
}

func (a *Ping) MustDo(c *Client) string {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

var pingEp = &Endpoint{
	Description:      "ping the api server",
	Path:             (&Ping{}).Path(),
	Timeout:          500,
	MaxBodyBytes:     KB,
	SkipXClientCheck: true,
	GetDefaultArgs: func() interface{} {
		return nil
	},
	GetExampleArgs: func() interface{} {
		return nil
	},
	GetExampleResponse: func() interface{} {
		return "pong"
	},
	Handler: func(tlbx Tlbx, _ interface{}) interface{} {
		return "pong"
	},
}

type Docs struct{}

func (_ *Docs) Path() string {
	return "/docs"
}

func (a *Docs) Do(c *Client) (*json.Json, error) {
	res := &json.Json{}
	err := Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Docs) MustDo(c *Client) *json.Json {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

var docsEp = &Endpoint{
	Description:      "get the api docs",
	Path:             (&Docs{}).Path(),
	Timeout:          500,
	MaxBodyBytes:     KB,
	SkipXClientCheck: true,
	GetDefaultArgs: func() interface{} {
		return nil
	},
	GetExampleArgs: func() interface{} {
		return nil
	},
	GetExampleResponse: func() interface{} {
		return nil
	},
	Handler: func(t Tlbx, _ interface{}) interface{} {
		// this endpoint exists just for docs, it is handled by
		// the static file server as the docs json is written to file.
		return nil
	},
}

type MDo map[string]*MDoReq

func (_ *MDo) Path() string {
	return "/mdo"
}

func (a *MDo) Do(c *Client) (map[string]*MDoResp, error) {
	res := &map[string]*MDoResp{}
	err := Call(c, a.Path(), a, &res)
	if res != nil {
		return *res, err
	}
	return nil, err
}

func (a *MDo) MustDo(c *Client) map[string]*MDoResp {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

var mDoEp = &Endpoint{
	Description:      "perform multiple requests in parallel",
	Path:             (&MDo{}).Path(),
	Timeout:          2000,
	MaxBodyBytes:     MB,
	SkipXClientCheck: true,
	GetDefaultArgs: func() interface{} {
		return &MDo{}
	},
	GetExampleArgs: func() interface{} {
		return &MDo{
			"0": {
				Header: true,
				Path:   "/api/users/get",
				Args: json.FromInterface(map[string]interface{}{
					"nameStartsWith": "joe",
				}),
			},
			"1": {
				Path: "/api/users/me",
			},
			"2": {
				Path: "/api/users/notfound",
			},
		}
	},
	GetExampleResponse: func() interface{} {
		return &map[string]*MDoResp{
			"0": {
				Status: http.StatusOK,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: json.MustFromString(`{"id":2,"name":"joe bloggs"}`),
			},
			"1": {
				Status: http.StatusOK,
				Body:   json.MustFromString(`{"id":1,"name":"bob"}`),
			},
			"2": {
				Status: http.StatusNotFound,
				Body:   json.MustFromString(`"Not Found"`),
			},
		}
	},
	Handler: func(t Tlbx, a interface{}) interface{} {
		tlbx := t.(*tlbx)
		mDoReqsPtr := a.(*MDo)
		if mDoReqsPtr == nil {
			return nil
		}
		mDoReqs := *mDoReqsPtr
		BadReqIf(tlbx.req.Header.Get("X-Client") == "", "X-Client header missing")
		BadReqIf(len(mDoReqs) == 0, "empty mdo req")
		BadReqIf(len(mDoReqs) > tlbx.mDoMax, "too many mdo reqs, max reqs allowed: %d", tlbx.mDoMax)
		fullMDoResp := map[string]*mDoResp{}
		fullMDoRespMtx := &sync.Mutex{}
		does := make([]func(), 0, len(mDoReqs))
		for key := range mDoReqs {
			does = append(does, func(key string, mdoReq *MDoReq) func() {
				return func() {
					argsBytes, err := json.Marshal(mdoReq.Args)
					PanicOn(err)
					subReq, err := http.NewRequest(http.MethodPut, StrLower(mdoReq.Path)+"?isSubMDo=true", bytes.NewReader(argsBytes))
					PanicOn(err)
					PanicIf(subReq.URL.Path == ApiPathPrefix+(&MDo{}).Path(), "can't have mdo request inside an mdo request")
					PanicIf(!strings.HasPrefix(subReq.URL.Path, ApiPathPrefixSegment), "can't have none api request inside an mdo request")
					for _, c := range tlbx.req.Cookies() {
						subReq.AddCookie(c)
					}
					for name := range tlbx.req.Header {
						subReq.Header.Add(name, tlbx.req.Header.Get(name))
					}
					subResp := &mDoResp{returnHeaders: mdoReq.Header, header: http.Header{}, body: new(bytes.Buffer)}
					tlbx.root(subResp, subReq)
					fullMDoRespMtx.Lock()
					defer fullMDoRespMtx.Unlock()
					for _, val := range subResp.Header().Values("Set-Cookie") {
						tlbx.Resp().Header().Add("Set-Cookie", val)
					}
					fullMDoResp[key] = subResp
				}
			}(key, mDoReqs[key]))
		}
		PanicOn(GoGroup(does...))
		return fullMDoResp
	},
}

var defaultEps = []*Endpoint{
	pingEp,
	docsEp,
	mDoEp,
}

type typeInfo struct {
	Name      string        `json:"name,omitempty"`
	Ptr       bool          `json:"ptr,omitempty"`
	Array     bool          `json:"array,omitempty"`
	OmitEmpty bool          `json:"omitEmpty,omitempty"`
	Type      interface{}   `json:"type,omitempty"`
	Fields    []interface{} `json:"fields,omitempty"`
}

func getTypeInfo(t reflect.Type, ti *typeInfo) {
	if ti == nil {
		return
	}
	switch t.Kind() {
	case reflect.Ptr:
		ti.Ptr = true
		getTypeInfo(t.Elem(), ti)
	case reflect.Array, reflect.Slice:
		if t.Name() == "ID" || t.Name() == "Key" || t.Name() == "string" {
			ti.Type = StrLower(t.Name())
		} else {
			ti.Array = true
			getTypeInfo(t.Elem(), ti)
		}
	case reflect.Map:
		ti.Type = "map[" + t.Key().Name() + "]"
		fInfo := &typeInfo{}
		getTypeInfo(t.Elem(), fInfo)
		ti.Fields = append(ti.Fields, fInfo)
	case reflect.Struct:
		if t.Name() == "Time" || t.Name() == "Json" {
			ti.Type = StrLower(t.Name())
		} else {
			ti.Type = "struct"
			for i := 0; i < t.NumField(); i++ {
				f := t.Field(i)
				if f.Anonymous {
					getTypeInfo(f.Type, ti)
				} else {
					fInfo := &typeInfo{Name: f.Name}
					if f.Tag != "" {
						jsonParts := StrSplit(f.Tag.Get("json"), ",")
						if jsonParts[0] == "-" {
							fInfo = nil
						} else {
							if jsonParts[0] != "" {
								fInfo.Name = jsonParts[0]
							}
							fInfo.OmitEmpty = len(jsonParts) > 1 && jsonParts[1] == "omitempty"
						}
					}
					if fInfo != nil {
						getTypeInfo(f.Type, fInfo)
						ti.Fields = append(ti.Fields, fInfo)
					}
				}
			}
		}
	default:
		ti.Type = t.Kind().String()
	}
}
