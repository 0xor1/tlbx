package app_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
		[]*app.Endpoint{},
		false,
		nil,
		nil,
		nil,
		nil)
	defer r.CleanUp()

	a := assert.New(t)
	c := r.NewClient()
	a.Equal("pong", (&app.Ping{}).MustDo(c))

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
		},
	)).MustToReader())
	req.Header.Add("X-Client", "tlbx-app-tests")
	PanicOn(err)
	w := httptest.NewRecorder()
	r.RootHandler().ServeHTTP(w, req)
	Println(string(w.Body.Bytes()))
	a.Equal(http.StatusOK, w.Result().StatusCode)
}
