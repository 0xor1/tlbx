package task

import (
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Task struct {
	ID                   ID        `json:"id"`
	Parent               *ID       `json:"parent"`
	FirstChild           *ID       `json:"firstChild"`
	NextSibling          *ID       `json:"nextSibling"`
	User                 *ID       `json:"user"`
	Name                 string    `json:"name"`
	Description          string    `json:"description"`
	CreatedOn            time.Time `json:"createdOn"`
	MinimumRemainingTime uint64    `json:"minimumRemainingTime"`
	EstimatedTime        uint64    `json:"estimatedTime"`
	LoggedTime           uint64    `json:"loggedTime"`
	EstimatedSubTime     uint64    `json:"estimatedSubTime"`
	LoggedSubTime        uint64    `json:"loggedSubTime"`
	FileCount            uint64    `json:"fileCount"`
	FileSize             uint64    `json:"fileSize"`
	SubFileCount         uint64    `json:"subFileCount"`
	SubFileSize          uint64    `json:"subFileSize"`
	ChildCount           uint64    `json:"childCount"`
	DescendantCount      uint64    `json:"descendantCount"`
	IsParallel           bool      `json:"isParallel"`
}

type Create struct {
}

func (_ *Create) Path() string {
	return "/task/create"
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
