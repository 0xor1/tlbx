package authtest

import (
	"regexp"
	"testing"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/auth"
	"github.com/0xor1/wtf/pkg/web/app/common/config"
	"github.com/0xor1/wtf/pkg/web/app/common/test"
	"github.com/stretchr/testify/assert"
)

func Everything(t *testing.T) {
	r := test.NewRig(config.GetProcessed(config.GetBase()), nil, func(tlbx app.Toolbox, id ID) {})
	defer r.CleanUp()

	a := assert.New(t)
	c := test.NewClient()
	email := "test@test.localhost"
	pwd := "1aA$_t;3"

	(&auth.Register{
		Email:      email,
		Pwd:        pwd,
		ConfirmPwd: pwd,
	}).MustDo(c)

	// check existing email err
	err := (&auth.Register{
		Email:      email,
		Pwd:        pwd,
		ConfirmPwd: pwd,
	}).Do(c)
	a.Equal(&app.ErrMsg{Status: 400, Msg: "email already registered"}, err)

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

	// check return ealry path
	(&auth.ResendActivateLink{
		Email: email,
	}).MustDo(c)

	id := (&auth.Login{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c).Me

	(&auth.ChangeEmail{
		NewEmail: "change@test.localhost",
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

	(&auth.Delete{
		Pwd: newPwd,
	}).MustDo(c)

	(&auth.Register{
		Email:      email,
		Pwd:        pwd,
		ConfirmPwd: pwd,
	}).MustDo(c)

	row = r.User().Primary().QueryRow(`SELECT activateCode FROM users WHERE email=?`, email)
	PanicOn(row.Scan(&code))

	(&auth.Activate{
		Email: email,
		Code:  code,
	}).MustDo(c)

	id = (&auth.Login{
		Email: email,
		Pwd:   pwd,
	}).MustDo(c).Me
	a.Equal(id, (&auth.Get{}).MustDo(c).Me)

	(&auth.ResetPwd{
		Email: email,
	}).MustDo(c)

	err = (&auth.ResetPwd{
		Email: email,
	}).Do(c)
	a.Equal(400, err.(*app.ErrMsg).Status)
	a.True(regexp.MustCompile(`must wait [1-9][0-9]{2} seconds before reseting pwd again`).MatchString(err.(*app.ErrMsg).Msg))

	_, err = r.User().Primary().Exec(`DELETE FROM users WHERE id=?`, id)
	PanicOn(err)
	_, err = r.Pwd().Primary().Exec(`DELETE FROM pwds WHERE id=?`, id)
	PanicOn(err)
}
