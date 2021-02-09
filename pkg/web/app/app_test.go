package app_test

import (
	"testing"

	"github.com/0xor1/tlbx/pkg/web/app/user/usertest"
)

func Test(t *testing.T) {
	// use usereps here because it's a common bundled
	// set of endpoints that tests many of app.go
	// functionality
	usertest.Everything(t)

	// Now test all the functionality that usereps tests
	// doesnt use, i.e. mdo/upstreams/downstreams/redirects

	// TODO
}
