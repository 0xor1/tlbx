package main

import (
	"github.com/0xor1/wtf/cmd/games/pkg/blockers/blockerseps"
	"github.com/0xor1/wtf/cmd/games/pkg/config"
	"github.com/0xor1/wtf/cmd/games/pkg/game"
	"github.com/0xor1/wtf/pkg/store"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/service"
)

func main() {
	config := config.Get()
	if config.IsLocal {
		defer config.Store.(store.LocalClient).MustDeleteStore()
	}
	app.Run(func(c *app.Config) {
		c.StaticDir = config.StaticDir
		c.ContentSecurityPolicies = config.ContentSecurityPolicies
		c.Name = "games"
		c.Description = "a web app to play turn based multiplayer games"
		if config.IsLocal {
			c.SessionSecure = false
		}
		c.SessionAuthKey64s = config.SessionAuthKey64s
		c.SessionEncrKey32s = config.SessionEncrKey32s
		c.Log = config.Log
		c.ToolboxMware = service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store)
		c.RateLimiterPool = config.Cache
		c.Endpoints = append(append(c.Endpoints, game.Eps...), blockerseps.Eps...)
	})
}
