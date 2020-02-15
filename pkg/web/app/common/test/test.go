package test

import (
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/email"
	"github.com/0xor1/wtf/pkg/iredis"
	"github.com/0xor1/wtf/pkg/isql"
	"github.com/0xor1/wtf/pkg/log"
	"github.com/0xor1/wtf/pkg/store"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/auth"
	"github.com/0xor1/wtf/pkg/web/app/common/auth/autheps"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
)

var (
	pwd         = "1aA$_t;3"
	emailSuffix = "@test.localhost"
)

type Rig interface {
	Log() log.Log
	Users() Users
	Service() ServiceLayer
	CleanUp()
}

type Users interface {
	Ali() User
	Bob() User
	Cat() User
	Dan() User
}

type User interface {
	Client() *app.Client
	ID() ID
}

type ServiceLayer interface {
	Cache() iredis.Pool
	User() isql.ReplicaSet
	Pwd() isql.ReplicaSet
	Data() isql.ReplicaSet
	Email() email.Client
	Store() store.LocalClient
}

type user struct {
	client *app.Client
	id     ID
}

func (u *user) Client() *app.Client {
	return u.client
}

func (u *user) ID() ID {
	return u.id
}

type rig struct {
	ali   *user
	bob   *user
	cat   *user
	dan   *user
	log   log.Log
	cache iredis.Pool
	user  isql.ReplicaSet
	pwd   isql.ReplicaSet
	data  isql.ReplicaSet
	email email.Client
	store store.LocalClient
}

func (r *rig) Users() Users {
	return r
}

func (r *rig) Service() ServiceLayer {
	return r
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

func (r *rig) Log() log.Log {
	return r.log
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

func (r *rig) Store() store.LocalClient {
	return r.store
}

func New(eps []*app.Endpoint, onDelete func(ID)) Rig {
	l := log.New()
	r := &rig{
		log:   l,
		cache: iredis.CreatePool("localhost:6379"),
		email: email.NewLocalClient(l),
		store: store.NewLocalClient("tmpTestStore"),
	}

	var err error
	r.user, err = isql.NewReplicaSet("users:C0-Mm-0n-U5-3r5@tcp(localhost:3306)/users?parseTime=true")
	PanicOn(err)
	r.pwd, err = isql.NewReplicaSet("pwds:C0-Mm-0n-Pwd5@tcp(localhost:3306)/pwds?parseTime=true")
	PanicOn(err)
	r.data, err = isql.NewReplicaSet("data:C0-Mm-0n-Da-Ta@tcp(localhost:3306)/data?parseTime=true")
	PanicOn(err)

	go app.Run(func(c *app.Config) {
		c.ToolboxMware = service.Mware(r.cache, r.user, r.pwd, r.data, r.email, r.store)
		c.RateLimiterPool = r.cache
		c.Endpoints = append(eps, autheps.New(onDelete, "test@test.localhost", "http://localhost:8080")...)
	})

	time.Sleep(20 * time.Millisecond)
	r.ali = r.createUser("ali"+emailSuffix, pwd)
	r.bob = r.createUser("bob"+emailSuffix, pwd)
	r.cat = r.createUser("cat"+emailSuffix, pwd)
	r.dan = r.createUser("dan"+emailSuffix, pwd)

	return r
}

func (r *rig) CleanUp() {
	r.Service().Store().MustDeleteStore()
	del := &auth.Delete{}
	del.MustDo(r.Ali().Client())
	del.MustDo(r.Bob().Client())
	del.MustDo(r.Cat().Client())
	del.MustDo(r.Dan().Client())
}

func (r *rig) createUser(email, pwd string) *user {
	c := app.NewClient("http://localhost:8080")
	(&auth.Register{
		Email:      email,
		Pwd:        pwd,
		ConfirmPwd: pwd,
	}).MustDo(c)

	(&auth.ResendActivateLink{
		Email: email,
	}).MustDo(c)

	var code string
	row := r.User().Primary().QueryRow(`SELECT activateCode FROM users WHERE email=?`, email)
	PanicOn(row.Scan(&code))

	(&auth.Activate{
		Email: email,
		Code:  code,
	}).MustDo(c)

	id := (&auth.Login{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c).Me

	(&auth.ChangeEmail{
		NewEmail: "change" + emailSuffix,
	}).MustDo(c)

	(&auth.ResendChangeEmailLink{}).MustDo(c)

	row = r.User().Primary().QueryRow(`SELECT changeEmailCode FROM users WHERE id=?`, id)
	PanicOn(row.Scan(&code))

	(&auth.ConfirmChangeEmail{
		Me:   id,
		Code: code,
	}).MustDo(c)

	(&auth.ChangeEmail{
		NewEmail: email,
	}).MustDo(c)

	row = r.User().Primary().QueryRow(`SELECT changeEmailCode FROM users WHERE id=?`, id)
	PanicOn(row.Scan(&code))

	(&auth.ConfirmChangeEmail{
		Me:   id,
		Code: code,
	}).MustDo(c)

	newPwd := pwd + "123abc"
	(&auth.SetPwd{
		CurrentPwd:    pwd,
		NewPwd:        newPwd,
		ConfirmNewPwd: newPwd,
	}).MustDo(c)

	(&auth.Logout{}).MustDo(c)

	(&auth.Login{
		Email: email,
		Pwd:   newPwd,
	}).MustDo(c)

	(&auth.ResetPwd{
		Email: email,
	}).MustDo(c)

	return &user{
		client: c,
		id:     id,
	}
}
