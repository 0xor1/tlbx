package main

import (
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
	"github.com/0xor1/trees/pkg/comment/commenteps"
	"github.com/0xor1/trees/pkg/config"
	"github.com/0xor1/trees/pkg/file/fileeps"
	"github.com/0xor1/trees/pkg/project/projecteps"
	"github.com/0xor1/trees/pkg/task/taskeps"
	"github.com/0xor1/trees/pkg/vitem/vitemeps"
)

func main() {
	config := config.Get("config.json")
	app.Run(func(c *app.Config) {
		c.StaticDir = config.Web.StaticDir
		c.ContentSecurityPolicies = config.Web.ContentSecurityPolicies
		c.Name = "trees"
		c.Description = "a simple project management app which stores tasks in trees"
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
		c.Endpoints = app.JoinEps(
			usereps.New(
				config.App.FromEmail,
				config.App.ActivateFmtLink,
				config.App.LoginLinkFmtLink,
				config.App.ConfirmChangeEmailFmtLink,
				nil,
				nil,
				projecteps.OnDelete,
				projecteps.OnSetSocials,
				projecteps.ValidateFCMTopic,
				true),
			projecteps.Eps,
			taskeps.Eps,
			vitemeps.Eps,
			fileeps.Eps,
			commenteps.Eps)
	})
}
