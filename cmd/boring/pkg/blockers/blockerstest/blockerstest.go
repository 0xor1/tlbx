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
	var err error

	// ali creates a new game
	aliGame := (&blockers.New{}).
		MustDo(r.Ali().Client())
	a.NotNil(aliGame)

	// bob joins ali's game
	bobGame := (&blockers.Join{Game: aliGame.ID}).
		MustDo(r.Bob().Client())
	a.NotNil(bobGame)

	// ali starts the game
	aliGame = (&blockers.Start{}).
		MustDo(r.Ali().Client())
	a.NotNil(aliGame)

	// ali takes her first turn missing her first corner cell
	aliGame, err = (&blockers.TakeTurn{
		PieceIdx: 0,
		Position: 1,
	}).Do(r.Ali().Client())
	a.Nil(aliGame)
	a.Regexp("first corner constraint not met", err)

	// ali takes a valid first turn with her first color
	aliGame = (&blockers.TakeTurn{
		PieceIdx: 0,
		Position: 0,
	}).MustDo(r.Ali().Client())
	a.NotNil(aliGame)

	// bob takes a valid first turn with his first color
	bobGame = (&blockers.TakeTurn{
		PieceIdx: 16,
		Position: 17,
		Rotation: 1,
	}).MustDo(r.Bob().Client())
	a.NotNil(bobGame)

	r.Log().Debug(Sprint("STOP"))
}
