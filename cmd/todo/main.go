package main

import (
	"github.com/0xor1/wtf/cmd/todo/pkg/item/itemeps"
	"github.com/0xor1/wtf/cmd/todo/pkg/list/listeps"
	"github.com/0xor1/wtf/pkg/store"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/auth/autheps"
	"github.com/0xor1/wtf/pkg/web/app/common/config"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
)

func main() {
	config := config.Get()
	if localStore, ok := config.Store.(store.LocalClient); ok {
		defer localStore.MustDeleteStore()
	}
	app.Run(func(c *app.Config) {
		c.Name = "Todo"
		c.Description = "A simple Todo list application, create multiple lists with many items which can be marked complete or uncomplete"
		c.Log = config.Log
		c.ToolboxMware = service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store)
		c.RateLimiterPool = config.Cache
		c.Endpoints = autheps.New(nil, "test@test.localhost", "http://localhost:8080")
		c.Endpoints = append(c.Endpoints, listeps.Eps...)
		c.Endpoints = append(c.Endpoints, itemeps.Eps...)
	})
}
