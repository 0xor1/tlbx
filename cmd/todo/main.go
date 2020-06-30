package main

import (
	"github.com/0xor1/tlbx/cmd/todo/pkg/config"
	"github.com/0xor1/tlbx/cmd/todo/pkg/item/itemeps"
	"github.com/0xor1/tlbx/cmd/todo/pkg/list/listeps"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
	"github.com/tomasen/realip"
)

func main() {
	config := config.Get()
	if config.IsLocal {
		defer config.Store.(store.LocalClient).MustDeleteStore()
	}
	app.Run(func(c *app.Config) {
		c.StaticDir = config.StaticDir
		c.ContentSecurityPolicies = config.ContentSecurityPolicies
		c.Name = "Todo"
		c.Description = "A simple Todo list application, create multiple lists with many items which can be marked complete or uncomplete"
		c.TlbxMwares = app.TlbxMwares{
			session.Mware(func(c *session.Config) {
				c.AuthKey64s = config.SessionAuthKey64s
				c.EncrKey32s = config.SessionEncrKey32s
				if config.IsLocal {
					c.Secure = false
				}
			}),
			ratelimit.Mware(func(c *ratelimit.Config) {
				c.KeyGen = func(tlbx app.Tlbx) string {
					var key string
					if me.Exists(tlbx) {
						key = me.Get(tlbx).String()
					}
					return Sprintf("rate-limiter-%s-%s", realip.RealIP(tlbx.Req()), key)
				}
				c.Pool = config.Cache
			}),
			service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store),
		}
		c.Log = config.Log
		c.Endpoints = append(append(usereps.New(nil, nil, config.FromEmail, config.ActivateFmtLink, config.ConfirmChangeEmailFmtLink), listeps.Eps...), itemeps.Eps...)
	})
}
