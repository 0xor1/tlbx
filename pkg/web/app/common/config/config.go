package config

import (
	"github.com/0xor1/wtf/pkg/config"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/email"
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/log"
	"github.com/0xor1/wtf/pkg/store"
)

type Config struct {
	Log   log.Log
	Email email.Client
	Store store.Client
	Cache iredis.Pool
	User  isql.ReplicaSet
	Pwd   isql.ReplicaSet
	Data  isql.ReplicaSet
}

func Get(file ...string) *Config {
	res := &Config{}
	c := config.New(file...)
	c.SetDefault("log.type", "local")
	c.SetDefault("email.type", "local")
	c.SetDefault("store.type", "local")
	c.SetDefault("store.dir", "tmpStoreDir")
	c.SetDefault("cache", "localhost:6379")
	c.SetDefault("user.primary", "users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/users?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("user.slaves", []string{})
	c.SetDefault("pwd.primary", "pwds:C0-Mm-0n-Pwd5@tcp(localhost:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("pwd.slaves", []string{})
	c.SetDefault("data.primary", "data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("data.slaves", []string{})

	switch c.GetString("log.type") {
	case "local":
		res.Log = log.New()
	default:
		PanicIf(true, "unsupported log type %s", c.GetString("log.type"))
	}

	switch c.GetString("email.type") {
	case "local":
		res.Email = email.NewLocalClient(res.Log)
	case "sparkpost":
		fallthrough // TODO
	default:
		PanicIf(true, "unsupported email type %s", c.GetString("email.type"))
	}

	switch c.GetString("store.type") {
	case "local":
		res.Store = store.NewLocalClient(c.GetString("store.dir"))
	default:
		PanicIf(true, "unsupported store type %s", c.GetString("store.type"))
	}

	res.Cache = iredis.CreatePool(c.GetString("cache"))

	var err error
	res.User, err = isql.NewReplicaSet(c.GetString("user.primary"), c.GetStringSlice("user.slaves")...)
	PanicOn(err)
	res.Pwd, err = isql.NewReplicaSet(c.GetString("pwd.primary"), c.GetStringSlice("pwd.slaves")...)
	PanicOn(err)
	res.Data, err = isql.NewReplicaSet(c.GetString("data.primary"), c.GetStringSlice("data.slaves")...)
	PanicOn(err)

	return res
}
