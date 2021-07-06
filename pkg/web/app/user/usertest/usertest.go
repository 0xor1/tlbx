package usertest

import (
	"encoding/base64"
	"io/ioutil"
	"regexp"
	"strings"
	"testing"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/config"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/test"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/user/usereps"
	"github.com/stretchr/testify/assert"
)

type appData struct {
	Foo int    `json:"foo"`
	Bar string `json:"bar"`
}

func (*appData) Default() interface{} {
	return &appData{}
}

func (*appData) Example() interface{} {
	return &appData{
		Foo: 1,
		Bar: "yolo",
	}
}

func (*appData) Validate(_ app.Tlbx, ad interface{}) {
	appData := ad.(*appData)
	app.BadReqIf(appData.Foo == 0, "appData.foo must be > 0")
	app.BadReqIf(appData.Bar == "", "appData.bar must be none empty string")
}

func Everything(t *testing.T) {
	r := test.NewMeRig(
		config.GetProcessed(config.GetBase()),
		nil,
		func(r test.Rig, reg *user.Register) {
			reg.AppData = &appData{
				Foo: r.Unique(),
				Bar: *reg.Alias,
			}
		},
		&appData{},
		func(tlbx app.Tlbx, user *user.User, ad interface{}) {
			appData := ad.(*appData)
			app.BadReqIf(appData.Foo == 0, "appData.foo must be > 0")
			app.BadReqIf(appData.Bar == "", "appData.bar must be none empty string")
		},
		func(tlbx app.Tlbx, id ID) {},
		usereps.NopOnSetSocials,
		func(t app.Tlbx, i IDs) (sql.Tx, error) {
			tx := service.Get(t).Data().BeginWrite()
			return tx, nil
		},
		true)
	defer r.CleanUp()

	a := assert.New(t)
	c := r.NewClient()
	handle := "test_" + r.UniqueStr()
	alias := "test 😂 alias"
	email := "test@test.localhost%s" + r.UniqueStr()
	pwd := "1aA$_t;3"

	err := (&user.Register{
		Handle: ptr.String(handle),
		Alias:  ptr.String(alias),
		Email:  email,
		Pwd:    pwd,
	}).Do(c)
	a.Equal(&app.ErrMsg{Status: 400, Msg: "missing appData value"}, err)

	err = (&user.Register{
		Handle: ptr.String(handle),
		Alias:  ptr.String(alias),
		Email:  email,
		Pwd:    pwd,
		AppData: struct {
			Baz string `json:"baz"`
		}{
			Baz: "yolo",
		},
	}).Do(c)
	a.Equal(&app.ErrMsg{Status: 400, Msg: `error unmarshalling json: json: unknown field "baz"`}, err)

	err = (&user.Register{
		Handle:  ptr.String(handle),
		Alias:   ptr.String(alias),
		Email:   email,
		Pwd:     pwd,
		AppData: &appData{},
	}).Do(c)
	a.Equal(&app.ErrMsg{Status: 400, Msg: `appData.foo must be > 0`}, err)

	err = (&user.Register{
		Handle: ptr.String(handle),
		Alias:  ptr.String(alias),
		Email:  email,
		Pwd:    pwd,
		AppData: &appData{
			Foo: 1,
		},
	}).Do(c)
	a.Equal(&app.ErrMsg{Status: 400, Msg: `appData.bar must be none empty string`}, err)

	(&user.Register{
		Handle: ptr.String(handle),
		Alias:  ptr.String(alias),
		Email:  email,
		Pwd:    pwd,
		AppData: &appData{
			Foo: 1,
			Bar: "yolo",
		},
	}).MustDo(c)

	// check existing email err
	err = (&user.Register{
		Handle: ptr.String("not_used"),
		Email:  email,
		Pwd:    pwd,
	}).Do(c)
	a.Equal(&app.ErrMsg{Status: 400, Msg: "email or handle already registered"}, err)

	// check existing handle err
	err = (&user.Register{
		Handle: ptr.String(handle),
		Email:  "email@email.test",
		Pwd:    pwd,
	}).Do(c)
	a.Equal(&app.ErrMsg{Status: 400, Msg: "email or handle already registered"}, err)

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

	tmpFirstID := id.Copy()
	defer func() {
		_, err = r.User().Primary().Exec(`DELETE FROM users WHERE id=?`, tmpFirstID)
		PanicOn(err)
		_, err = r.Pwd().Primary().Exec(`DELETE FROM pwds WHERE id=?`, tmpFirstID)
		PanicOn(err)
	}()

	(&user.SendLoginLinkEmail{Email: email}).MustDo(c)
	row = r.User().Primary().QueryRow(`SELECT loginLinkCode FROM users WHERE id=?`, id)
	loginLinkCode := ""
	PanicOn(row.Scan(&loginLinkCode))
	id = (&user.LoginLinkLogin{
		Me:   id,
		Code: loginLinkCode,
	}).MustDo(c).ID

	(&user.ChangeEmail{
		NewEmail: Strf("change@test.localhost%d", r.Unique()),
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
		OldPwd: pwd,
		NewPwd: newPwd,
	}).MustDo(c)

	(&user.Logout{}).MustDo(c)

	(&user.Login{
		Email: email,
		Pwd:   newPwd,
	}).MustDo(c)

	handle = "new_" + r.UniqueStr()
	(&user.SetHandle{
		Handle: handle,
	}).MustDo(c)

	alias = "shabba!"
	(&user.SetAlias{
		Alias: ptr.String(alias),
	}).MustDo(c)

	me := (&user.GetMe{}).MustDo(c)
	a.Equal(handle, *me.Handle)
	a.Equal(alias, *me.Alias)
	a.False(*me.HasAvatar)

	(&user.SetAvatar{
		Avatar: ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(testImgOk))),
	}).MustDo(c)

	me = (&user.GetMe{}).MustDo(c)
	a.True(*me.HasAvatar)

	avatar := (&user.GetAvatar{
		User: me.ID,
	}).MustDo(c)
	a.Equal("image/png", avatar.Type)
	a.True(me.ID.Equal(avatar.ID))
	a.False(avatar.IsDownload)
	a.Equal(int64(126670), avatar.Size)
	avatar.Content.Close()

	(&user.SetAvatar{
		Avatar: ioutil.NopCloser(base64.NewDecoder(base64.StdEncoding, strings.NewReader(testImgNotSquare))),
	}).MustDo(c)

	me = (&user.GetMe{}).MustDo(c)
	a.True(*me.HasAvatar)

	(&user.SetAvatar{
		Avatar: nil,
	}).MustDo(c)

	me = (&user.GetMe{}).MustDo(c)
	a.False(*me.HasAvatar)

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
		Handle: ptr.String(handle),
		Email:  email,
		Pwd:    pwd,
		AppData: &appData{
			Foo: 1,
			Bar: "yolo",
		},
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
	a.Equal(id, (&user.GetMe{}).MustDo(c).ID)

	defer func() {
		_, err = r.User().Primary().Exec(`DELETE FROM users WHERE id=?`, id)
		PanicOn(err)
		_, err = r.Pwd().Primary().Exec(`DELETE FROM pwds WHERE id=?`, id)
		PanicOn(err)
	}()

	(&user.ResetPwd{
		Email: email,
	}).MustDo(c)

	err = (&user.ResetPwd{
		Email: email,
	}).Do(c)
	a.Equal(400, err.(*app.ErrMsg).Status)
	a.True(regexp.MustCompile(`must wait [1-9][0-9]{2} seconds before reseting pwd again`).MatchString(err.(*app.ErrMsg).Msg))

	// test fcm eps
	ac := r.Ali().Client()
	fcmToken := "123:abc"
	(&user.SetFCMEnabled{
		Val: false,
	}).MustDo(ac)

	(&user.SetFCMEnabled{
		Val: true,
	}).MustDo(ac)

	client1 := (&user.RegisterForFCM{
		Topic: IDs{app.ExampleID()},
		Token: fcmToken,
	}).MustDo(ac)
	a.NotNil(client1)

	idGen := NewIDGen()
	// using client1 so this should overwrite existing fcmTokens row.
	client2 := (&user.RegisterForFCM{
		Topic:  IDs{idGen.MustNew(), idGen.MustNew()},
		Client: client1,
		Token:  fcmToken,
	}).MustDo(ac)
	a.True(client1.Equal(*client2))

	client2 = (&user.RegisterForFCM{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	client2 = (&user.RegisterForFCM{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	client2 = (&user.RegisterForFCM{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	//registered to 5 topics now which is max allowed
	client2 = (&user.RegisterForFCM{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	// this 6th topic should cause the oldest to be bumped out
	// leaving this as the newest of the allowed 5
	client2 = (&user.RegisterForFCM{
		Topic: IDs{idGen.MustNew(), idGen.MustNew()},
		Token: fcmToken,
	}).MustDo(ac)
	a.False(client1.Equal(*client2))

	// toggle off and back on to test sending fcmEnabled:true/false data push
	(&user.SetFCMEnabled{
		Val: false,
	}).MustDo(ac)

	(&user.SetFCMEnabled{
		Val: true,
	}).MustDo(ac)

	js := (&user.GetJin{}).MustDo(ac)
	a.Nil(js)

	(&user.SetJin{
		Val: json.MustFromString(`{"test":"yolo"}`),
	}).MustDo(ac)

	js = (&user.GetJin{}).MustDo(ac)
	a.Equal("yolo", js.MustString("test"))

	(&user.SetJin{}).MustDo(ac)

	js = (&user.GetJin{}).MustDo(ac)
	a.Nil(js)

	(&user.UnregisterFromFCM{
		Client: app.ExampleID(),
	}).MustDo(ac)
}
