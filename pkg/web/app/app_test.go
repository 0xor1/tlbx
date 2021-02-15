package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/tlbx/pkg/web/app/user/usertest"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	// use usereps here because it's a common bundled
	// set of endpoints that tests many of app.go
	// functionality
	usertest.Everything(t)

	// Now test all the functionality that usereps tests
	// doesnt use, i.e. mdo/upstreams/downstreams/redirects

	r := test.NewRig(
		config.GetProcessed(config.GetBase()),
		[]*app.Endpoint{
			{
				Description:  "echo back the json obj args",
				Path:         "/echo",
				Timeout:      500,
				MaxBodyBytes: app.KB,
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
				Handler: func(tlbx app.Tlbx, args interface{}) interface{} {
					return args
				},
			},
			{
				Description:  "redirect",
				Path:         "/redirect",
				Timeout:      500,
				MaxBodyBytes: app.KB,
				GetDefaultArgs: func() interface{} {
					return nil
				},
				GetExampleArgs: func() interface{} {
					return nil
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, args interface{}) interface{} {
					app.Redirect(http.StatusMovedPermanently, "https://github.com/0xor1/tlbx")
					return nil
				},
			},
			{
				Description:  "timeout",
				Path:         "/timeout",
				Timeout:      100,
				MaxBodyBytes: app.KB,
				GetDefaultArgs: func() interface{} {
					return nil
				},
				GetExampleArgs: func() interface{} {
					return nil
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, args interface{}) interface{} {
					time.Sleep(150 * time.Millisecond)
					return nil
				},
			},
			{
				Description:  "panic",
				Path:         "/panic",
				Timeout:      100,
				MaxBodyBytes: app.KB,
				GetDefaultArgs: func() interface{} {
					return nil
				},
				GetExampleArgs: func() interface{} {
					return nil
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, args interface{}) interface{} {
					PanicOn("yolo")
					return nil
				},
			},
		},
		false,
		nil,
		nil,
		nil,
		nil,
		false)
	defer r.CleanUp()

	a := assert.New(t)
	c := r.NewClient()
	a.Equal("pong", (&app.Ping{}).MustDo(c))

	// test mdo and basic reponse statuses
	req, err := http.NewRequest(http.MethodPut, "/api/mdo", json.MustFromBytes(json.MustMarshal(
		map[string]interface{}{
			"0": map[string]interface{}{
				"path": "/api/ping",
			},
			"1": map[string]interface{}{
				"path": "/api/echo",
				"args": map[string]interface{}{
					"msg": "yolo",
				},
			},
			"2": map[string]interface{}{
				"path": "/api/redirect",
			},
			"3": map[string]interface{}{
				"path": "/api/timeout",
			},
			"4": map[string]interface{}{
				"path": "/api/panic",
			},
		},
	)).MustToReader())
	req.Header.Add("X-Client", "tlbx-app-tests")
	PanicOn(err)
	w := httptest.NewRecorder()
	r.RootHandler().ServeHTTP(w, req)
	a.Equal(http.StatusOK, w.Result().StatusCode)
	body := json.MustFromReader(w.Body)
	a.Equal(http.StatusOK, body.MustInt("0", "status"))
	a.Equal(http.StatusOK, body.MustInt("1", "status"))
	a.Equal(http.StatusMovedPermanently, body.MustInt("2", "status"))
	a.Equal(http.StatusServiceUnavailable, body.MustInt("3", "status"))
	a.Equal(http.StatusInternalServerError, body.MustInt("4", "status"))

	// test static file headers
	req, err = http.NewRequest(http.MethodGet, "/notfound", nil)
	req.Header.Add("X-Client", "tlbx-app-tests")
	PanicOn(err)
	w = httptest.NewRecorder()
	r.RootHandler().ServeHTTP(w, req)
	Println(string(w.Body.Bytes()))
	a.Equal(http.StatusNotFound, w.Result().StatusCode)
	a.Equal(w.Header().Get("Cache-Control"), "public, max-age=3600, immutable")
	a.Equal(w.Header().Get("X-Frame-Options"), "DENY")
	a.Equal(w.Header().Get("X-XSS-Protection"), "1; mode=block")
	a.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'self'")
}
