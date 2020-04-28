package blockers

import (
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/field"
	"github.com/0xor1/wtf/pkg/web/app"
)

type Game struct {
	ID        ID        `json:"id"`
	CreatedOn time.Time `json:"createdOn"`
	UpdatedOn time.Time `json:"updatedOn"`
}

type Create struct {
	Name string `json:"name"`
}

func (_ *Create) Path() string {
	return "/game/create"
}

func (a *Create) Do(c *app.Client) (*Game, error) {
	res := &Game{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *Game {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	ID ID `json:"id"`
}

func (_ *Get) Path() string {
	return "/game/get"
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

type TakeTurn struct {
	ID   ID           `json:"id"`
	Name field.String `json:"name"`
}

func (_ *TakeTurn) Path() string {
	return "/game/takeTurn"
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
	return "/game/delete"
}

func (a *Delete) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Delete) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}
