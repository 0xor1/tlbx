package main

import (
	"github.com/0xor1/tlbx/cmd/todo/pkg/config"
	"github.com/0xor1/tlbx/cmd/todo/pkg/item/itemeps"
	"github.com/0xor1/tlbx/cmd/todo/pkg/list/listeps"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
)

func main() {
	config := config.Get()
	eps := []*app.Endpoint{}
	app.Run(func(c *app.Config) {
		c.StaticDir = config.StaticDir
		c.ContentSecurityPolicies = config.ContentSecurityPolicies
		c.Name = "Todo"
		c.Description = "A simple Todo list application, create multiple lists with many items which can be marked complete or uncomplete"
		c.TlbxSetup = app.TlbxMwares{
			session.BasicMware(
				config.Session.AuthKey64s,
				config.Session.EncrKey32s,
				config.Session.Secure),
			ratelimit.MeMware(config.Cache),
			service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store),
		}
		c.Log = config.Log
		c.Endpoints = append(
			append(
				append(
					eps,
					usereps.New(
						config.FromEmail,
						config.ActivateFmtLink,
						config.ConfirmChangeEmailFmtLink,
						nil,
						listeps.OnDelete,
						true,
						nil)...),
				listeps.Eps...),
			itemeps.Eps...)
	})
}
