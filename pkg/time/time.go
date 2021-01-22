package time

import (
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/trees/pkg/task"
)

type Time struct {
	Task      ID        `json:"task"`
	ID        ID        `json:"id"`
	CreatedBy ID        `json:"createdBy"`
	CreatedOn time.Time `json:"createdOn"`
	Value     uint64    `json:"value"`
	Note      string    `json:"note"`
}

type Create struct {
	Host    ID      `json:"host"`
	Project ID      `json:"project"`
	Task    ID      `json:"task"`
	TimeEst *uint64 `json:"timeEst,omitempty"`
	Value   uint64  `json:"value"`
	Note    string  `json:"note,omitempty"`
}

type TimeRes struct {
	Task *task.Task `json:"task,omitempty"`
	Time *Time      `json:"time,omitempty"`
}

func (_ *Create) Path() string {
	return "/time/create"
}

func (a *Create) Do(c *app.Client) (*TimeRes, error) {
	res := &TimeRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *TimeRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Update struct {
	Host    ID            `json:"host"`
	Project ID            `json:"project"`
	Task    ID            `json:"task"`
	ID      ID            `json:"id"`
	Value   *field.UInt64 `json:"value,omitempty"`
	Note    *field.String `json:"note,omitempty"`
}

func (_ *Update) Path() string {
	return "/time/update"
}

func (a *Update) Do(c *app.Client) (*TimeRes, error) {
	res := &TimeRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Update) MustDo(c *app.Client) *TimeRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	Host         ID         `json:"host"`
	Project      ID         `json:"project"`
	Task         *ID        `json:"task,omitempty"`
	IDs          IDs        `json:"ids,omitempty"`
	CreatedOnMin *time.Time `json:"createdOnMin,omitempty"`
	CreatedOnMax *time.Time `json:"createdOnMax,omitempty"`
	CreatedBy    *ID        `json:"createdBy,omitempty"`
	After        *ID        `json:"after,omitempty"`
	Asc          *bool      `json:"asc,omitempty"`
	Limit        uint16     `json:"limit,omitempty"`
}

type GetRes struct {
	Set  []*Time `json:"set"`
	More bool    `json:"more"`
}

func (_ *Get) Path() string {
	return "/time/get"
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
	return "/time/delete"
}

func (a *Delete) Do(c *app.Client) (*task.Task, error) {
	res := &task.Task{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Delete) MustDo(c *app.Client) *task.Task {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}
