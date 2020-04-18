package config

import (
	"encoding/base64"
	"time"

	"github.com/0xor1/wtf/pkg/config"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/email"
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/log"
	"github.com/0xor1/wtf/pkg/store"
)

type Config struct {
	IsLocal           bool
	FromEmail         string
	BaseHref          string
	SessionAuthKey64s [][]byte
	SessionEncrKey32s [][]byte
	Log               log.Log
	Email             email.Client
	Store             store.Client
	Cache             iredis.Pool
	User              isql.ReplicaSet
	Pwd               isql.ReplicaSet
	Data              isql.ReplicaSet
}

func Get(file ...string) *Config {
	res := &Config{}
	c := config.New(file...)
	c.SetDefault("isLocal", true)
	c.SetDefault("fromEmail", "test@test.localhost")
	c.SetDefault("baseHref", "http://localhost:8081")
	c.SetDefault("log.type", "local")
	c.SetDefault("email.type", "local")
	c.SetDefault("store.type", "local")
	c.SetDefault("store.dir", "tmpStoreDir")
	// session cookie store
	c.SetDefault("sessionAuthKey64s", []interface{}{
		"Va3ZMfhH4qSfolDHLU7oPal599DMcL93A80rV2KLM_om_HBFFUbodZKOHAGDYg4LCvjYKaicodNmwLXROKVgcA",
		"WK_2RgRx6vjfWVkpiwOCB1fvv1yklnltstBjYlQGfRsl6LyVV4mkt6UamUylmkwC8MEgb9bSGr1FYgM2Zk20Ug",
	})
	c.SetDefault("sessionEncrKey32s", []interface{}{
		"3ICuYRUelY-4Fhak0Iw0_5CW24bJvxFWM0jAA78IIp8",
		"u80sYkgbBav52fJXbENYhN3Iyof7WhuLHHMaS_rmUQw",
	})
	c.SetDefault("cache", "localhost:6379")
	c.SetDefault("user.primary", "users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/users?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("user.slaves", []string{})
	c.SetDefault("pwd.primary", "pwds:C0-Mm-0n-Pwd5@tcp(localhost:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("pwd.slaves", []string{})
	c.SetDefault("data.primary", "data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("data.slaves", []string{})
	c.SetDefault("sql.connMaxLifetime", 5*time.Second)
	c.SetDefault("sql.maxIdleConns", 100)
	c.SetDefault("sql.maxOpenConns", 100)

	if c.GetBool("isLocal") {
		res.IsLocal = true
		res.Log = log.New()
		res.Email = email.NewLocalClient(res.Log)
		res.Store = store.NewLocalClient(c.GetString("store.dir"))
	} else {
		res.IsLocal = false
		switch c.GetString("log.type") {
		default:
			PanicIf(true, "unsupported log type %s", c.GetString("log.type"))
		}

		switch c.GetString("email.type") {
		case "sparkpost":
			fallthrough // TODO
		default:
			PanicIf(true, "unsupported email type %s", c.GetString("email.type"))
		}

		switch c.GetString("store.type") {
		default:
			PanicIf(true, "unsupported store type %s", c.GetString("store.type"))
		}
	}

	res.FromEmail = c.GetString("fromEmail")
	res.BaseHref = c.GetString("baseHref")

	authKey64s := c.GetStringSlice("sessionAuthKey64s")
	encrKey32s := c.GetStringSlice("sessionEncrKey32s")
	for i := range authKey64s {
		authBytes, err := base64.RawURLEncoding.DecodeString(authKey64s[i])
		PanicOn(err)
		PanicIf(len(authBytes) != 64, "sessionAuthBytes length is not 64")
		res.SessionAuthKey64s = append(res.SessionAuthKey64s, authBytes)
		encrBytes, err := base64.RawURLEncoding.DecodeString(encrKey32s[i])
		PanicOn(err)
		PanicIf(len(encrBytes) != 32, "sessionEncrBytes length is not 32")
		res.SessionEncrKey32s = append(res.SessionEncrKey32s, encrBytes)
	}

	res.Cache = iredis.CreatePool(c.GetString("cache"))

	sqlMaxLifetime := c.GetDuration("sql.connMaxLifetime")
	sqlMaxIdleConns := c.GetInt("sql.maxIdleConns")
	sqlMaxOpenConns := c.GetInt("sql.maxOpenConns")

	var err error
	res.User, err = isql.NewReplicaSet(c.GetString("user.primary"), c.GetStringSlice("user.slaves")...)
	PanicOn(err)
	res.User.SetConnMaxLifetime(sqlMaxLifetime)
	res.User.SetMaxIdleConns(sqlMaxIdleConns)
	res.User.SetMaxOpenConns(sqlMaxOpenConns)

	res.Pwd, err = isql.NewReplicaSet(c.GetString("pwd.primary"), c.GetStringSlice("pwd.slaves")...)
	PanicOn(err)
	res.Pwd.SetConnMaxLifetime(sqlMaxLifetime)
	res.Pwd.SetMaxIdleConns(sqlMaxIdleConns)
	res.Pwd.SetMaxOpenConns(sqlMaxOpenConns)

	res.Data, err = isql.NewReplicaSet(c.GetString("data.primary"), c.GetStringSlice("data.slaves")...)
	PanicOn(err)
	res.Data.SetConnMaxLifetime(sqlMaxLifetime)
	res.Data.SetMaxIdleConns(sqlMaxIdleConns)
	res.Data.SetMaxOpenConns(sqlMaxOpenConns)

	return res
}
