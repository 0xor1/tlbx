package projecttest

import (
	"testing"

	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(config.Get(), projecteps.Eps, true, projecteps.OnActivate, projecteps.OnDelete, true, func(tlbx app.Tlbx, id ID, alias *string) error { return nil }, true, func(tlbx app.Tlbx, id ID, hasAvatar bool) error { return nil })
	defer r.CleanUp()

	a.NotNil(r)
}
