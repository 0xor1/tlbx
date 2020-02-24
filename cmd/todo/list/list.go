package list

import (
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
)

type Create struct {
	Name string `json:"name"`
}

type CreateRes struct {
	ID   ID     `json:"id"`
	Name string `json:"name"`
}

func (_ *Create) Path() string {
	return "/list/create"
}

func (a *Create) Do(c *app.Client) error {
	return app.Call(c, a.Path(), a, nil)
}

func (a *Create) MustDo(c *app.Client) {
	PanicOn(a.Do(c))
}
