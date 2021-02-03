package main

import (
	"github.com/0xor1/tlbx/cmd/games/pkg/blockers/blockerseps"
	"github.com/0xor1/tlbx/cmd/games/pkg/config"
	"github.com/0xor1/tlbx/cmd/games/pkg/game"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
)

func main() {
	config := config.Get()
	app.Run(func(c *app.Config) {
		c.StaticDir = config.StaticDir
		c.ContentSecurityPolicies = config.ContentSecurityPolicies
		c.Name = "games"
		c.Description = "a web app to play turn based multiplayer games"
		c.TlbxSetup = app.TlbxMwares{
			session.BasicMware(
				config.Session.AuthKey64s,
				config.Session.EncrKey32s,
				config.Session.Secure),
			ratelimit.MeMware(config.Cache),
			service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store, config.FCM),
		}
		c.Log = config.Log
		c.Endpoints = append(append(c.Endpoints, game.Eps...), blockerseps.Eps...)
	})
}
