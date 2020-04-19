package main

import (
	"github.com/0xor1/wtf/cmd/pandemic/pkg/game/gameeps"
	"github.com/0xor1/wtf/pkg/store"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/config"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
)

func main() {
	config := config.Get()
	if config.IsLocal {
		defer config.Store.(store.LocalClient).MustDeleteStore()
	}
	app.Run(func(c *app.Config) {
		c.Name = "Pandemic"
		c.Description = "A multiplayer web app version of the popular boardgame pandemic"
		if config.IsLocal {
			c.SessionSecure = false
		}
		c.SessionAuthKey64s = config.SessionAuthKey64s
		c.SessionEncrKey32s = config.SessionEncrKey32s
		c.Log = config.Log
		c.ToolboxMware = service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store)
		c.RateLimiterPool = config.Cache
		c.Endpoints = gameeps.Eps
	})
}
