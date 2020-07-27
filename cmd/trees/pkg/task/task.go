package task

import (
	"github.com/0xor1/tlbx/cmd/trees/pkg/project"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Task project.Task

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
