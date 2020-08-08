package main

import (
	"github.com/0xor1/tlbx/cmd/trees/pkg/config"
	"github.com/0xor1/tlbx/cmd/trees/pkg/project/projecteps"
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
		c.Name = "Trees"
		c.Description = "A simple project management app which stores tasks in trees"
		c.TlbxMwares = app.TlbxMwares{
			session.BasicMware(config.SessionAuthKey64s, config.SessionEncrKey32s, config.IsLocal),
			ratelimit.MeMware(config.Cache),
			service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store),
		}
		c.Log = config.Log
		c.Endpoints = append(
			append(
				eps,
				usereps.New(
					config.FromEmail,
					config.ActivateFmtLink,
					config.ConfirmChangeEmailFmtLink,
					projecteps.OnActivate,
					projecteps.OnDelete,
					true,
					nil,
					true,
					config.AvatarBucket,
					config.AvatarPrefix,
					nil)...),
			projecteps.Eps...)
	})
}
