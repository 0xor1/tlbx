package gametest

import (
	"testing"

	"github.com/0xor1/wtf/cmd/pandemic/pkg/config"
	"github.com/0xor1/wtf/cmd/pandemic/pkg/game"
	"github.com/0xor1/wtf/cmd/pandemic/pkg/game/gameeps"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app/common/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(config.Get(), gameeps.Eps, nil)
	defer r.CleanUp()

	newGame := (&game.Create{
		Name: "new game",
	}).MustDo(r.Ali().Client())
	a.NotNil(newGame)
	NewIDGen()
}
