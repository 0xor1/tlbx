package endpoint

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/json"
	"github.com/0xor1/wtf/pkg/web/returnerror"
)

func Serve(isDocs func(*http.Request) bool, eps *Endpoints) http.HandlerFunc {
	router := map[string]*Endpoint{}
	docs := &endpointsDocs{
		Name:        eps.Name,
		Description: eps.Description,
		Endpoints:   make([]*endpointDoc, 0, len(eps.Endpoints)),
	}
	for _, ep := range eps.Endpoints {
		PanicIf(ep.Handler == nil,
			"endpoint: %q, missing Handler", ep.Path)
		PanicIf(ep.GetDefaultArgs == nil,
			"endpoint: %q, missing GetDefaultArgs", ep.Path)
		PanicIf(ep.GetExampleArgs == nil,
			"endpoint: %q, missing GetExampleArgs", ep.Path)
		PanicIf(ep.GetExampleResponse == nil,
			"endpoint: %q, missing GetExampleResponse", ep.Path)
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
	docsBytes := json.MustMarshalIndent(docs, "", "    ")
	return func(w http.ResponseWriter, r *http.Request) {
		if isDocs(r) {
			json.WriteHttpRaw(w, http.StatusOK, docsBytes)
			return
		}

		ep, exists := router[r.URL.Path]
		returnerror.If(!exists, http.StatusNotFound, http.StatusText(http.StatusNotFound))

		if ep.MaxBodyBytes > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, ep.MaxBodyBytes)
		}

		ctx := r.Context()
		cancel := func() {}
		timeout := time.Duration(ep.Timeout) * time.Millisecond
		if ep.Timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
			r.WithContext(ctx)
		}

		do := func() {
			argsStr := r.URL.Query().Get("args")
			argsBytes := []byte(argsStr)
			if argsStr == "" {
				var err error
				argsBytes, err = ioutil.ReadAll(r.Body)
				PanicOn(err)
			}
			args := ep.GetDefaultArgs()
			if len(argsBytes) > 0 {
				err := json.Unmarshal(argsBytes, args)
				returnerror.If(err != nil, http.StatusBadRequest, "error unmarshalling json: %s", err)
			}
			res := ep.Handler(r, args)
			if bs, ok := res.(*ByteStream); ok {
				w.Header().Add("Content-Type", bs.Type)
				w.Header().Add("Content-Length", Sprintf("%d", bs.Size))
				w.Header().Add("Name", bs.Name)
				w.WriteHeader(http.StatusOK)
				_, err := io.Copy(w, bs.Stream)
				PanicOn(err)
				return
			}
			json.WriteHttpOk(w, res)
			cancel()
		}

		if ep.Timeout > 0 {
			errCh := make(chan Error)
			Go(do, func(err Error) {
				errCh <- err
			})
			select {
			case err := <-errCh:
				PanicOn(err)
			case <-ctx.Done():
				return
			case <-time.After(timeout):
				json.WriteHttp(w, http.StatusServiceUnavailable, http.StatusText(http.StatusServiceUnavailable))
			}
		} else {
			do()
		}
	}
}

type Endpoints struct {
	Name        string
	Description string
	Endpoints   []*Endpoint
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
	Handler            func(r *http.Request, args interface{}) interface{}
}

type endpointsDocs struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Endpoints   []*endpointDoc `json:"endpoints"`
}

type endpointDoc struct {
	Description     string      `json:"description"`
	Path            string      `json:"path"`
	Timeout         int64       `json:"timeout"`
	MaxBodyBytes    int64       `json:"maxBodyBytes"`
	DefaultArgs     interface{} `json:"defaultArgs"`
	ExampleArgs     interface{} `json:"exampleArgs"`
	ExampleResponse interface{} `json:"exampleResponse"`
}

type ByteStream struct {
	Type   string        `json:"type"`
	Name   string        `json:"name"`
	Size   int64         `json:"size"`
	Stream io.ReadCloser `json:"-"`
}

func (bs *ByteStream) MarshalJSON() ([]byte, error) {
	return []byte(
		`body contains content bytes plus required headers:
"Content-Type": "mime_type",
"Content-Length": bytes_count,
"Name": "name"`), nil
}

func ExampleID() ID {
	id := ID{}
	id.UnmarshalText([]byte("01DWWXG07ZKYXGWJFP1XMBM45C"))
	return id
}
