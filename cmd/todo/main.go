package main

import (
	"github.com/0xor1/wtf/cmd/todo/pkg/config"
	"github.com/0xor1/wtf/cmd/todo/pkg/item/itemeps"
	"github.com/0xor1/wtf/cmd/todo/pkg/list/listeps"
	"github.com/0xor1/wtf/pkg/store"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/auth/autheps"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
)

func main() {
	config := config.Get()
	if config.IsLocal {
		defer config.Store.(store.LocalClient).MustDeleteStore()
	}
	app.Run(func(c *app.Config) {
		c.Name = "Todo"
		c.Description = "A simple Todo list application, create multiple lists with many items which can be marked complete or uncomplete"
		if config.IsLocal {
			c.SessionSecure = false
		}
		c.SessionAuthKey64s = config.SessionAuthKey64s
		c.SessionEncrKey32s = config.SessionEncrKey32s
		c.Log = config.Log
		c.ToolboxMware = service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store)
		c.RateLimiterPool = config.Cache
		c.Endpoints = autheps.New(nil, config.FromEmail, config.BaseHref)
		c.Endpoints = append(c.Endpoints, listeps.Eps...)
		c.Endpoints = append(c.Endpoints, itemeps.Eps...)
	})
}
