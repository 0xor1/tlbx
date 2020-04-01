package main

import (
	"github.com/0xor1/wtf/cmd/todo/pkg/item/itemeps"
	"github.com/0xor1/wtf/cmd/todo/pkg/list/listeps"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/email"
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/log"
	"github.com/0xor1/wtf/pkg/store"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/auth/autheps"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
)

func main() {
	l := log.New()
	cache := iredis.CreatePool("localhost:6379")
	user, err := isql.NewReplicaSet("users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/users?parseTime=true&loc=UTC&multiStatements=true")
	PanicOn(err)
	pwd, err := isql.NewReplicaSet("pwds:C0-Mm-0n-Pwd5@tcp(localhost:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true")
	PanicOn(err)
	data, err := isql.NewReplicaSet("data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data?parseTime=true&loc=UTC&multiStatements=true")
	PanicOn(err)
	email := email.NewLocalClient(l)
	store := store.NewLocalClient("tmpStoreDir")
	defer store.MustDeleteStore()
	app.Run(func(c *app.Config) {
		c.Name = "Todo"
		c.Description = "A simple Todo list application, create multiple lists with many items which can be marked complete or uncomplete"
		c.Log = l
		c.ToolboxMware = service.Mware(cache, user, pwd, data, email, store)
		c.RateLimiterPool = cache
		c.Endpoints = autheps.New(nil, "test@test.localhost", "http://localhost:8080")
		c.Endpoints = append(c.Endpoints, listeps.Eps...)
		c.Endpoints = append(c.Endpoints, itemeps.Eps...)
	})
}
