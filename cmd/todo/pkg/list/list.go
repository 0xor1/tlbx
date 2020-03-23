package list

import (
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/field"
	"github.com/0xor1/wtf/pkg/web/app"
)

type sort string

const (
	SortName               sort = "name"
	SortCreatedOn          sort = "createdOn"
	SortTodoItemCount      sort = "todoItemCount"
	SortCompletedItemCount sort = "completedItemCount"
)

type List struct {
	ID                 ID        `json:"id"`
	Name               string    `json:"name"`
	CreatedOn          time.Time `json:"createdOn"`
	TodoItemCount      int       `json:"todoItemCount"`
	CompletedItemCount int       `json:"completedItemCount"`
}

type Create struct {
	Name string `json:"name"`
}

func (_ *Create) Path() string {
	return "/list/create"
}

func (a *Create) Do(c *app.Client) (*List, error) {
	res := &List{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *List {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	ID ID `json:"id"`
}

func (_ *Get) Path() string {
	return "/list/get"
}

func (a *Get) Do(c *app.Client) (*List, error) {
	res := &List{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Get) MustDo(c *app.Client) *List {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type GetSet struct {
	IDs                   IDs        `json:"ids,omitempty"`
	NameStartsWith        *string    `json:"nameStartsWith,omitempty"`
	CreatedOnMin          *time.Time `json:"createdOnAfter,omitempty"`
	CreatedOnMax          *time.Time `json:"createdOnBefore,omitempty"`
	TodoItemCountMin      *int       `json:"todoItemCountMin,omitempty"`
	TodoItemCountMax      *int       `json:"todoItemCountMax,omitempty"`
	CompletedItemCountMin *int       `json:"completedItemCountMin,omitempty"`
	CompletedItemCountMax *int       `json:"completedItemCountMax,omitempty"`
	After                 *ID        `json:"after,omitempty"`
	Sort                  sort       `json:"sort,omitempty"`
	Asc                   *bool      `json:"asc,omitempty"`
	Limit                 *int       `json:"limit,omitempty"`
}

type GetSetRes struct {
	Set  []*List `json:"set"`
	More bool    `json:"bool"`
}

func (_ *GetSet) Path() string {
	return "/list/getSet"
}

func (a *GetSet) Do(c *app.Client) (*GetSetRes, error) {
	res := &GetSetRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *GetSet) MustDo(c *app.Client) *GetSetRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Update struct {
	ID   ID           `json:"id"`
	Name field.String `json:"name"`
}

func (_ *Update) Path() string {
	return "/list/update"
}

func (a *Update) Do(c *app.Client) (*List, error) {
	res := &List{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Update) MustDo(c *app.Client) *List {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Delete struct {
	IDs IDs `json:"ids"`
}

func (_ *Delete) Path() string {
	return "/list/delete"
}

func (a *Delete) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Delete) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}
