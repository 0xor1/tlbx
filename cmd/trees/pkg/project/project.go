package project

import (
	"time"

	"github.com/0xor1/tlbx/cmd/trees/pkg/task"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

// type Task struct {
// 	ID                   ID        `json:"id"`
// 	Parent               *ID       `json:"parent"`
// 	FirstChild           *ID       `json:"firstChild"`
// 	NextSibling          *ID       `json:"nextSibling"`
// 	User                 *ID       `json:"user"`
// 	Name                 string    `json:"name"`
// 	Description          string    `json:"description"`
// 	CreatedOn            time.Time `json:"createdOn"`
// 	MinimumRemainingTime uint64    `json:"minimumRemainingTime"`
// 	EstimatedTime        uint64    `json:"estimatedTime"`
// 	LoggedTime           uint64    `json:"loggedTime"`
// 	EstimatedSubTime     uint64    `json:"estimatedSubTime"`
// 	LoggedSubTime        uint64    `json:"loggedSubTime"`
// 	FileCount            uint64    `json:"fileCount"`
// 	FileSize             uint64    `json:"fileSize"`
// 	SubFileCount         uint64    `json:"subFileCount"`
// 	SubFileSize          uint64    `json:"subFileSize"`
// 	ChildCount           uint64    `json:"childCount"`
// 	DescendantCount      uint64    `json:"descendantCount"`
// 	IsParallel           bool      `json:"isParallel"`
// }

type Project struct {
	task.Task
	Base
	IsArchived bool `json:"isArchived"`
}

type Base struct {
	HoursPerDay uint8      `json:"hoursPerDay"`
	DaysPerWeek uint8      `json:"daysPerWeek"`
	StartOn     *time.Time `json:"startOn"`
	DueOn       *time.Time `json:"dueOn"`
	IsPublic    bool       `json:"isPublic"`
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
