package account

import (
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
)

type Create struct {
	Name string `json:"name"`
}

func (_ *Create) Path() string {
	return "/account/create"
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
