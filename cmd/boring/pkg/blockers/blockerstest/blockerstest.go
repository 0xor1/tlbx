package blockerstest

import (
	"testing"

	"github.com/0xor1/wtf/cmd/boring/pkg/blockers"
	"github.com/0xor1/wtf/cmd/boring/pkg/blockers/blockerseps"
	"github.com/0xor1/wtf/cmd/boring/pkg/config"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app/common/test"
	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	a := assert.New(t)
	r := test.NewRig(config.Get(), blockerseps.Eps, false, nil)
	defer r.CleanUp()

	play2PGame(a, r)
	play3PGame(a, r)
	play4PGame(a, r)
}

func play2PGame(a *assert.Assertions, r test.Rig) {
	var err error

	// ali creates a new game
	g := (&blockers.New{}).
		MustDo(r.Ali().Client())
	a.NotNil(g)

	// bob joins ali's game
	g = (&blockers.Join{Game: g.ID}).
		MustDo(r.Bob().Client())
	a.NotNil(g)

	// ali starts the game
	g = (&blockers.Start{}).
		MustDo(r.Ali().Client())
	a.NotNil(g)

	// ali 1:1 - invalid -missing her first corner cell
	g, err = (&blockers.TakeTurn{
		Piece:    0,
		Position: 1,
	}).Do(r.Ali().Client())
	a.Nil(g)
	a.Regexp("first corner constraint not met", err)

	// ali 1:1 - valid
	g = (&blockers.TakeTurn{
		Piece:    0,
		Position: 0,
	}).MustDo(r.Ali().Client())
	a.NotNil(g)

	// bob 1:1 - valid
	g = (&blockers.TakeTurn{
		Piece:    16,
		Position: 17,
		Rotation: 1,
	}).MustDo(r.Bob().Client())
	a.NotNil(g)

	// ali 1:2 - valid
	g = (&blockers.TakeTurn{
		Piece:    10,
		Position: 377,
		Flip:     1,
	}).MustDo(r.Ali().Client())
	a.NotNil(g)

	// bob 1:2 - valid
	g = (&blockers.TakeTurn{
		Piece:    15,
		Position: 360,
		Flip:     1,
	}).MustDo(r.Bob().Client())
	a.NotNil(g)

	// ali 2:1 - invalid - reuse an already placed piece
	g, err = (&blockers.TakeTurn{
		Piece:    0,
		Position: 21,
	}).Do(r.Ali().Client())
	a.Nil(g)
	a.Regexp("invalid piece, that piece has already been used", err)

	// ali 2:1 - invalid - place outside the board boundaries
	g, err = (&blockers.TakeTurn{
		Piece:    1,
		Position: 19,
	}).Do(r.Ali().Client())
	a.Nil(g)
	a.Regexp("piece/position/rotation combination is not contained on the board", err)

	// ali 2:1 - invalid - place on top of existing piece
	g, err = (&blockers.TakeTurn{
		Piece:    1,
		Position: 0,
	}).Do(r.Ali().Client())
	a.Nil(g)
	a.Regexp("cell already occupied", err)

	// ali 2:1 - invalid - faces touching
	g, err = (&blockers.TakeTurn{
		Piece:    1,
		Position: 20,
	}).Do(r.Ali().Client())
	a.Nil(g)
	a.Regexp("face to face constraint not met", err)

	// ali 2:1 - invalid - no touching diagonals
	g, err = (&blockers.TakeTurn{
		Piece:    1,
		Position: 22,
	}).Do(r.Ali().Client())
	a.Nil(g)
	a.Regexp("diagonal touch constraint not met", err)

	// ali 2:1 - valid
	g = (&blockers.TakeTurn{
		Piece:    1,
		Position: 21,
	}).MustDo(r.Ali().Client())
	a.NotNil(g)

	// bob 2:1 - valid
	g = (&blockers.TakeTurn{
		Piece:    20,
		Position: 14,
	}).MustDo(r.Bob().Client())
	a.NotNil(g)

	// bob 2:2 - invalid - not his turn
	g, err = (&blockers.TakeTurn{
		Piece:    20,
		Position: 14,
	}).Do(r.Bob().Client())
	a.Nil(g)
	a.Regexp("it's not your turn", err)

	// ali 2:2 - valid
	g = (&blockers.TakeTurn{
		Piece:    20,
		Position: 354,
	}).MustDo(r.Ali().Client())
	a.NotNil(g)

	// bob 2:2 - valid
	g = (&blockers.TakeTurn{
		Piece:    19,
		Position: 300,
		Rotation: 3,
	}).MustDo(r.Bob().Client())
	a.NotNil(g)

	// ali 3:1 - valid
	g = (&blockers.TakeTurn{
		Piece:    17,
		Position: 3,
		Rotation: 1,
	}).MustDo(r.Ali().Client())
	a.NotNil(g)

	// bob 3:1 - valid
	g = (&blockers.TakeTurn{
		Piece:    19,
		Position: 77,
		Rotation: 1,
	}).MustDo(r.Bob().Client())
	a.NotNil(g)

	// ali 3:2 - valid
	g = (&blockers.TakeTurn{
		Piece:    19,
		Position: 316,
		Flip:     1,
	}).MustDo(r.Ali().Client())
	a.NotNil(g)

	// bob 3:2 - valid
	// g = (&blockers.TakeTurn{
	// 	Piece:    20,
	// 	Position: 316,
	// 	Flip:     1,
	// }).MustDo(r.Bob().Client())
	// a.NotNil(g)

	printGame(g)
}

func play3PGame(a *assert.Assertions, r test.Rig) {

}

func play4PGame(a *assert.Assertions, r test.Rig) {

}

func printGame(g *blockers.Game) {
	Println()
	Println("turn", g.Turn, "state", g.State, "ended", g.PieceSetsEnded)
	Println()
	printPieceSets(g)
	Println()
	printBoard(g)
	Println()
}

func printPieceSets(g *blockers.Game) {
	for j := uint8(0); j < 4; j++ {
		row := ""
		for i := uint8(0); i < blockers.PiecesCount(); i++ {
			color := aurora.White
			switch j {
			case 0:
				color = aurora.Red
			case 1:
				color = aurora.Green
			case 2:
				color = aurora.Blue
			case 3:
				color = aurora.Yellow
			}
			if g.PieceSets[blockers.PiecesCount()*j+i] == 1 {
				row += aurora.Sprintf(color(`■ `))
			} else {
				row += aurora.Sprintf(color(`□ `))
			}
		}
		Println(row)
	}
}

func printBoard(g *blockers.Game) {
	for y := 0; y < 20; y++ {
		rowStart := 20 * y
		row := "%d "
		if rowStart < 100 {
			row += " "
		}
		if rowStart < 10 {
			row += " "
		}
		row = Sprintf(row, rowStart)
		for x := 0; x < 20; x++ {
			color := aurora.White
			switch g.Board[20*y+x] {
			case 0:
				color = aurora.Red
			case 1:
				color = aurora.Green
			case 2:
				color = aurora.Blue
			case 3:
				color = aurora.Yellow
			}
			row += aurora.Sprintf(color(`■ `))
		}
		row += Sprintf(" %d", 20*(y+1)-1)
		Println(row)
	}
}
