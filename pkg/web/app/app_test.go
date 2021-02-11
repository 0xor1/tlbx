package app_test

import (
	"testing"

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
		nil,
		false,
		nil,
		nil,
		nil,
		nil)
	defer r.CleanUp()

	a := assert.New(t)
	c := r.NewClient()
	a.Equal("pong", (&app.Ping{}).MustDo(c))
}
