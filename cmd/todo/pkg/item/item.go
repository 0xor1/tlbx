package item

import (
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type sort string

const (
	SortName        sort = "name"
	SortCreatedOn   sort = "createdOn"
	SortCompletedOn sort = "completedOn"
)

type Item struct {
	ID          ID         `json:"id"`
	Name        string     `json:"name"`
	CreatedOn   time.Time  `json:"createdOn"`
	CompletedOn *time.Time `json:"completedOn"`
}

type Create struct {
	List ID     `json:"list"`
	Name string `json:"name"`
}

func (_ *Create) Path() string {
	return "/item/create"
}

func (a *Create) Do(c *app.Client) (*Item, error) {
	res := &Item{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *Item {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type One struct {
	List ID `json:"list"`
	ID   ID `json:"id"`
}

func (a *One) Do(c *app.Client) (*Item, error) {
	res, err := (&Get{
		List: a.List,
		IDs:  IDs{a.ID},
	}).Do(c)
	if res != nil && len(res.Set) == 1 {
		return res.Set[0], err
	}
	return nil, err
}

func (a *One) MustDo(c *app.Client) *Item {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	List           ID         `json:"list"`
	IDs            IDs        `json:"ids,omitempty"`
	NamePrefix     *string    `json:"namePrefix,omitempty"`
	CreatedOnMin   *time.Time `json:"createdOnMin,omitempty"`
	CreatedOnMax   *time.Time `json:"createdOnMax,omitempty"`
	Completed      *bool      `json:"completed,omitempty"`
	CompletedOnMin *time.Time `json:"completedOnMin,omitempty"`
	CompletedOnMax *time.Time `json:"completedOnMax,omitempty"`
	After          *ID        `json:"after,omitempty"`
	Sort           sort       `json:"sort,omitempty"`
	Asc            *bool      `json:"asc,omitempty"`
	Limit          uint16     `json:"limit,omitempty"`
}

type GetRes struct {
	Set  []*Item `json:"set"`
	More bool    `json:"more"`
}

func (_ *Get) Path() string {
	return "/item/get"
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

type Update struct {
	List     ID            `json:"list"`
	ID       ID            `json:"id"`
	Name     *field.String `json:"name"`
	Complete *field.Bool   `json:"complete"`
}

func (_ *Update) Path() string {
	return "/item/update"
}

func (a *Update) Do(c *app.Client) (*Item, error) {
	res := &Item{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Update) MustDo(c *app.Client) *Item {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Delete struct {
	List ID  `json:"list"`
	IDs  IDs `json:"ids"`
}

func (_ *Delete) Path() string {
	return "/item/delete"
}

func (a *Delete) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Delete) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}
