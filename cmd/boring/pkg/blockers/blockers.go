package blockers

import (
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
)

func Pieces() []*piece {
	return []*piece{
		// #
		{BoundingBox: []uint8{1, 1}, Shape: []uint8{1}},

		// ##
		{BoundingBox: []uint8{2, 1}, Shape: []uint8{1, 1}},

		// ###
		{BoundingBox: []uint8{3, 1}, Shape: []uint8{1, 1, 1}},

		// #
		// ##
		{BoundingBox: []uint8{2, 2}, Shape: []uint8{1, 0, 1, 1}},

		// ####
		{BoundingBox: []uint8{4, 1}, Shape: []uint8{1, 1, 1, 1}},

		// ##
		// ##
		{BoundingBox: []uint8{2, 2}, Shape: []uint8{1, 1, 1, 1}},

		//  #
		// ###
		{BoundingBox: []uint8{3, 2}, Shape: []uint8{0, 1, 0, 1, 1, 1}},

		//   #
		// ###
		{BoundingBox: []uint8{3, 2}, Shape: []uint8{0, 0, 1, 1, 1, 1}},

		//  ##
		// ##
		{BoundingBox: []uint8{3, 2}, Shape: []uint8{0, 1, 1, 1, 1, 0}},

		// #####
		{BoundingBox: []uint8{5, 1}, Shape: []uint8{1, 1, 1, 1, 1}},

		// ###
		// ##
		{BoundingBox: []uint8{3, 2}, Shape: []uint8{1, 1, 1, 1, 1, 0}},

		//  #
		// ###
		//  #
		{BoundingBox: []uint8{3, 3}, Shape: []uint8{0, 1, 0, 1, 1, 1, 0, 1, 0}},

		// #
		// ###
		//   #
		{BoundingBox: []uint8{3, 3}, Shape: []uint8{1, 0, 0, 1, 1, 1, 0, 0, 1}},

		//    #
		// ####
		{BoundingBox: []uint8{4, 2}, Shape: []uint8{0, 0, 0, 1, 1, 1, 1, 1}},

		//   #
		// ####
		{BoundingBox: []uint8{4, 2}, Shape: []uint8{0, 0, 1, 0, 1, 1, 1, 1}},

		// ###
		//   ##
		{BoundingBox: []uint8{4, 2}, Shape: []uint8{1, 1, 1, 0, 0, 0, 1, 1}},

		// #
		// ###
		//  #
		{BoundingBox: []uint8{3, 3}, Shape: []uint8{1, 0, 0, 1, 1, 1, 0, 1, 0}},

		// ###
		// # #
		{BoundingBox: []uint8{3, 2}, Shape: []uint8{1, 1, 1, 1, 0, 1}},

		// #
		// ###
		// #
		{BoundingBox: []uint8{3, 3}, Shape: []uint8{1, 0, 0, 1, 1, 1, 1, 0, 0}},

		// ##
		//  ##
		//   #
		{BoundingBox: []uint8{3, 3}, Shape: []uint8{1, 1, 0, 0, 1, 1, 0, 0, 1}},

		// #
		// #
		// ###
		{BoundingBox: []uint8{3, 3}, Shape: []uint8{1, 0, 0, 1, 0, 0, 1, 1, 1}},
	}
}

type piece struct {
	BoundingBox []uint8 `json:"bb"`
	Shape       []uint8 `json:"shape"`
}

type Transformation struct {
	Rotation int   `json:"rotation"`
	Flip     uint8 `json:"flip"`
}

type Game struct {
	ID            ID                `json:"id"`
	CreatedOn     time.Time         `json:"createdOn"`
	UpdatedOn     time.Time         `json:"updatedOn"`
	Started       bool              `json:"started"`
	Players       []ID              `json:"players"`
	PieceSetsIdxs []map[uint8]uint8 `json:"pieceSetsIdxs"`
	TurnIdx       uint8             `json:"turnIdx"`
	Board         []uint8           `json:"board"`
}

type New struct{}

func (_ *New) Path() string {
	return "/blockers/new"
}

func (a *New) Do(c *app.Client) (*Game, error) {
	res := &Game{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *New) MustDo(c *app.Client) *Game {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	ID ID `json:"id"`
}

func (_ *Get) Path() string {
	return "/blockers/get"
}

func (a *Get) Do(c *app.Client) (*Game, error) {
	res := &Game{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) *Game {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Start struct{}

func (_ *Start) Path() string {
	return "/blockers/start"
}

func (a *Start) Do(c *app.Client) (*Game, error) {
	res := &Game{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Start) MustDo(c *app.Client) *Game {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type TakeTurn struct {
	Pass           bool            `json:"pass"`
	PieceIdx       uint8           `json:"pieceIdx"`
	Position       uint8           `json:"position"`
	Transformation *Transformation `json:"transformation"`
}

func (_ *TakeTurn) Path() string {
	return "/blockers/takeTurn"
}

func (a *TakeTurn) Do(c *app.Client) (*Game, error) {
	res := &Game{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *TakeTurn) MustDo(c *app.Client) *Game {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Delete struct {
	ID ID `json:"id"`
}

func (_ *Delete) Path() string {
	return "/blockers/delete"
}

func (a *Delete) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Delete) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}
