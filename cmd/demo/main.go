package main

import (
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/email"
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/log"
	"github.com/0xor1/wtf/pkg/store"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/auth"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
)

func main() {
	l := log.New()
	cache := iredis.CreatePool("localhost:6379")
	user, err := isql.NewReplicaSet("users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/users?parseTime=true")
	PanicOn(err)
	pwd, err := isql.NewReplicaSet("pwds:C0-Mm-0n-Pwd5@tcp(localhost:3306)/pwds?parseTime=true")
	PanicOn(err)
	data, err := isql.NewReplicaSet("data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data?parseTime=true")
	PanicOn(err)
	email := email.NewLocalClient(l)
	store := store.NewLocalClient(".")
	app.Run(func(c *app.Config) {
		c.Log = l
		c.ToolboxMware = service.Mware(cache, user, pwd, data, email, store)
		c.RateLimiterPool = cache
		c.Endpoints = auth.Endpoints(nil, nil, "local@host.test", "http://localhost:8080", 64, 32768, 8, 1, 32)
	})
}
