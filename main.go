package main

import (
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
	"github.com/0xor1/trees/pkg/comment/commenteps"
	"github.com/0xor1/trees/pkg/config"
	"github.com/0xor1/trees/pkg/expense/expenseeps"
	"github.com/0xor1/trees/pkg/file/fileeps"
	"github.com/0xor1/trees/pkg/project/projecteps"
	"github.com/0xor1/trees/pkg/task/taskeps"
	"github.com/0xor1/trees/pkg/time/timeeps"
)

func main() {
	config := config.Get("config.json")
	app.Run(func(c *app.Config) {
		c.StaticDir = config.StaticDir
		c.ContentSecurityPolicies = config.ContentSecurityPolicies
		c.Name = "Trees"
		c.Description = "A simple project management app which stores tasks in trees"
		c.TlbxSetup = app.TlbxMwares{
			session.BasicMware(
				config.Session.AuthKey64s,
				config.Session.EncrKey32s,
				config.Session.Secure),
			ratelimit.MeMware(config.Cache),
			service.Mware(config.Cache, config.User, config.Pwd, config.Data, config.Email, config.Store),
		}
		c.Log = config.Log
		c.Endpoints = app.JoinEps(
			usereps.New(
				config.FromEmail,
				config.ActivateFmtLink,
				config.ConfirmChangeEmailFmtLink,
				nil,
				projecteps.OnDelete,
				true,
				nil),
			projecteps.Eps,
			taskeps.Eps,
			timeeps.Eps,
			expenseeps.Eps,
			fileeps.Eps,
			commenteps.Eps)
	})
}
