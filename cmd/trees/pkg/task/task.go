package task

import (
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Task struct {
	ID                  ID        `json:"id"`
	Parent              *ID       `json:"parent"`
	FirstChild          *ID       `json:"firstChild"`
	NextSibling         *ID       `json:"nextSibling"`
	User                *ID       `json:"user"`
	Name                string    `json:"name"`
	Description         *string   `json:"description"`
	CreatedBy           ID        `json:"createdBy"`
	CreatedOn           time.Time `json:"createdOn"`
	MinimumTime         uint64    `json:"minimumTime"`
	EstimatedTime       uint64    `json:"estimatedTime"`
	LoggedTime          uint64    `json:"loggedTime"`
	EstimatedSubTime    uint64    `json:"estimatedSubTime"`
	LoggedSubTime       uint64    `json:"loggedSubTime"`
	EstimatedExpense    uint64    `json:"estimatedExpense"`
	LoggedExpense       uint64    `json:"loggedExpense"`
	EstimatedSubExpense uint64    `json:"estimatedSubExpense"`
	LoggedSubExpense    uint64    `json:"loggedSubExpense"`
	FileCount           uint64    `json:"fileCount"`
	FileSize            uint64    `json:"fileSize"`
	FileSubCount        uint64    `json:"fileSubCount"`
	FileSubSize         uint64    `json:"fileSubSize"`
	ChildCount          uint64    `json:"childCount"`
	DescendantCount     uint64    `json:"descendantCount"`
	IsParallel          bool      `json:"isParallel"`
}

type Create struct {
	Host             ID      `json:"host"`
	Project          ID      `json:"project"`
	Parent           ID      `json:"parent"`
	PreviousSibling  *ID     `json:"previousSibling,omitempty"`
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	IsParallel       bool    `json:"isParallel"`
	User             *ID     `json:"user,omitempty"`
	EstimatedTime    uint64  `json:"estimatedTime"`
	EstimatedExpense uint64  `json:"estimatedExpense"`
}

func (_ *Create) Path() string {
	return "/task/create"
}

func (a *Create) Do(c *app.Client) (*Task, error) {
	res := &Task{}
	err := app.Call(c, a.Path(), a, &res)
	return res, err
}

func (a *Create) MustDo(c *app.Client) *Task {
	res, err := a.Do(c)
	PanicOn(err)
	return res
}
