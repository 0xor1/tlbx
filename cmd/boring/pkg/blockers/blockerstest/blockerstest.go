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

	// ali attempts her first turn missing her first corner cell
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

	// ali takes a valid first turn with her second color
	aliGame = (&blockers.TakeTurn{
		PieceIdx: 10,
		Position: 377,
		Flip:     1,
	}).MustDo(r.Ali().Client())
	a.NotNil(aliGame)

	// bob takes a valid first turn with his second color
	bobGame = (&blockers.TakeTurn{
		PieceIdx: 15,
		Position: 360,
		Flip:     1,
	}).MustDo(r.Bob().Client())
	a.NotNil(bobGame)

	// ali attempts an invalid second turn with her first color
	// trying to reuse an already placed piece
	aliGame, err = (&blockers.TakeTurn{
		PieceIdx: 0,
		Position: 21,
	}).Do(r.Ali().Client())
	a.Nil(aliGame)
	a.Regexp("invalid pieceIdx, that piece has already been used", err)

	// ali attempts an invalid second turn with her first color
	// trying to place outside the board boundaries
	aliGame, err = (&blockers.TakeTurn{
		PieceIdx: 1,
		Position: 19,
	}).Do(r.Ali().Client())
	a.Nil(aliGame)
	a.Regexp("piece/position/rotation combination is not contained on the board", err)

	// ali attempts an invalid second turn with her first color
	// trying to place on top of her existing piece
	aliGame, err = (&blockers.TakeTurn{
		PieceIdx: 1,
		Position: 0,
	}).Do(r.Ali().Client())
	a.Nil(aliGame)
	a.Regexp("cell already occupied", err)

	// ali attempts an invalid second turn with her first color
	// trying to place face touching pieces
	aliGame, err = (&blockers.TakeTurn{
		PieceIdx: 1,
		Position: 20,
	}).Do(r.Ali().Client())
	a.Nil(aliGame)
	a.Regexp("face to face constraint not met", err)

	// ali attempts an invalid second turn with her first color
	// trying to place without touching diagonals
	aliGame, err = (&blockers.TakeTurn{
		PieceIdx: 1,
		Position: 22,
	}).Do(r.Ali().Client())
	a.Nil(aliGame)
	a.Regexp("diagonal touch constraint not met", err)

	// ali takes a valid second turn with her first color
	aliGame = (&blockers.TakeTurn{
		PieceIdx: 1,
		Position: 21,
	}).MustDo(r.Ali().Client())
	a.NotNil(aliGame)

	// bob takes a valid second turn with his first color
	bobGame = (&blockers.TakeTurn{
		PieceIdx: 20,
		Position: 14,
	}).MustDo(r.Bob().Client())
	a.NotNil(bobGame)

	// bob attempts an invalid turn when it is not his go
	bobGame, err = (&blockers.TakeTurn{
		PieceIdx: 20,
		Position: 14,
	}).Do(r.Bob().Client())
	a.Nil(bobGame)
	a.Regexp("it's not your turn", err)

	// ali takes a valid second turn with her second color
	aliGame = (&blockers.TakeTurn{
		PieceIdx: 20,
		Position: 354,
	}).MustDo(r.Ali().Client())
	a.NotNil(aliGame)

	// bob takes a valid second turn with his second color
	bobGame = (&blockers.TakeTurn{
		PieceIdx: 19,
		Position: 300,
		Rotation: 3,
	}).MustDo(r.Bob().Client())
	a.NotNil(bobGame)

	printGame(bobGame)
}

func printGame(g *blockers.Game) {
	Println()
	Println("turnIdx", g.TurnIdx, "state", g.State, "ended", g.PieceSetsEnded)
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
		row := ""
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
		Println(row)
	}
}
