package blockerstest

import (
	"testing"

	"github.com/0xor1/wtf/cmd/boring/pkg/blockers"
	"github.com/0xor1/wtf/cmd/boring/pkg/blockers/blockerseps"
	"github.com/0xor1/wtf/cmd/boring/pkg/config"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app/common/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(config.Get(), blockerseps.Eps, false, nil)
	defer r.CleanUp()

	ali := test.NewClient()
	newGame := (&blockers.New{}).
		MustDo(ali)
	a.NotNil(newGame)
	NewIDGen()
}
