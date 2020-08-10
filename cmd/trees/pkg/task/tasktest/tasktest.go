package tasktest

import (
	"testing"

	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task/taskeps"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(
		config.Get(),
		append(projecteps.Eps, taskeps.Eps...),
		true,
		nil,
		projecteps.OnDelete,
		true,
		projecteps.OnSetSocials)
	defer r.CleanUp()

	a.NotNil(r)
}
