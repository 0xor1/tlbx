package main

import (
	"net/http"

	"github.com/0xor1/tlbx/cmd/games/pkg/blockers/blockerseps"
	"github.com/0xor1/tlbx/cmd/games/pkg/config"
	"github.com/0xor1/tlbx/cmd/games/pkg/game"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/server"
)

func main() {
	config := config.Get("config.json")
	app.Run(func(c *app.Config) {
		c.StaticDir = config.Web.StaticDir
		c.ContentSecurityPolicies = config.Web.ContentSecurityPolicies
		c.Name = "games"
		c.Description = "a web app to play turn based multiplayer games"
		c.TlbxSetup = app.TlbxMwares{
			session.BasicMware(
				config.Web.Session.AuthKey64s,
				config.Web.Session.EncrKey32s,
				config.Web.Session.Secure),
			ratelimit.MeMware(config.Redis.RateLimit, config.Web.RateLimit),
			service.Mware(config.Redis.Cache, config.SQL.User, config.SQL.Pwd, config.SQL.Data, config.Email, config.Store, config.FCM),
		}
		c.Version = config.Version
		c.Log = config.Log
		c.Endpoints = append(append(c.Endpoints, game.Eps...), blockerseps.Eps...)
		c.Serve = func(h http.HandlerFunc) {
			server.Run(func(c *server.Config) {
				c.AppBindTo = config.Web.AppBindTo
				c.Log = config.Log
				c.Handler = h
			})
		}
	})
}
