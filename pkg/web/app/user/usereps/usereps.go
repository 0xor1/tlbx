package usereps

//go:generate go get -u github.com/valyala/quicktemplate/qtc
//go:generate qtc -file=usereps.sql

import (
	"bytes"
	"io/ioutil"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"image"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/crypt"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/json"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/service/sql"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	sqlh "github.com/0xor1/tlbx/pkg/web/app/sql"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/disintegration/imaging"
	"github.com/go-sql-driver/mysql"
)

const (
	AvatarBucket = "avatars"
	AvatarPrefix = ""
)

var NopOnSetSocials = func(_ app.Tlbx, _ *user.User) {}

type AppData interface {
	Default() interface{}
	Example() interface{}
	Validate(app.Tlbx, interface{})
}

func New(
	fromEmail,
	activateFmtLink,
	loginLinkFmtLink,
	confirmChangeEmailFmtLink string,
	appData AppData,
	onActivate func(app.Tlbx, *user.User, interface{}),
	onDelete func(app.Tlbx, ID),
	onSetSocials func(app.Tlbx, *user.User),
	validateFcmTopic func(app.Tlbx, IDs) (sql.Tx, error),
	enableJin bool,
) []*app.Endpoint {
	enableSocials := onSetSocials != nil
	enableFCM := validateFcmTopic != nil
	eps := []*app.Endpoint{
		{
			Description:  "register a new account (requires email link)",
			Path:         (&user.Register{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				d := &user.Register{}
				if enableSocials {
					d.Handle = ptr.String("")
					d.Alias = ptr.String("")
				}
				if appData != nil {
					d.AppData = appData.Default()
				}
				return d
			},
			GetExampleArgs: func() interface{} {
				ex := &user.Register{
					Email: "joe@bloggs.example",
					Pwd:   "J03-8l0-Gg5-Pwd",
				}
				if enableSocials {
					ex.Handle = ptr.String("bloe_joggs")
					ex.Alias = ptr.String("Joe Bloggs")
				}
				if appData != nil {
					ex.AppData = appData.Example()
				}
				return ex
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				app.BadReqIf(me.AuthedExists(tlbx), "already logged in")
				args := a.(*user.Register)
				args.Email = StrTrimWS(args.Email)
				if !enableSocials {
					args.Handle = nil
					args.Alias = nil
				}
				app.BadReqIf(enableSocials && args.Handle == nil, "social system requires a user handle")
				if args.Handle != nil {
					args.Handle = ptr.String(
						strings.ReplaceAll(
							StrLower(
								StrTrimWS(*args.Handle)), " ", "_"))
					validate.Str("handle", *args.Handle, tlbx, handleMinLen, handleMaxLen, handleRegex)
				}
				if args.Alias != nil {
					args.Alias = ptr.String(StrTrimWS(*args.Alias))
					validate.Str("alias", *args.Alias, tlbx, 0, aliasMaxLen)
				}
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				activateCode := crypt.UrlSafeString(250)
				id := me.Get(tlbx).ID()
				srv := service.Get(tlbx)
				var hasAvatar *bool
				if enableSocials {
					hasAvatar = ptr.Bool(false)
				}
				var fcmEnabled *bool
				if enableFCM {
					fcmEnabled = ptr.Bool(false)
				}
				usrtx := srv.User().BeginWrite()
				defer usrtx.Rollback()
				_, err := usrtx.Exec("INSERT INTO users (id, email, handle, alias, hasAvatar, fcmEnabled, registeredOn, activatedOn, activateCode) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)", id, args.Email, args.Handle, args.Alias, hasAvatar, fcmEnabled, Now(), time.Time{}, activateCode)
				if err != nil {
					mySqlErr, ok := err.(*mysql.MySQLError)
					app.BadReqIf(ok && mySqlErr.Number == 1062, "email or handle already registered")
					PanicOn(err)
				}
				app.BadReqIf((appData == nil && args.AppData != nil) ||
					(appData != nil && args.AppData == nil), "missing appData value")
				if args.AppData != nil {
					appData.Validate(tlbx, args.AppData)
					// if app requires init ctx data store it in jin
					qryArgs := sqlh.NewArgs(0)
					qry := qryJinInsert(qryArgs, id, args.AppData)
					_, err = usrtx.Exec(qry, qryArgs.Is()...)
					PanicOn(err)
				}
				pwdtx := srv.Pwd().BeginWrite()
				defer pwdtx.Rollback()
				setPwd(tlbx, pwdtx, id, args.Pwd)
				sendActivateEmail(srv, args.Email, fromEmail, Strf(activateFmtLink, id, activateCode), args.Handle)
				usrtx.Commit()
				pwdtx.Commit()
				return nil
			},
		},
		{
			Description:  "resend activate link",
			Path:         (&user.ResendActivateLink{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.ResendActivateLink{}
			},
			GetExampleArgs: func() interface{} {
				return &user.ResendActivateLink{
					Email: "joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.ResendActivateLink)
				srv := service.Get(tlbx)
				tx := srv.User().BeginRead()
				defer tx.Rollback()
				fullUser := getUser(tx, &args.Email, nil)
				tx.Commit()
				if fullUser == nil || fullUser.ActivateCode == nil {
					return nil
				}
				sendActivateEmail(srv, args.Email, fromEmail, Strf(activateFmtLink, fullUser.ID, *fullUser.ActivateCode), fullUser.Handle)
				return nil
			},
		},
		{
			Description:  "activate a new account",
			Path:         (&user.Activate{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.Activate{}
			},
			GetExampleArgs: func() interface{} {
				return &user.Activate{
					Me:   app.ExampleID(),
					Code: "123abc",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.Activate)
				srv := service.Get(tlbx)
				tx := srv.User().BeginWrite()
				defer tx.Rollback()
				user := getUser(tx, nil, &args.Me)
				app.BadReqIf(*user.ActivateCode != args.Code, "")
				var ad interface{}
				if appData != nil {
					ad = appData.Default()
					getJin(tx, user.ID, &ad)
				}
				now := Now()
				user.ActivatedOn = now
				user.ActivateCode = nil
				updateUser(tx, user)
				if onActivate != nil {
					onActivate(tlbx, &user.User, ad)
				}
				delJin(tx, user.ID)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "change email address (requires email link)",
			Path:         (&user.ChangeEmail{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.ChangeEmail{}
			},
			GetExampleArgs: func() interface{} {
				return &user.ChangeEmail{
					NewEmail: "new_joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.ChangeEmail)
				args.NewEmail = StrTrimWS(args.NewEmail)
				validate.Str("email", args.NewEmail, tlbx, 0, emailMaxLen, emailRegex)
				srv := service.Get(tlbx)
				me := me.AuthedGet(tlbx)
				changeEmailCode := crypt.UrlSafeString(250)
				tx := srv.User().BeginWrite()
				defer tx.Rollback()
				existingUser := getUser(tx, &args.NewEmail, nil)
				app.BadReqIf(existingUser != nil, "email already registered")
				fullUser := getUser(tx, nil, &me)
				fullUser.NewEmail = &args.NewEmail
				fullUser.ChangeEmailCode = &changeEmailCode
				updateUser(tx, fullUser)
				tx.Commit()
				sendConfirmChangeEmailEmail(srv, args.NewEmail, fromEmail, Strf(confirmChangeEmailFmtLink, me, changeEmailCode))
				return nil
			},
		},
		{
			Description:  "resend change email link",
			Path:         (&user.ResendChangeEmailLink{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return nil
			},
			GetExampleArgs: func() interface{} {
				return nil
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				srv := service.Get(tlbx)
				me := me.AuthedGet(tlbx)
				tx := srv.User().BeginRead()
				defer tx.Rollback()
				fullUser := getUser(tx, nil, &me)
				tx.Commit()
				sendConfirmChangeEmailEmail(srv, *fullUser.NewEmail, fromEmail, Strf(confirmChangeEmailFmtLink, me, *fullUser.ChangeEmailCode))
				return nil
			},
		},
		{
			Description:  "confirm change email",
			Path:         (&user.ConfirmChangeEmail{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.ConfirmChangeEmail{}
			},
			GetExampleArgs: func() interface{} {
				return &user.ConfirmChangeEmail{
					Me:   app.ExampleID(),
					Code: "123abc",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.ConfirmChangeEmail)
				srv := service.Get(tlbx)
				tx := srv.User().BeginWrite()
				defer tx.Rollback()
				user := getUser(tx, nil, &args.Me)
				app.BadReqIf(*user.ChangeEmailCode != args.Code, "")
				user.ChangeEmailCode = nil
				user.Email = *user.NewEmail
				user.NewEmail = nil
				updateUser(tx, user)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "reset password (requires email link)",
			Path:         (&user.ResetPwd{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.ResetPwd{}
			},
			GetExampleArgs: func() interface{} {
				return &user.ResetPwd{
					Email: "joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.ResetPwd)
				srv := service.Get(tlbx)
				tx := srv.User().BeginWrite()
				defer tx.Rollback()
				user := getUser(tx, &args.Email, nil)
				if user != nil {
					now := Now()
					if user.LastPwdResetOn != nil {
						mustWaitDur := (10 * time.Minute) - Now().Sub(*user.LastPwdResetOn)
						app.BadReqIf(mustWaitDur > 0, "must wait %d seconds before reseting pwd again", int64(math.Ceil(mustWaitDur.Seconds())))
					}
					user.LastPwdResetOn = &now
					updateUser(tx, user)
					pwdtx := srv.Pwd().BeginWrite()
					defer pwdtx.Rollback()
					newPwd := `$aA1` + crypt.UrlSafeString(12)
					setPwd(tlbx, pwdtx, user.ID, newPwd)
					sendResetPwdEmail(srv, args.Email, fromEmail, newPwd)
					pwdtx.Commit()
				}
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "set password",
			Path:         (&user.SetPwd{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.SetPwd{}
			},
			GetExampleArgs: func() interface{} {
				return &user.SetPwd{
					OldPwd: "J03-8l0-Gg5-Pwd",
					NewPwd: "N3w-J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.SetPwd)
				srv := service.Get(tlbx)
				me := me.AuthedGet(tlbx)
				pwdtx := srv.Pwd().BeginWrite()
				defer pwdtx.Rollback()
				pwd := getPwd(pwdtx, me)
				app.BadReqIf(!bytes.Equal(crypt.ScryptKey([]byte(args.OldPwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen), pwd.Pwd), "current pwd does not match")
				setPwd(tlbx, pwdtx, me, args.NewPwd)
				pwdtx.Commit()
				return nil
			},
		},
		{
			Description:  "delete account",
			Path:         (&user.Delete{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &user.Delete{
					Pwd: "J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.Delete)
				srv := service.Get(tlbx)
				m := me.AuthedGet(tlbx)
				pwdtx := srv.Pwd().BeginWrite()
				defer pwdtx.Rollback()
				pwd := getPwd(pwdtx, m)
				app.BadReqIf(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)), "incorrect pwd")
				tx := srv.User().BeginWrite()
				defer tx.Rollback()
				// jin and fcm tokens tables are cleared by foreign key cascade
				_, err := tx.Exec(`DELETE FROM users WHERE id=?`, m)
				PanicOn(err)
				_, err = pwdtx.Exec(`DELETE FROM pwds WHERE id=?`, m)
				PanicOn(err)
				if onDelete != nil {
					onDelete(tlbx, m)
				}
				me.Del(tlbx)
				tx.Commit()
				pwdtx.Commit()
				return nil
			},
		},
		{
			Description:  "login",
			Path:         (&user.Login{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.Login{}
			},
			GetExampleArgs: func() interface{} {
				return &user.Login{
					Email: "joe@bloggs.example",
					Pwd:   "J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				ex := &user.Me{}
				ex.ID = app.ExampleID()
				if enableSocials {
					ex.Handle = ptr.String("bloe_joggs")
					ex.Alias = ptr.String("Joe Bloggs")
					ex.HasAvatar = ptr.Bool(true)
				}
				if enableFCM {
					ex.FcmEnabled = ptr.Bool(true)
				}
				return ex
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				emailOrPwdMismatch := func(condition bool) {
					app.ReturnIf(condition, http.StatusNotFound, "email and/or pwd are not valid")
				}
				args := a.(*user.Login)
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				validate.Str("pwd", args.Pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
				srv := service.Get(tlbx)
				tx := srv.User().BeginRead()
				defer tx.Rollback()
				user := getUser(tx, &args.Email, nil)
				emailOrPwdMismatch(user == nil)
				pwdtx := srv.Pwd().BeginWrite()
				defer pwdtx.Rollback()
				pwd := getPwd(pwdtx, user.ID)
				emailOrPwdMismatch(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)))
				// if encryption params have changed re encrypt on successful login
				if len(pwd.Salt) != scryptSaltLen || len(pwd.Pwd) != scryptKeyLen || pwd.N != scryptN || pwd.R != scryptR || pwd.P != scryptP {
					setPwd(tlbx, pwdtx, user.ID, args.Pwd)
				}
				tx.Commit()
				pwdtx.Commit()
				me.AuthedSet(tlbx, user.ID)
				return &user.Me
			},
		},
		{
			Description:  "send login link email",
			Path:         (&user.SendLoginLinkEmail{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.SendLoginLinkEmail{}
			},
			GetExampleArgs: func() interface{} {
				return &user.SendLoginLinkEmail{
					Email: "joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.SendLoginLinkEmail)
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				srv := service.Get(tlbx)
				tx := srv.User().BeginWrite()
				defer tx.Rollback()
				user := getUser(tx, &args.Email, nil)
				app.BadReqIf(user == nil, "unknown email")
				app.BadReqIf(user.LoginLinkCodeCreatedOn != nil && user.LoginLinkCodeCreatedOn.After(Now().Add(-8*time.Minute)), "An unused login link code still exists")
				user.LoginLinkCodeCreatedOn = ptr.Time(NowMilli())
				user.LoginLinkCode = ptr.String(crypt.UrlSafeString(250))
				updateUser(tx, user)
				sendLoginLinkEmail(srv, user.Email, fromEmail, Strf(loginLinkFmtLink, user.ID, *user.LoginLinkCode), user.Handle)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "login link login",
			Path:         (&user.LoginLinkLogin{}).Path(),
			Timeout:      1000,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.LoginLinkLogin{}
			},
			GetExampleArgs: func() interface{} {
				return &user.LoginLinkLogin{
					Me:   app.ExampleID(),
					Code: "123abc",
				}
			},
			GetExampleResponse: func() interface{} {
				ex := &user.Me{}
				ex.ID = app.ExampleID()
				if enableSocials {
					ex.Handle = ptr.String("bloe_joggs")
					ex.Alias = ptr.String("Joe Bloggs")
					ex.HasAvatar = ptr.Bool(true)
				}
				if enableFCM {
					ex.FcmEnabled = ptr.Bool(true)
				}
				return ex
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.LoginLinkLogin)
				srv := service.Get(tlbx)
				tx := srv.User().BeginWrite()
				defer tx.Rollback()
				user := getUser(tx, nil, &args.Me)
				app.BadReqIf(user == nil, "unknown user")
				app.BadReqIf(user.LoginLinkCodeCreatedOn == nil ||
					user.LoginLinkCodeCreatedOn.Before(Now().Add(-10*time.Minute)) ||
					*user.LoginLinkCode != args.Code, "login code invalid (only valid for 10 minutes from time of creation)")
				user.LoginLinkCodeCreatedOn = nil
				user.LoginLinkCode = nil
				updateUser(tx, user)
				tx.Commit()
				me.AuthedSet(tlbx, user.ID)
				return &user.Me
			},
		},
		{
			Description:  "logout",
			Path:         (&user.Logout{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return nil
			},
			GetExampleArgs: func() interface{} {
				return nil
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				if me.AuthedExists(tlbx) {
					m := me.AuthedGet(tlbx)
					srv := service.Get(tlbx)
					tokens := make([]string, 0, 5)
					tx := srv.User().BeginWrite()
					defer tx.Rollback()
					tx.Query(func(rows isql.Rows) {
						for rows.Next() {
							token := ""
							PanicOn(rows.Scan(&token))
							tokens = append(tokens, token)
						}
					}, `SELECT DISTINCT token FROM fcmTokens WHERE user=?`, m)
					_, err := tx.Exec(`DELETE FROM fcmTokens WHERE user=?`, m)
					PanicOn(err)
					srv.FCM().RawAsyncSend("logout", tokens, map[string]string{}, 0)
					tx.Commit()
					me.Del(tlbx)
				}
				return nil
			},
		},
		{
			Description:  "get me",
			Path:         (&user.GetMe{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return nil
			},
			GetExampleArgs: func() interface{} {
				return nil
			},
			GetExampleResponse: func() interface{} {
				ex := &user.Me{}
				ex.ID = app.ExampleID()
				if enableSocials {
					ex.Handle = ptr.String("bloe_joggs")
					ex.Alias = ptr.String("Joe Bloggs")
					ex.HasAvatar = ptr.Bool(true)
				}
				if enableFCM {
					ex.FcmEnabled = ptr.Bool(true)
				}
				return ex
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				if !me.AuthedExists(tlbx) {
					return nil
				}
				me := me.AuthedGet(tlbx)
				tx := service.Get(tlbx).User().BeginRead()
				defer tx.Rollback()
				user := getUser(tx, nil, &me)
				tx.Commit()
				return &user.Me
			},
		},
	}
	if enableJin {
		eps = append(eps,
			&app.Endpoint{
				Description:  "set users jin (json bin), adhoc json content",
				Path:         (&user.SetJin{}).Path(),
				Timeout:      500,
				MaxBodyBytes: 10 * app.KB,
				IsPrivate:    false,
				GetDefaultArgs: func() interface{} {
					return &user.SetJin{}
				},
				GetExampleArgs: func() interface{} {
					return &user.SetJin{
						Val: exampleJin,
					}
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*user.SetJin)
					me := me.AuthedGet(tlbx)
					srv := service.Get(tlbx)
					qryArgs := sqlh.NewArgs(0)
					var qry string
					if args.Val == nil {
						qry = qryJinDelete(qryArgs, me)
						_, err := srv.User().Exec(qry, qryArgs.Is()...)
						PanicOn(err)
					} else {
						// if app requires init ctx data store it in jin
						qry := qryJinInsert(qryArgs, me, args.Val)
						_, err := srv.User().Exec(qry, qryArgs.Is()...)
						PanicOn(err)
					}
					return nil
				},
			},
			&app.Endpoint{
				Description:  "get users jin (json bin), adhoc json content",
				Path:         (&user.GetJin{}).Path(),
				Timeout:      500,
				MaxBodyBytes: app.KB,
				IsPrivate:    false,
				GetDefaultArgs: func() interface{} {
					return nil
				},
				GetExampleArgs: func() interface{} {
					return nil
				},
				GetExampleResponse: func() interface{} {
					return exampleJin
				},
				Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
					me := me.AuthedGet(tlbx)
					srv := service.Get(tlbx)
					res := &json.Json{}
					tx := srv.User().BeginRead()
					defer tx.Rollback()
					getJin(tx, me, res)
					tx.Commit()
					return res
				},
			})
	}
	if enableSocials {
		eps = append(eps,
			&app.Endpoint{
				Description:  "get users",
				Path:         (&user.Get{}).Path(),
				Timeout:      500,
				MaxBodyBytes: app.KB,
				IsPrivate:    false,
				GetDefaultArgs: func() interface{} {
					return &user.Get{}
				},
				GetExampleArgs: func() interface{} {
					return &user.Get{
						Users: []ID{app.ExampleID()},
					}
				},
				GetExampleResponse: func() interface{} {
					return []user.User{
						{
							ID:        app.ExampleID(),
							Handle:    ptr.String("bloe_joggs"),
							Alias:     ptr.String("Joe Bloggs"),
							HasAvatar: ptr.Bool(true),
						},
					}
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*user.Get)
					if len(args.Users) == 0 {
						return nil
					}
					validate.MaxIDs(tlbx, "users", args.Users, 1000)
					srv := service.Get(tlbx)
					query := bytes.NewBufferString(`SELECT id, handle, alias, hasAvatar FROM users WHERE id IN(?`)
					queryArgs := make([]interface{}, 0, len(args.Users))
					queryArgs = append(queryArgs, args.Users[0])
					for _, id := range args.Users[1:] {
						query.WriteString(`,?`)
						queryArgs = append(queryArgs, id)
					}
					query.WriteString(`)`)
					res := make([]*user.User, 0, len(args.Users))
					PanicOn(srv.User().Query(func(rows isql.Rows) {
						for rows.Next() {
							u := &user.User{}
							PanicOn(rows.Scan(&u.ID, &u.Handle, &u.Alias, &u.HasAvatar))
							res = append(res, u)
						}
					}, query.String(), queryArgs...))
					return res
				},
			}, &app.Endpoint{
				Description:  "set handle",
				Path:         (&user.SetHandle{}).Path(),
				Timeout:      500,
				MaxBodyBytes: app.KB,
				IsPrivate:    false,
				GetDefaultArgs: func() interface{} {
					return &user.SetHandle{}
				},
				GetExampleArgs: func() interface{} {
					return &user.SetHandle{
						Handle: "joe_bloggs",
					}
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*user.SetHandle)
					validate.Str("handle", args.Handle, tlbx, handleMinLen, handleMaxLen, handleRegex)
					srv := service.Get(tlbx)
					me := me.AuthedGet(tlbx)
					tx := srv.User().BeginWrite()
					defer tx.Rollback()
					user := getUser(tx, nil, &me)
					user.Handle = &args.Handle
					updateUser(tx, user)
					if onSetSocials != nil {
						onSetSocials(tlbx, &user.User)
					}
					tx.Commit()
					return nil
				},
			}, &app.Endpoint{
				Description:  "set alias",
				Path:         (&user.SetAlias{}).Path(),
				Timeout:      500,
				MaxBodyBytes: app.KB,
				IsPrivate:    false,
				GetDefaultArgs: func() interface{} {
					return &user.SetAlias{}
				},
				GetExampleArgs: func() interface{} {
					return &user.SetAlias{
						Alias: ptr.String("Boe Jloggs"),
					}
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*user.SetAlias)
					if args.Alias != nil {
						validate.Str("alias", *args.Alias, tlbx, 0, aliasMaxLen)
					}
					srv := service.Get(tlbx)
					me := me.AuthedGet(tlbx)
					tx := srv.User().BeginWrite()
					defer tx.Rollback()
					user := getUser(tx, nil, &me)
					user.Alias = args.Alias
					updateUser(tx, user)
					if onSetSocials != nil {
						onSetSocials(tlbx, &user.User)
					}
					tx.Commit()
					return nil
				},
			}, &app.Endpoint{
				Description:  "set avatar",
				Path:         (&user.SetAvatar{}).Path(),
				Timeout:      500,
				MaxBodyBytes: app.MB,
				IsPrivate:    false,
				GetDefaultArgs: func() interface{} {
					return &app.UpStream{}
				},
				GetExampleArgs: func() interface{} {
					return &app.UpStream{}
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*app.UpStream)
					defer args.Content.Close()
					me := me.AuthedGet(tlbx)
					srv := service.Get(tlbx)
					tx := srv.User().BeginWrite()
					defer tx.Rollback()
					user := getUser(tx, nil, &me)
					content, err := ioutil.ReadAll(args.Content)
					PanicOn(err)
					args.Size = int64(len(content))
					if args.Size > 0 {
						if *user.HasAvatar {
							srv.Store().MustDelete(AvatarBucket, store.Key(AvatarPrefix, me))
						}
						avatar, _, err := image.Decode(bytes.NewBuffer(content))
						PanicOn(err)
						bounds := avatar.Bounds()
						xDiff := bounds.Max.X - bounds.Min.X
						yDiff := bounds.Max.Y - bounds.Min.Y
						if xDiff != yDiff || xDiff != avatarDim || yDiff != avatarDim {
							avatar = imaging.Fill(avatar, avatarDim, avatarDim, imaging.Center, imaging.Lanczos)
						}
						buff := &bytes.Buffer{}
						PanicOn(png.Encode(buff, avatar))
						srv.Store().MustPut(
							AvatarBucket,
							store.Key(AvatarPrefix, me),
							args.Name,
							"image/png",
							int64(buff.Len()),
							true,
							false,
							bytes.NewReader(buff.Bytes()))
					} else if *user.HasAvatar == true {
						srv.Store().MustDelete(AvatarBucket, store.Key(AvatarPrefix, me))
					}
					nowHasAvatar := args.Size > 0
					if *user.HasAvatar != nowHasAvatar {
						user.HasAvatar = ptr.Bool(nowHasAvatar)
						if onSetSocials != nil {
							onSetSocials(tlbx, &user.User)
						}
					}
					updateUser(tx, user)
					tx.Commit()
					return nil
				},
			},
			&app.Endpoint{
				Description:      "get avatar",
				Path:             (&user.GetAvatar{}).Path(),
				Timeout:          500,
				MaxBodyBytes:     app.KB,
				SkipXClientCheck: true,
				IsPrivate:        false,
				GetDefaultArgs: func() interface{} {
					return &user.GetAvatar{}
				},
				GetExampleArgs: func() interface{} {
					return &user.GetAvatar{
						User: app.ExampleID(),
					}
				},
				GetExampleResponse: func() interface{} {
					return &app.DownStream{}
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*user.GetAvatar)
					srv := service.Get(tlbx)
					name, mimeType, size, content := srv.Store().MustGet(AvatarBucket, store.Key(AvatarPrefix, args.User))
					ds := &app.DownStream{}
					ds.ID = args.User
					ds.Name = name
					ds.Type = mimeType
					ds.Size = size
					ds.Content = content
					return ds
				},
			})
	}
	if validateFcmTopic != nil {
		eps = append(eps,
			&app.Endpoint{
				Description:  "set fcm enabled",
				Path:         (&user.SetFCMEnabled{}).Path(),
				Timeout:      500,
				MaxBodyBytes: app.KB,
				IsPrivate:    false,
				GetDefaultArgs: func() interface{} {
					return &user.SetFCMEnabled{
						Val: true,
					}
				},
				GetExampleArgs: func() interface{} {
					return &user.SetFCMEnabled{
						Val: true,
					}
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*user.SetFCMEnabled)
					me := me.AuthedGet(tlbx)
					tx := service.Get(tlbx).User().BeginWrite()
					defer tx.Rollback()
					u := getUser(tx, nil, &me)
					if *u.FcmEnabled == args.Val {
						// not changing anything
						return nil
					}
					u.FcmEnabled = &args.Val
					updateUser(tx, u)
					tokens := make([]string, 0, 5)
					tx.Query(func(rows isql.Rows) {
						for rows.Next() {
							token := ""
							PanicOn(rows.Scan(&token))
							tokens = append(tokens, token)
						}
					}, `SELECT DISTINCT token FROM fcmTokens WHERE user=?`, me)
					tx.Commit()
					if len(tokens) == 0 {
						// no tokens to notify
						return nil
					}
					fcmType := "enabled"
					if !args.Val {
						fcmType = "disabled"
					}
					service.Get(tlbx).FCM().RawAsyncSend(fcmType, tokens, map[string]string{}, 0)
					return nil
				},
			},
			&app.Endpoint{
				Description:  "register for fcm",
				Path:         (&user.RegisterForFCM{}).Path(),
				Timeout:      500,
				MaxBodyBytes: app.KB,
				IsPrivate:    false,
				GetDefaultArgs: func() interface{} {
					return &user.RegisterForFCM{}
				},
				GetExampleArgs: func() interface{} {
					return &user.RegisterForFCM{
						Topic:  IDs{app.ExampleID()},
						Client: ptr.ID(app.ExampleID()),
						Token:  "abc:123",
					}
				},
				GetExampleResponse: func() interface{} {
					return app.ExampleID()
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*user.RegisterForFCM)
					app.BadReqIf(len(args.Topic) == 0 || len(args.Topic) > 5, "topic must contain 1 to 5 ids")
					app.BadReqIf(args.Token == "", "empty string is not a valid fcm token")
					client := args.Client
					if client == nil {
						client = ptr.ID(tlbx.NewID())
					}
					me := me.AuthedGet(tlbx)
					tx := service.Get(tlbx).User().BeginWrite()
					defer tx.Rollback()
					u := getUser(tx, nil, &me)
					app.BadReqIf(u.FcmEnabled == nil || !*u.FcmEnabled, "fcm not enabled for user, please enable first then register for topics")
					// this query is used to get a users 5th token createdOn value if they have one
					row := tx.QueryRow(`SELECT createdOn FROM fcmTokens WHERE user=? ORDER BY createdOn DESC LIMIT 4, 1`, me)
					fifthYoungestTokenCreatedOn := time.Time{}
					sqlh.PanicIfIsntNoRows(row.Scan(&fifthYoungestTokenCreatedOn))
					if !fifthYoungestTokenCreatedOn.IsZero() {
						// this user has 5 topics they're subscribed too already so delete the older ones
						// to make room for this new one
						_, err := tx.Exec(`DELETE FROM fcmTokens WHERE user=? AND createdOn<=?`, me, fifthYoungestTokenCreatedOn)
						PanicOn(err)
					}
					appTx, err := validateFcmTopic(tlbx, args.Topic)
					if appTx != nil {
						defer appTx.Rollback()
					}
					PanicOn(err)
					_, err = tx.Exec(`INSERT INTO fcmTokens (topic, token, user, client, createdOn) VALUES (?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE topic=VALUES(topic), token=VALUES(token), user=VALUES(user), client=VALUES(client), createdOn=VALUES(createdOn)`, args.Topic.StrJoin("_"), args.Token, me, client, tlbx.Start())
					PanicOn(err)
					tx.Commit()
					if appTx != nil {
						appTx.Commit()
					}
					return client
				},
			},
			&app.Endpoint{
				Description:      "unregister from fcm",
				SkipXClientCheck: true,
				Path:             (&user.UnregisterFromFCM{}).Path(),
				Timeout:          500,
				MaxBodyBytes:     app.KB,
				IsPrivate:        false,
				GetDefaultArgs: func() interface{} {
					return &user.UnregisterFromFCM{}
				},
				GetExampleArgs: func() interface{} {
					return &user.UnregisterFromFCM{
						Client: app.ExampleID(),
					}
				},
				GetExampleResponse: func() interface{} {
					return nil
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*user.UnregisterFromFCM)
					me := me.AuthedGet(tlbx)
					tx := service.Get(tlbx).User().BeginWrite()
					defer tx.Rollback()
					_, err := tx.Exec(`DELETE FROM fcmTokens WHERE user=? AND client=?`, me, args.Client)
					PanicOn(err)
					tx.Commit()
					return nil
				},
			})
	}
	return eps
}

var (
	handleRegex  = regexp.MustCompile(`\A[_a-z0-9]{1,20}\z`)
	handleMinLen = 1
	handleMaxLen = 20
	emailRegex   = regexp.MustCompile(`\A.+@.+\..+\z`)
	emailMaxLen  = 250
	aliasMaxLen  = 50
	pwdRegexs    = []*regexp.Regexp{
		regexp.MustCompile(`[0-9]`),
		regexp.MustCompile(`[a-z]`),
		regexp.MustCompile(`[A-Z]`),
		regexp.MustCompile(`[\w]`),
	}
	pwdMinLen     = 8
	pwdMaxLen     = 100
	scryptN       = 32768
	scryptR       = 8
	scryptP       = 1
	scryptSaltLen = 256
	scryptKeyLen  = 256
	avatarDim     = 250
	exampleJin    = json.MustFromString(`{"v":1, "saveDir":"/my/save/dir", "startTab":"favourites"}`)
)

func sendActivateEmail(srv service.Layer, sendTo, from, link string, handle *string) {
	html := `<p>Thank you for registering.</p><p>Click this link to activate your account:</p><p><a href="` + link + `">Activate</a></p><p>If you didn't register for this account you can simply ignore this email.</p>`
	txt := "Thank you for registering.\nClick this link to activate your account:\n\n" + link + "\n\nIf you didn't register for this account you can simply ignore this email."
	if handle != nil {
		html = Strf("Hi %s,\n\n%s", *handle, html)
		txt = Strf("Hi %s,\n\n%s", *handle, txt)
	}
	srv.Email().MustSend([]string{sendTo}, from, "Activate", html, txt)
}

func sendLoginLinkEmail(srv service.Layer, sendTo, from, link string, handle *string) {
	html := `<p>Here is the login link you requested.</p><p>Click this link to login to your account:</p><p><a href="` + link + `">Login</a></p><p>This link will only be valid for 10 minutes.</p><p>If you didn't request this link you can simply ignore this email.</p>`
	txt := "Here is the login link you requested.\nClick this link to login to your account:\n\n" + link + "\n\nThis link will only be valid for 10 minutes.\n\nIf you didn't request this link you can simply ignore this email."
	if handle != nil {
		html = Strf("Hi %s,\n\n%s", *handle, html)
		txt = Strf("Hi %s,\n\n%s", *handle, txt)
	}
	srv.Email().MustSend([]string{sendTo}, from, "Login Link", html, txt)
}

func sendConfirmChangeEmailEmail(srv service.Layer, sendTo, from, link string) {
	srv.Email().MustSend([]string{sendTo}, from, "Confirm change email",
		`<p>Click this link to change the email associated with your account:</p><p><a href="`+link+`">Confirm change email</a></p>`,
		"Confirm change email:\n\n"+link)
}

func sendResetPwdEmail(srv service.Layer, sendTo, from, newPwd string) {
	srv.Email().MustSend([]string{sendTo}, from, "Pwd Reset", `<p>New Pwd: `+newPwd+`</p>`, `New Pwd: `+newPwd)
}

type fullUser struct {
	user.Me
	Email                  string
	RegisteredOn           time.Time
	ActivatedOn            time.Time
	NewEmail               *string
	ActivateCode           *string
	ChangeEmailCode        *string
	LastPwdResetOn         *time.Time
	LoginLinkCodeCreatedOn *time.Time
	LoginLinkCode          *string
}

func getUser(tx sql.Tx, email *string, id *ID) *fullUser {
	PanicIf(email == nil && id == nil, "one of email or id must not be nil")
	query := `SELECT id, email, handle, alias, hasAvatar, fcmEnabled, registeredOn, activatedOn, newEmail, activateCode, changeEmailCode, lastPwdResetOn, loginLinkCodeCreatedOn, loginLinkCode FROM users WHERE `
	var arg interface{}
	if email != nil {
		query += `email=?`
		arg = *email
	} else {
		query += `id=?`
		arg = *id
	}
	row := tx.QueryRow(query, arg)
	res := &fullUser{}
	err := row.Scan(&res.ID, &res.Email, &res.Handle, &res.Alias, &res.HasAvatar, &res.FcmEnabled, &res.RegisteredOn, &res.ActivatedOn, &res.NewEmail, &res.ActivateCode, &res.ChangeEmailCode, &res.LastPwdResetOn, &res.LoginLinkCodeCreatedOn, &res.LoginLinkCode)
	if err == isql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func updateUser(tx sql.Tx, user *fullUser) {
	_, err := tx.Exec(`UPDATE users SET email=?, handle=?, alias=?, hasAvatar=?, fcmEnabled=?, registeredOn=?, activatedOn=?, newEmail=?, activateCode=?, changeEmailCode=?, lastPwdResetOn=?, loginLinkCodeCreatedOn=?, loginLinkCode=? WHERE id=?`, user.Email, user.Handle, user.Alias, user.HasAvatar, user.FcmEnabled, user.RegisteredOn, user.ActivatedOn, user.NewEmail, user.ActivateCode, user.ChangeEmailCode, user.LastPwdResetOn, user.LoginLinkCodeCreatedOn, user.LoginLinkCode, user.ID)
	PanicOn(err)
}

type pwd struct {
	ID   ID
	Salt []byte
	Pwd  []byte
	N    int
	R    int
	P    int
}

func getPwd(pwdtx sql.Tx, id ID) *pwd {
	row := pwdtx.QueryRow(`SELECT id, salt, pwd, n, r, p FROM pwds WHERE id=?`, id)
	res := &pwd{}
	err := row.Scan(&res.ID, &res.Salt, &res.Pwd, &res.N, &res.R, &res.P)
	if err == isql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func setPwd(tlbx app.Tlbx, pwdtx sql.Tx, id ID, pwd string) {
	validate.Str("pwd", pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
	salt := crypt.Bytes(scryptSaltLen)
	pwdBs := crypt.ScryptKey([]byte(pwd), salt, scryptN, scryptR, scryptP, scryptKeyLen)
	_, err := pwdtx.Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE salt=VALUE(salt), pwd=VALUE(pwd), n=VALUE(n), r=VALUE(r), p=VALUE(p)`, id, salt, pwdBs, scryptN, scryptR, scryptP)
	PanicOn(err)
}

func getJin(tx sql.Tx, me ID, dst interface{}) {
	qryArgs := sqlh.NewArgs(0)
	qry := qryJinSelect(qryArgs, me)
	row := tx.QueryRow(qry, qryArgs.Is()...)
	var err error
	if js, ok := dst.(*json.Json); ok {
		err = row.Scan(js)
	} else {
		bs := []byte{}
		err = row.Scan(&bs)
		json.MustUnmarshal(bs, &dst)
	}
	sqlh.PanicIfIsntNoRows(err)
}

func delJin(tx sql.Tx, me ID) {
	qryArgs := sqlh.NewArgs(0)
	qry := qryJinDelete(qryArgs, me)
	_, err := tx.Exec(qry, qryArgs.Is()...)
	PanicOn(err)
}
