package gameeps

import (
	"github.com/0xor1/wtf/cmd/pandemic/pkg/game"
	"github.com/0xor1/wtf/pkg/web/app"
)

var (
	Eps = []*app.Endpoint{
		{
			Description:  "Create a new game",
			Path:         (&game.Create{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &game.Create{}
			},
			GetExampleArgs: func() interface{} {
				return &game.Create{
					Name: "My Game",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*game.Create)
				return args
			},
		},
	}
)
