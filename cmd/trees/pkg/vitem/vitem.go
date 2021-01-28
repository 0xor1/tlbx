package vitem

import (
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/field"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Type string

func (t *Type) Validate() {
	app.BadReqIf(t != nil && !(*t == TypeTime || *t == TypeCost), "invalid type")
}

func (t *Type) String() string {
	return string(*t)
}

func (t *Type) UnmarshalJSON(raw []byte) error {
	val := StrTrim(StrLower(string(raw)), `"`)
	*t = Type(val)
	t.Validate()
	return nil
}

const (
	TypeTime Type = "time"
	TypeCost Type = "cost"
)

type Vitem struct {
	Task      ID        `json:"task"`
	Type      Type      `json:"type"`
	ID        ID        `json:"id"`
	CreatedBy ID        `json:"createdBy"`
	CreatedOn time.Time `json:"createdOn"`
	Inc       uint64    `json:"inc"`
	Note      string    `json:"note"`
}

type Create struct {
	Host    ID      `json:"host"`
	Project ID      `json:"project"`
	Task    ID      `json:"task"`
	Type    Type    `json:"type"`
	Est     *uint64 `json:"est,omitempty"`
	Inc     uint64  `json:"inc"`
	Note    string  `json:"note,omitempty"`
}

type VitemRes struct {
	Task *task.Task `json:"task,omitempty"`
	Item *Vitem     `json:"item,omitempty"`
}

func (_ *Create) Path() string {
	return "/vitem/create"
}

func (a *Create) Do(c *app.Client) (*VitemRes, error) {
	res := &VitemRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *VitemRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Update struct {
	Host    ID            `json:"host"`
	Project ID            `json:"project"`
	Task    ID            `json:"task"`
	Type    Type          `json:"type"`
	ID      ID            `json:"id"`
	Inc     *field.UInt64 `json:"inc,omitempty"`
	Note    *field.String `json:"note,omitempty"`
}

func (_ *Update) Path() string {
	return "/vitem/update"
}

func (a *Update) Do(c *app.Client) (*VitemRes, error) {
	res := &VitemRes{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Update) MustDo(c *app.Client) *VitemRes {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}

type Get struct {
	Host         ID         `json:"host"`
	Project      ID         `json:"project"`
	Type         Type       `json:"type"`
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
	Set  []*Vitem `json:"set"`
	More bool     `json:"more"`
}

func (_ *Get) Path() string {
	return "/vitem/get"
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
	Host    ID   `json:"host"`
	Project ID   `json:"project"`
	Task    ID   `json:"task"`
	Type    Type `json:"type"`
	ID      ID   `json:"id"`
}

func (_ *Delete) Path() string {
	return "/vitem/delete"
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
