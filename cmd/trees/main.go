package main

import (
	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/item/itemeps"
	"github.com/0xor1/tlbx/cmd/trees/pkg/list/listeps"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
)

func main() {
	config := config.Get()
	eps := []*app.Endpoint{}
	if config.IsLocal {
		store := config.Store.(store.LocalClient)
		defer store.MustDeleteStore()
		eps = append(eps, store.Endpoints()...)
	}
	app.Run(func(c *app.Config) {
		c.StaticDir = config.StaticDir
		c.ContentSecurityPolicies = config.ContentSecurityPolicies
		c.Name = "Todo"
		c.Description = "A simple project management web application which stores tasks in trees"
		c.TlbxMwares = app.TlbxMwares{
			session.BasicMware(config.SessionAuthKey64s, config.SessionEncrKey32s, config.IsLocal),
			me.RateLimitMware(config.Cache),
			service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store),
		}
		c.Log = config.Log
		c.Endpoints = append(append(append(eps, usereps.New(config.FromEmail, config.ActivateFmtLink, config.ConfirmChangeEmailFmtLink, listeps.OnDelete, true, nil, true, nil)...), listeps.Eps...), itemeps.Eps...)
	})
}
