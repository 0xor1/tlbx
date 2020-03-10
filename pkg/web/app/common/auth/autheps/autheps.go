package autheps

import (
	"bytes"
	"database/sql"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/crypt"
	"github.com/0xor1/wtf/pkg/json"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/auth"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
	"github.com/0xor1/wtf/pkg/web/app/common/validate"
	"github.com/go-sql-driver/mysql"
)

func New(onDelete func(app.Toolbox, ID), fromEmail, baseHref string) []*app.Endpoint {
	return []*app.Endpoint{
		{
			Description:  "register a new account (requires email link)",
			Path:         (&auth.Register{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.Register{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.Register{
					Email:      "joe@bloggs.example",
					Pwd:        "J03-8l0-Gg5-Pwd",
					ConfirmPwd: "J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				tlbx.ReturnMsgIf(tlbx.Session().IsAuthed(), http.StatusBadRequest, "already logged in")
				args := a.(*auth.Register)
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				activateCode := crypt.UrlSafeString(250)
				id := tlbx.NewID()
				serv := service.Get(tlbx)
				_, err := serv.User().Exec("INSERT INTO users (id, email, registeredOn, activateCode) VALUES (?, ?, ?, ?)", id, args.Email, Now(), activateCode)
				if err != nil {
					mySqlErr, ok := err.(*mysql.MySQLError)
					tlbx.ReturnMsgIf(ok && mySqlErr.Number == 1062, http.StatusBadRequest, "email already registered")
					PanicOn(err)
				}
				setPwd(tlbx, id, args.Pwd, args.ConfirmPwd)
				sendActivateEmail(serv, args.Email, fromEmail, baseHref, &auth.Activate{Email: args.Email, Code: activateCode})
				return nil
			},
		},
		{
			Description:  "resend activate link",
			Path:         (&auth.ResendActivateLink{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.ResendActivateLink{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.ResendActivateLink{
					Email: "joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*auth.ResendActivateLink)
				serv := service.Get(tlbx)
				user := getUser(serv, &args.Email, nil)
				if user == nil || user.ActivateCode == nil {
					return nil
				}
				sendActivateEmail(serv, args.Email, fromEmail, baseHref, &auth.Activate{Email: args.Email, Code: *user.ActivateCode})
				return nil
			},
		},
		{
			Description:  "activate a new account",
			Path:         (&auth.Activate{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.Activate{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.Activate{
					Email: "joe@bloggs.example",
					Code:  "123abc",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*auth.Activate)
				serv := service.Get(tlbx)
				user := getUser(serv, &args.Email, nil)
				tlbx.ReturnMsgIf(*user.ActivateCode != args.Code, http.StatusBadRequest, "")
				now := Now()
				user.ActivatedOn = &now
				user.ActivateCode = nil
				updateUser(serv, user)
				return nil
			},
		},
		{
			Description:  "change email address (requires email link)",
			Path:         (&auth.ChangeEmail{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.ChangeEmail{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.ChangeEmail{
					NewEmail: "new_joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*auth.ChangeEmail)
				validate.Str("email", args.NewEmail, tlbx, 0, emailMaxLen, emailRegex)
				serv := service.Get(tlbx)
				me := tlbx.Me()
				changeEmailCode := crypt.UrlSafeString(250)
				existingUser := getUser(serv, &args.NewEmail, nil)
				tlbx.ReturnMsgIf(existingUser != nil, http.StatusBadRequest, "email already registered")
				user := getUser(serv, nil, &me)
				user.NewEmail = &args.NewEmail
				user.ChangeEmailCode = &changeEmailCode
				updateUser(serv, user)
				sendConfirmChangeEmailEmail(serv, args.NewEmail, fromEmail, baseHref, &auth.ConfirmChangeEmail{Me: me, Code: changeEmailCode})
				return nil
			},
		},
		{
			Description:  "resend change email link",
			Path:         (&auth.ResendChangeEmailLink{}).Path(),
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
			Handler: func(tlbx app.Toolbox, _ interface{}) interface{} {
				serv := service.Get(tlbx)
				me := tlbx.Me()
				user := getUser(serv, nil, &me)
				sendConfirmChangeEmailEmail(serv, *user.NewEmail, fromEmail, baseHref, &auth.ConfirmChangeEmail{Me: me, Code: *user.ChangeEmailCode})
				return nil
			},
		},
		{
			Description:  "confirm change email",
			Path:         (&auth.ConfirmChangeEmail{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.ConfirmChangeEmail{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.ConfirmChangeEmail{
					Me:   app.ExampleID(),
					Code: "123abc",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*auth.ConfirmChangeEmail)
				serv := service.Get(tlbx)
				user := getUser(serv, nil, &args.Me)
				tlbx.ReturnMsgIf(*user.ChangeEmailCode != args.Code, http.StatusBadRequest, "")
				user.ChangeEmailCode = nil
				user.Email = *user.NewEmail
				user.NewEmail = nil
				updateUser(serv, user)
				return nil
			},
		},
		{
			Description:  "reset password (requires email link)",
			Path:         (&auth.ResetPwd{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.ResetPwd{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.ResetPwd{
					Email: "joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*auth.ResetPwd)
				serv := service.Get(tlbx)
				user := getUser(serv, &args.Email, nil)
				if user != nil {
					now := Now()
					if user.LastPwdResetOn != nil {
						mustWaitDur := (10 * time.Minute) - Now().Sub(*user.LastPwdResetOn)
						tlbx.ReturnMsgIf(mustWaitDur > 0, http.StatusBadRequest, "must wait %d seconds before reseting pwd again", int64(math.Ceil(mustWaitDur.Seconds())))
					}
					newPwd := `$aA1` + crypt.UrlSafeString(12)
					setPwd(tlbx, user.ID, newPwd, newPwd)
					sendResetPwdEmail(serv, args.Email, fromEmail, baseHref, newPwd)
					user.LastPwdResetOn = &now
					updateUser(serv, user)
				}
				return nil
			},
		},
		{
			Description:  "set password",
			Path:         (&auth.SetPwd{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.SetPwd{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.SetPwd{
					CurrentPwd:    "J03-8l0-Gg5-Pwd",
					NewPwd:        "N3w-J03-8l0-Gg5-Pwd",
					ConfirmNewPwd: "N3w-J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*auth.SetPwd)
				serv := service.Get(tlbx)
				me := tlbx.Me()
				pwd := getPwd(serv, me)
				tlbx.ReturnMsgIf(!bytes.Equal(crypt.ScryptKey([]byte(args.CurrentPwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen), pwd.Pwd), http.StatusBadRequest, "current pwd does not match")
				setPwd(tlbx, me, args.NewPwd, args.ConfirmNewPwd)
				return nil
			},
		},
		{
			Description:  "delete account",
			Path:         (&auth.Delete{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.Delete{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.Delete{
					Pwd: "J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*auth.Delete)
				serv := service.Get(tlbx)
				me := tlbx.Me()
				tlbx.Session().Logout()
				pwd := getPwd(serv, me)
				tlbx.ReturnMsgIf(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)), http.StatusBadRequest, "incorrect pwd")
				_, err := serv.User().Exec(`DELETE FROM users WHERE id=?`, me)
				PanicOn(err)
				_, err = serv.Pwd().Exec(`DELETE FROM pwds WHERE id=?`, me)
				PanicOn(err)
				if onDelete != nil {
					onDelete(tlbx, me)
				}
				return nil
			},
		},
		{
			Description:  "login",
			Path:         (&auth.Login{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &auth.Login{}
			},
			GetExampleArgs: func() interface{} {
				return &auth.Login{
					Email: "joe@bloggs.example",
					Pwd:   "J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return &auth.LoginRes{
					Me: app.ExampleID(),
				}
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				emailOrPwdMismatch := func(condition bool) {
					tlbx.ReturnMsgIf(condition, http.StatusNotFound, "email and/or pwd are not valid")
				}
				args := a.(*auth.Login)
				serv := service.Get(tlbx)
				user := getUser(serv, &args.Email, nil)
				emailOrPwdMismatch(user == nil)
				pwd := getPwd(serv, user.ID)
				emailOrPwdMismatch(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)))
				// if encryption params have changed re encrypt on successful login
				if len(pwd.Salt) != scryptSaltLen || len(pwd.Pwd) != scryptKeyLen || pwd.N != scryptN || pwd.R != scryptR || pwd.P != scryptP {
					setPwd(tlbx, user.ID, args.Pwd, args.Pwd)
				}
				tlbx.Session().Login(user.ID)
				return &auth.LoginRes{
					Me: user.ID,
				}
			},
		},
		{
			Description:  "logout",
			Path:         (&auth.Logout{}).Path(),
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
			Handler: func(tlbx app.Toolbox, _ interface{}) interface{} {
				tlbx.Session().Logout()
				return nil
			},
		},
		{
			Description:  "get",
			Path:         (&auth.Get{}).Path(),
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
				return &auth.GetRes{
					Me: app.ExampleID(),
				}
			},
			Handler: func(tlbx app.Toolbox, _ interface{}) interface{} {
				return &auth.GetRes{
					Me: tlbx.Me(),
				}
			},
		},
	}
}

var (
	emailRegex  = regexp.MustCompile(`\A.+@.+\..+\z`)
	emailMaxLen = 250
	pwdRegexs   = []*regexp.Regexp{
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
)

func sendActivateEmail(serv service.Layer, sendTo, from, baseHref string, args *auth.Activate) {
	bs, err := json.Marshal(args)
	PanicOn(err)
	link := baseHref + args.Path() + `?args=` + url.QueryEscape(string(bs))
	serv.Email().MustSend([]string{sendTo}, from, "Activate", `<a href="`+link+`">Activate</a>`, `Activate `+link)
}

func sendConfirmChangeEmailEmail(serv service.Layer, sendTo, from, baseHref string, args *auth.ConfirmChangeEmail) {
	bs, err := json.Marshal(args)
	PanicOn(err)
	link := baseHref + args.Path() + `?args=` + url.QueryEscape(string(bs))
	serv.Email().MustSend([]string{sendTo}, from, "Confirm change email", `<a href="`+link+`">Confirm change email</a>`, `Confirm change email `+link)
}

func sendResetPwdEmail(serv service.Layer, sendTo, from, baseHref, newPwd string) {
	serv.Email().MustSend([]string{sendTo}, from, "Pwd Reset", `<p>New Pwd: `+newPwd+`</p>`, `New Pwd: `+newPwd)
}

type user struct {
	ID              ID
	Email           string
	RegisteredOn    time.Time
	ActivatedOn     *time.Time
	NewEmail        *string
	ActivateCode    *string
	ChangeEmailCode *string
	LastPwdResetOn  *time.Time
}

func getUser(serv service.Layer, email *string, id *ID) *user {
	PanicIf(email == nil && id == nil, "one of email or id must not be nil")
	query := `SELECT id, email, registeredOn, activatedOn, newEmail, activateCode, changeEmailCode, lastPwdResetOn FROM users WHERE `
	var arg interface{}
	if email != nil {
		query += `email=?`
		arg = *email
	} else {
		query += `id=?`
		arg = *id
	}
	row := serv.User().QueryRow(query, arg)
	res := &user{}
	err := row.Scan(&res.ID, &res.Email, &res.RegisteredOn, &res.ActivatedOn, &res.NewEmail, &res.ActivateCode, &res.ChangeEmailCode, &res.LastPwdResetOn)
	if err == sql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func updateUser(serv service.Layer, user *user) {
	_, err := serv.User().Exec(`UPDATE users SET email=?, registeredOn=?, activatedOn=?, newEmail=?, activateCode=?, changeEmailCode=?, lastPwdResetOn=? WHERE id=?`, user.Email, user.RegisteredOn, user.ActivatedOn, user.NewEmail, user.ActivateCode, user.ChangeEmailCode, user.LastPwdResetOn, user.ID)
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

func getPwd(serv service.Layer, id ID) *pwd {
	row := serv.Pwd().QueryRow(`SELECT id, salt, pwd, n, r, p FROM pwds WHERE id=?`, id)
	res := &pwd{}
	err := row.Scan(&res.ID, &res.Salt, &res.Pwd, &res.N, &res.R, &res.P)
	if err == sql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func setPwd(tlbx app.Toolbox, id ID, pwd, confirmPwd string) {
	tlbx.ReturnMsgIf(pwd != confirmPwd, http.StatusBadRequest, "pwds do not match")
	validate.Str("pwd", pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
	serv := service.Get(tlbx)
	salt := crypt.Bytes(scryptSaltLen)
	pwdBs := crypt.ScryptKey([]byte(pwd), salt, scryptN, scryptR, scryptP, scryptKeyLen)
	_, err := serv.Pwd().Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE salt=VALUE(salt), pwd=VALUE(pwd), n=VALUE(n), r=VALUE(r), p=VALUE(p)`, id, salt, pwdBs, scryptN, scryptR, scryptP)
	PanicOn(err)
}
