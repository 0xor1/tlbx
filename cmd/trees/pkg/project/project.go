package project

import (
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
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

func (a *Create) Do(c *app.Client) (interface{}, error) {
	res := &struct{}{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) interface{} {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}
