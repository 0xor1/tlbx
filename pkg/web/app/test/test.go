package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/email"
	"github.com/0xor1/tlbx/pkg/fcm"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/log"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
)

const (
	baseHref    = "http://localhost"
	pwd         = "1aA$_t;3"
	emailSuffix = "@test.localhost"
)

type Rig interface {
	// unique
	Unique() string
	// http
	NewClient() *app.Client
	// log
	Log() log.Log
	// users
	Ali() User
	Bob() User
	Cat() User
	Dan() User
	// services
	Cache() iredis.Pool
	User() isql.ReplicaSet
	Pwd() isql.ReplicaSet
	Data() isql.ReplicaSet
	Email() email.Client
	Store() store.Client
	// cleanup
	CleanUp()
}

type User interface {
	Client() *app.Client
	ID() ID
	Email() string
	Pwd() string
}

type testUser struct {
	client *app.Client
	id     ID
	email  string
	pwd    string
}

func (u *testUser) Client() *app.Client {
	return u.client
}

func (u *testUser) ID() ID {
	return u.id
}

func (u *testUser) Email() string {
	return u.email
}

func (u *testUser) Pwd() string {
	return u.pwd
}

type rig struct {
	rootHandler http.HandlerFunc
	unique      string
	ali         *testUser
	bob         *testUser
	cat         *testUser
	dan         *testUser
	log         log.Log
	rateLimit   iredis.Pool
	cache       iredis.Pool
	user        isql.ReplicaSet
	pwd         isql.ReplicaSet
	data        isql.ReplicaSet
	email       email.Client
	store       store.Client
	fcm         fcm.Client
	useAuth     bool
}

func (r *rig) Unique() string {
	return r.unique
}

func (r *rig) Log() log.Log {
	return r.log
}

func (r *rig) Ali() User {
	return r.ali
}

func (r *rig) Bob() User {
	return r.bob
}

func (r *rig) Cat() User {
	return r.cat
}

func (r *rig) Dan() User {
	return r.dan
}

func (r *rig) RateLimit() iredis.Pool {
	return r.rateLimit
}

func (r *rig) Cache() iredis.Pool {
	return r.cache
}

func (r *rig) User() isql.ReplicaSet {
	return r.user
}

func (r *rig) Pwd() isql.ReplicaSet {
	return r.pwd
}

func (r *rig) Data() isql.ReplicaSet {
	return r.data
}

func (r *rig) Email() email.Client {
	return r.email
}

func (r *rig) Store() store.Client {
	return r.store
}

func (r *rig) FCM() fcm.Client {
	return r.fcm
}

func (r *rig) NewClient() *app.Client {
	return app.NewClient(baseHref, r)
}

func (r *rig) Do(req *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	r.rootHandler(rec, req)
	return rec.Result(), nil
}

func NewRig(
	config *config.Config,
	eps []*app.Endpoint,
	useUsers bool,
	onActivate func(app.Tlbx, *user.User),
	onDelete func(app.Tlbx, ID),
	onSetSocials func(app.Tlbx, *user.User) error,
	validateFcmTopic func(app.Tlbx, IDs) (sql.Tx, error),
	buckets ...string,
) Rig {
	r := &rig{
		unique:    Strf("%d", os.Getpid()),
		log:       config.Log,
		rateLimit: config.Redis.RateLimit,
		cache:     config.Redis.Cache,
		email:     config.Email,
		store:     config.Store,
		fcm:       config.FCM,
		user:      config.SQL.User,
		pwd:       config.SQL.Pwd,
		data:      config.SQL.Data,
		useAuth:   useUsers,
	}

	for _, bucket := range buckets {
		r.store.MustCreateBucket(bucket, "private")
	}

	if useUsers {
		r.store.MustCreateBucket(usereps.AvatarBucket, "public_read")
		eps = append(
			eps,
			usereps.New(
				config.App.FromEmail,
				config.App.ActivateFmtLink,
				config.App.ConfirmChangeEmailFmtLink,
				onActivate,
				onDelete,
				onSetSocials,
				validateFcmTopic)...)
	}
	go app.Run(func(c *app.Config) {
		c.TlbxSetup = app.TlbxMwares{
			session.BasicMware(
				config.Web.Session.AuthKey64s,
				config.Web.Session.EncrKey32s,
				config.Web.Session.Secure),
			ratelimit.MeMware(r.rateLimit, 1000000),
			service.Mware(r.cache, r.user, r.pwd, r.data, r.email, r.store, r.fcm),
		}
		c.Endpoints = eps
		c.Serve = func(h http.HandlerFunc) {
			r.rootHandler = h
		}
	})

	// sleep to ensure r.rootHandler has been passed to rig struct
	time.Sleep(20 * time.Millisecond)
	r.ali = r.createUser("ali", emailSuffix, pwd)
	r.bob = r.createUser("bob", emailSuffix, pwd)
	r.cat = r.createUser("cat", emailSuffix, pwd)
	r.dan = r.createUser("dan", emailSuffix, pwd)
	return r
}

func (r *rig) CleanUp() {
	if r.useAuth {
		(&user.Delete{
			Pwd: r.Ali().Pwd(),
		}).MustDo(r.Ali().Client())
		(&user.Delete{
			Pwd: r.Bob().Pwd(),
		}).MustDo(r.Bob().Client())
		(&user.Delete{
			Pwd: r.Cat().Pwd(),
		}).MustDo(r.Cat().Client())
		(&user.Delete{
			Pwd: r.Dan().Pwd(),
		}).MustDo(r.Dan().Client())
	}
}

func (r *rig) createUser(handleSuffix, emailSuffix, pwd string) *testUser {
	email := Strf("%s%s%s", handleSuffix, emailSuffix, r.unique)
	c := r.NewClient()
	if r.useAuth {
		(&user.Register{
			Handle:     ptr.String(Strf("%s%s", handleSuffix, r.unique)),
			Alias:      ptr.String(handleSuffix),
			Email:      email,
			Pwd:        pwd,
			ConfirmPwd: pwd,
		}).MustDo(c)

		var code string
		row := r.User().Primary().QueryRow(`SELECT activateCode FROM users WHERE email=?`, email)
		PanicOn(row.Scan(&code))

		(&user.Activate{
			Email: email,
			Code:  code,
		}).MustDo(c)

		id := (&user.Login{
			Email: email,
			Pwd:   pwd,
		}).MustDo(c).ID

		return &testUser{
			client: c,
			id:     id,
			email:  email,
			pwd:    pwd,
		}
	}
	return &testUser{client: c}
}
