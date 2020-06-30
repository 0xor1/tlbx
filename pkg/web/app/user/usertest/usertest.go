package usertest

import (
	"regexp"
	"testing"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	r := test.NewRig(config.GetProcessed(config.GetBase()), nil, true, func(tlbx app.Tlbx, id ID, alias *string) error { return nil }, func(tlbx app.Tlbx, id ID) {})
	defer r.CleanUp()

	a := assert.New(t)
	c := test.NewClient()
	email := "test@test.localhost"
	pwd := "1aA$_t;3"

	(&user.Register{
		Email:      email,
		Pwd:        pwd,
		ConfirmPwd: pwd,
	}).MustDo(c)

	// check existing email err
	err := (&user.Register{
		Email:      email,
		Pwd:        pwd,
		ConfirmPwd: pwd,
	}).Do(c)
	a.Equal(&app.ErrMsg{Status: 400, Msg: "email already registered"}, err)

	(&user.ResendActivateLink{
		Email: email,
	}).MustDo(c)

	var code string
	row := r.User().Primary().QueryRow(`SELECT activateCode FROM users WHERE email=?`, email)
	PanicOn(row.Scan(&code))

	(&user.Activate{
		Email: email,
		Code:  code,
	}).MustDo(c)

	// check return ealry path
	(&user.ResendActivateLink{
		Email: email,
	}).MustDo(c)

	id := (&user.Login{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c).ID

	(&user.ChangeEmail{
		NewEmail: "change@test.localhost",
	}).MustDo(c)

	(&user.ResendChangeEmailLink{}).MustDo(c)

	row = r.User().Primary().QueryRow(`SELECT changeEmailCode FROM users WHERE id=?`, id)
	PanicOn(row.Scan(&code))

	(&user.ConfirmChangeEmail{
		Me:   id,
		Code: code,
	}).MustDo(c)

	(&user.ChangeEmail{
		NewEmail: email,
	}).MustDo(c)

	row = r.User().Primary().QueryRow(`SELECT changeEmailCode FROM users WHERE id=?`, id)
	PanicOn(row.Scan(&code))

	(&user.ConfirmChangeEmail{
		Me:   id,
		Code: code,
	}).MustDo(c)

	newPwd := pwd + "123abc"
	(&user.SetPwd{
		CurrentPwd:    pwd,
		NewPwd:        newPwd,
		ConfirmNewPwd: newPwd,
	}).MustDo(c)

	(&user.Logout{}).MustDo(c)

	(&user.Login{
		Email: email,
		Pwd:   newPwd,
	}).MustDo(c)

	alias := "shabba!"
	(&user.SetAlias{
		Alias: ptr.String(alias),
	}).MustDo(c)

	me := (&user.Me{}).MustDo(c)
	a.Equal(alias, *me.Alias)

	users := (&user.Get{
		Users: []ID{
			id,
			r.Ali().ID(),
			r.Bob().ID(),
			r.Cat().ID(),
			r.Dan().ID(),
		},
	}).MustDo(c)
	a.Equal(5, len(users))

	(&user.Delete{
		Pwd: newPwd,
	}).MustDo(c)

	(&user.Register{
		Email:      email,
		Pwd:        pwd,
		ConfirmPwd: pwd,
	}).MustDo(c)

	row = r.User().Primary().QueryRow(`SELECT activateCode FROM users WHERE email=?`, email)
	PanicOn(row.Scan(&code))

	(&user.Activate{
		Email: email,
		Code:  code,
	}).MustDo(c)

	id = (&user.Login{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c).ID
	a.Equal(id, (&user.Me{}).MustDo(c).ID)

	(&user.ResetPwd{
		Email: email,
	}).MustDo(c)

	err = (&user.ResetPwd{
		Email: email,
	}).Do(c)
	a.Equal(400, err.(*app.ErrMsg).Status)
	a.True(regexp.MustCompile(`must wait [1-9][0-9]{2} seconds before reseting pwd again`).MatchString(err.(*app.ErrMsg).Msg))

	_, err = r.User().Primary().Exec(`DELETE FROM users WHERE id=?`, id)
	PanicOn(err)
	_, err = r.Pwd().Primary().Exec(`DELETE FROM pwds WHERE id=?`, id)
	PanicOn(err)
}
