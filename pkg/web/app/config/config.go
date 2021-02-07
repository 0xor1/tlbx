package config

import (
	"context"
	"encoding/base64"
	"time"

	firebase "firebase.google.com/go"
	"github.com/0xor1/tlbx/pkg/config"
	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/email"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/store"
	sp "github.com/SparkPost/gosparkpost"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"google.golang.org/api/option"
)

type Config struct {
	Session struct {
		Secure     bool
		AuthKey64s [][]byte
		EncrKey32s [][]byte
	}
	FromEmail                 string
	ActivateFmtLink           string
	ConfirmChangeEmailFmtLink string
	StaticDir                 string
	ContentSecurityPolicies   []string
	RateLimit                 struct {
		PerMinute int
	}
	Log   log.Log
	Email email.Client
	Store store.Client
	FCM   fcm.Client
	Cache iredis.Pool
	User  isql.ReplicaSet
	Pwd   isql.ReplicaSet
	Data  isql.ReplicaSet
}

func GetBase(file ...string) *config.Config {
	c := config.New(file...)
	// session cookie store
	c.SetDefault("session.secure", false)
	c.SetDefault("session.authKey64s", []interface{}{
		"Va3ZMfhH4qSfolDHLU7oPal599DMcL93A80rV2KLM_om_HBFFUbodZKOHAGDYg4LCvjYKaicodNmwLXROKVgcA",
		"WK_2RgRx6vjfWVkpiwOCB1fvv1yklnltstBjYlQGfRsl6LyVV4mkt6UamUylmkwC8MEgb9bSGr1FYgM2Zk20Ug",
	})
	c.SetDefault("session.encrKey32s", []interface{}{
		"3ICuYRUelY-4Fhak0Iw0_5CW24bJvxFWM0jAA78IIp8",
		"u80sYkgbBav52fJXbENYhN3Iyof7WhuLHHMaS_rmUQw",
	})
	c.SetDefault("aws.region", "local")
	c.SetDefault("aws.s3.endpoint", "http://localhost:9000")
	c.SetDefault("aws.s3.creds.id", "localtest")
	c.SetDefault("aws.s3.creds.secret", "localtest")
	c.SetDefault("fromEmail", "test@test.localhost")
	c.SetDefault("activateFmtLink", "http://localhost:8081/#/activate?email=%s&code=%s")
	c.SetDefault("confirmChangeEmailFmtLink", "http://localhost:8081/#/confirmChangeEmail?me=%s&code=%s")
	c.SetDefault("staticDir", "client/dist")
	c.SetDefault("contentSecurityPolicies", []interface{}{
		"style-src-elem 'self' https://fonts.googleapis.com",
		"font-src 'self' https://fonts.gstatic.com",
	})
	c.SetDefault("log.type", "local")
	c.SetDefault("email.type", "local")
	c.SetDefault("email.apikey", "")
	c.SetDefault("cache", "localhost:6379")
	c.SetDefault("user.primary", "users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/users?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("user.slaves", []string{})
	c.SetDefault("pwd.primary", "pwds:C0-Mm-0n-Pwd5@tcp(localhost:3306)/pwds?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("pwd.slaves", []string{})
	c.SetDefault("data.primary", "data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data?parseTime=true&loc=UTC&multiStatements=true")
	c.SetDefault("data.slaves", []string{})
	c.SetDefault("sql.connMaxLifetime", 5*time.Second)
	c.SetDefault("sql.maxIdleConns", 50)
	c.SetDefault("sql.maxOpenConns", 100)
	c.SetDefault("rateLimit.perMinute", 300)
	c.SetDefault("fcm.serviceAccountKeyFile", "")

	return c
}

func GetProcessed(c *config.Config) *Config {
	res := &Config{}

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
		spClient := &sp.Client{}
		PanicOn(spClient.Init(&sp.Config{
			BaseUrl:    "https://api.eu.sparkpost.com",
			ApiKey:     c.GetString("email.apikey"),
			ApiVersion: 1,
		}))
		res.Email = email.NewSparkPostClient(spClient)
	default:
		PanicIf(true, "unsupported email type %s", c.GetString("email.type"))
	}

	res.Store = store.New(
		s3.New(
			session.New(
				&aws.Config{
					Region:           ptr.String(c.GetString("aws.region")),
					Endpoint:         ptr.String(c.GetString("aws.s3.endpoint")),
					Credentials:      credentials.NewStaticCredentials(c.GetString("aws.s3.creds.id"), c.GetString("aws.s3.creds.secret"), ""),
					DisableSSL:       ptr.Bool(true),
					S3ForcePathStyle: ptr.Bool(true),
				})))

	if c.GetString("fcm.serviceAccountKeyFile") != "" {
		opt := option.WithCredentialsFile(c.GetString("fcm.serviceAccountKeyFile"))
		app, err := firebase.NewApp(context.Background(), nil, opt)
		PanicOn(err)
		client, err := app.Messaging(context.Background())
		PanicOn(err)
		res.FCM = fcm.NewClient(client)
	} else {
		res.FCM = fcm.NewNopClient(res.Log)
	}

	res.StaticDir = c.GetString("staticDir")
	res.ContentSecurityPolicies = c.GetStringSlice("contentSecurityPolicies")
	res.FromEmail = c.GetString("fromEmail")
	res.ActivateFmtLink = c.GetString("activateFmtLink")
	res.ConfirmChangeEmailFmtLink = c.GetString("confirmChangeEmailFmtLink")
	res.RateLimit.PerMinute = c.GetInt("rateLimit.perMinute")

	res.Session.Secure = c.GetBool("session.secure")
	authKey64s := c.GetStringSlice("session.authKey64s")
	encrKey32s := c.GetStringSlice("session.encrKey32s")
	for i := range authKey64s {
		authBytes, err := base64.RawURLEncoding.DecodeString(authKey64s[i])
		PanicOn(err)
		PanicIf(len(authBytes) != 64, "sessionAuthBytes length is not 64")
		res.Session.AuthKey64s = append(res.Session.AuthKey64s, authBytes)
		encrBytes, err := base64.RawURLEncoding.DecodeString(encrKey32s[i])
		PanicOn(err)
		PanicIf(len(encrBytes) != 32, "sessionEncrBytes length is not 32")
		res.Session.EncrKey32s = append(res.Session.EncrKey32s, encrBytes)
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
