package comment

import (
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Comment struct {
	Task      ID        `json:"task"`
	ID        ID        `json:"id"`
	CreatedBy ID        `json:"createdBy"`
	CreatedOn time.Time `json:"createdOn"`
	Body      string    `json:"body"`
}

type Create struct {
	Host    ID     `json:"host"`
	Project ID     `json:"project"`
	Task    ID     `json:"task"`
	Body    string `json:"Body"`
}

func (_ *Create) Path() string {
	return "/comment/create"
}

func (a *Create) Do(c *app.Client) (*Comment, error) {
	res := &Comment{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *Comment {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Update struct {
	Host    ID     `json:"host"`
	Project ID     `json:"project"`
	Task    ID     `json:"task"`
	ID      ID     `json:"id"`
	Body    string `json:"body"`
}

func (_ *Update) Path() string {
	return "/comment/update"
}

func (a *Update) Do(c *app.Client) (*Comment, error) {
	res := &Comment{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Update) MustDo(c *app.Client) *Comment {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	Host    ID     `json:"host"`
	Project ID     `json:"project"`
	Task    *ID    `json:"task,omitempty"`
	After   *ID    `json:"after,omitempty"`
	Limit   uint16 `json:"limit,omitempty"`
}

type GetRes struct {
	Set  []*Comment `json:"set"`
	More bool       `json:"more"`
}

func (_ *Get) Path() string {
	return "/comment/get"
}

func (a *Get) Do(c *app.Client) (*GetRes, error) {
	res := &GetRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) *GetRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Delete struct {
	Host    ID `json:"host"`
	Project ID `json:"project"`
	Task    ID `json:"task"`
	ID      ID `json:"id"`
}

func (_ *Delete) Path() string {
	return "/comment/delete"
}

func (a *Delete) Do(c *app.Client) error {
	err := app.Call(c, a.Path(), a, nil)
	return err
}

func (a *Delete) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}
