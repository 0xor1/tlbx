package project

import (
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/consts"
	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Project struct {
	task.Task
	Base
	IsArchived bool `json:"isArchived"`
}

type Base struct {
	CurrencyCode string     `json:"currencyCode"`
	HoursPerDay  uint8      `json:"hoursPerDay"`
	DaysPerWeek  uint8      `json:"daysPerWeek"`
	StartOn      *time.Time `json:"startOn"`
	DueOn        *time.Time `json:"dueOn"`
	IsPublic     bool       `json:"isPublic"`
}

type Create struct {
	Name string `json:"name"`
	Base
}

func (_ *Create) Path() string {
	return "/project/create"
}

func (a *Create) Do(c *app.Client) (*Project, error) {
	res := &Project{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *Project {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type One struct {
	ID ID `json:"id"`
}

func (a *One) Do(c *app.Client) (*Project, error) {
	res, err := (&Get{IDs: IDs{a.ID}}).Do(c)
	if res != nil && len(res.Set) == 1 {
		return res.Set[0], err
	}
	return nil, err
}

func (a *One) MustDo(c *app.Client) *Project {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	IDs                   IDs          `json:"ids,omitempty"`
	NameStartsWith        *string      `json:"nameStartsWith,omitempty"`
	IsArchived            bool         `json:"isArchived"`
	CreatedOnMin          *time.Time   `json:"createdOnMin,omitempty"`
	CreatedOnMax          *time.Time   `json:"createdOnMax,omitempty"`
	StartOnMin            *time.Time   `json:"startOnMin,omitempty"`
	StartOnMax            *time.Time   `json:"startOnMax,omitempty"`
	DueOnMin              *time.Time   `json:"dueOnMin,omitempty"`
	DueOnMax              *time.Time   `json:"dueOnMax,omitempty"`
	TodoItemCountMin      *int         `json:"todoItemCountMin,omitempty"`
	TodoItemCountMax      *int         `json:"todoItemCountMax,omitempty"`
	CompletedItemCountMin *int         `json:"completedItemCountMin,omitempty"`
	CompletedItemCountMax *int         `json:"completedItemCountMax,omitempty"`
	After                 *ID          `json:"after,omitempty"`
	Sort                  *consts.Sort `json:"sort,omitempty"`
	Asc                   *bool        `json:"asc,omitempty"`
	Limit                 *int         `json:"limit,omitempty"`
}

type GetRes struct {
	Set  []*Project `json:"set"`
	More bool       `json:"more"`
}

func (_ *Get) Path() string {
	return "/list/get"
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
	ID   ID           `json:"id"`
	Name field.String `json:"name"`
}

func (_ *Update) Path() string {
	return "/list/update"
}

func (a *Update) Do(c *app.Client) (*Project, error) {
	res := &Project{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Update) MustDo(c *app.Client) *Project {
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

type Activity struct {
	OccurredOn         time.Time     `json:"occurredOn"`
	User               ID            `json:"user"`
	Item               ID            `json:"item"`
	ItemType           consts.Type   `json:"itemType"`
	ItemHasBeenDeleted bool          `json:"itemHasBeenDeleted"`
	Action             consts.Action `json:"action"`
	ItemName           *string       `json:"itemName,omitempty"`
	ExtraInfo          *string       `json:"extraInfo,omitempty"`
}
