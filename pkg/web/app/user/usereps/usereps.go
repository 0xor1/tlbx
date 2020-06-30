package usereps

import (
	"bytes"
	"database/sql"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/crypt"
	"github.com/0xor1/tlbx/pkg/isql"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/go-sql-driver/mysql"
)

func New(onSetAlias func(app.Tlbx, ID, *string) error, onDelete func(app.Tlbx, ID), fromEmail, activateFmtLink, confirmChangeEmailFmtLink string) []*app.Endpoint {
	eps := []*app.Endpoint{
		{
			Description:  "register a new account (requires email link)",
			Path:         (&user.Register{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.Register{}
			},
			GetExampleArgs: func() interface{} {
				return &user.Register{
					Alias:      ptr.String("Joe Bloggs"),
					Email:      "joe@bloggs.example",
					Pwd:        "J03-8l0-Gg5-Pwd",
					ConfirmPwd: "J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				tlbx.BadReqIf(me.Exists(tlbx), "already logged in")
				args := a.(*user.Register)
				args.Email = strings.Trim(args.Email, " ")
				if args.Alias != nil {
					args.Alias = ptr.String(strings.Trim(*args.Alias, " "))
					validate.Str("alias", *args.Alias, tlbx, 0, aliasMaxLen)
				}
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				activateCode := crypt.UrlSafeString(250)
				id := tlbx.NewID()
				srv := service.Get(tlbx)
				_, err := srv.User().Exec("INSERT INTO users (id, email, alias, registeredOn, activateCode) VALUES (?, ?, ?, ?, ?)", id, args.Email, args.Alias, Now(), activateCode)
				if err != nil {
					mySqlErr, ok := err.(*mysql.MySQLError)
					tlbx.BadReqIf(ok && mySqlErr.Number == 1062, "email already registered")
					PanicOn(err)
				}
				setPwd(tlbx, id, args.Pwd, args.ConfirmPwd)
				sendActivateEmail(srv, args.Email, fromEmail, Sprintf(activateFmtLink, args.Email, activateCode))
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
				fullUser := getUser(srv, &args.Email, nil)
				if fullUser == nil || fullUser.ActivateCode == nil {
					return nil
				}
				sendActivateEmail(srv, args.Email, fromEmail, Sprintf(activateFmtLink, args.Email, *fullUser.ActivateCode))
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
					Email: "joe@bloggs.example",
					Code:  "123abc",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.Activate)
				srv := service.Get(tlbx)
				user := getUser(srv, &args.Email, nil)
				tlbx.BadReqIf(*user.ActivateCode != args.Code, "")
				now := Now()
				user.ActivatedOn = &now
				user.ActivateCode = nil
				tx := srv.User().Begin()
				updateUser(tx, user)
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
				args.NewEmail = strings.Trim(args.NewEmail, " ")
				validate.Str("email", args.NewEmail, tlbx, 0, emailMaxLen, emailRegex)
				srv := service.Get(tlbx)
				me := me.Get(tlbx)
				changeEmailCode := crypt.UrlSafeString(250)
				existingUser := getUser(srv, &args.NewEmail, nil)
				tlbx.BadReqIf(existingUser != nil, "email already registered")
				fullUser := getUser(srv, nil, &me)
				fullUser.NewEmail = &args.NewEmail
				fullUser.ChangeEmailCode = &changeEmailCode
				tx := srv.User().Begin()
				updateUser(tx, fullUser)
				tx.Commit()
				sendConfirmChangeEmailEmail(srv, args.NewEmail, fromEmail, Sprintf(confirmChangeEmailFmtLink, me, changeEmailCode))
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
				me := me.Get(tlbx)
				fullUser := getUser(srv, nil, &me)
				sendConfirmChangeEmailEmail(srv, *fullUser.NewEmail, fromEmail, Sprintf(confirmChangeEmailFmtLink, me, *fullUser.ChangeEmailCode))
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
				user := getUser(srv, nil, &args.Me)
				tlbx.BadReqIf(*user.ChangeEmailCode != args.Code, "")
				user.ChangeEmailCode = nil
				user.Email = *user.NewEmail
				user.NewEmail = nil
				tx := srv.User().Begin()
				updateUser(tx, user)
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "reset password (requires email link)",
			Path:         (&user.ResetPwd{}).Path(),
			Timeout:      500,
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
				user := getUser(srv, &args.Email, nil)
				if user != nil {
					now := Now()
					if user.LastPwdResetOn != nil {
						mustWaitDur := (10 * time.Minute) - Now().Sub(*user.LastPwdResetOn)
						tlbx.BadReqIf(mustWaitDur > 0, "must wait %d seconds before reseting pwd again", int64(math.Ceil(mustWaitDur.Seconds())))
					}
					newPwd := `$aA1` + crypt.UrlSafeString(12)
					setPwd(tlbx, user.ID, newPwd, newPwd)
					sendResetPwdEmail(srv, args.Email, fromEmail, newPwd)
					user.LastPwdResetOn = &now
					tx := srv.User().Begin()
					updateUser(tx, user)
					tx.Commit()
				}
				return nil
			},
		},
		{
			Description:  "change alias",
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
				me := me.Get(tlbx)
				user := getUser(srv, nil, &me)
				user.Alias = args.Alias
				tx := srv.User().Begin()
				defer tx.Rollback()
				updateUser(tx, user)
				if onSetAlias != nil {
					PanicOn(onSetAlias(tlbx, me, user.Alias))
				}
				tx.Commit()
				return nil
			},
		},
		{
			Description:  "set password",
			Path:         (&user.SetPwd{}).Path(),
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &user.SetPwd{}
			},
			GetExampleArgs: func() interface{} {
				return &user.SetPwd{
					CurrentPwd:    "J03-8l0-Gg5-Pwd",
					NewPwd:        "N3w-J03-8l0-Gg5-Pwd",
					ConfirmNewPwd: "N3w-J03-8l0-Gg5-Pwd",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.SetPwd)
				srv := service.Get(tlbx)
				me := me.Get(tlbx)
				pwd := getPwd(srv, me)
				tlbx.BadReqIf(!bytes.Equal(crypt.ScryptKey([]byte(args.CurrentPwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen), pwd.Pwd), "current pwd does not match")
				setPwd(tlbx, me, args.NewPwd, args.ConfirmNewPwd)
				return nil
			},
		},
		{
			Description:  "delete account",
			Path:         (&user.Delete{}).Path(),
			Timeout:      500,
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
				m := me.Get(tlbx)
				me.Del(tlbx)
				pwd := getPwd(srv, m)
				tlbx.BadReqIf(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)), "incorrect pwd")
				if onDelete != nil {
					onDelete(tlbx, m)
				}
				_, err := srv.User().Exec(`DELETE FROM users WHERE id=?`, m)
				PanicOn(err)
				_, err = srv.Pwd().Exec(`DELETE FROM pwds WHERE id=?`, m)
				PanicOn(err)
				return nil
			},
		},
		{
			Description:  "login",
			Path:         (&user.Login{}).Path(),
			Timeout:      500,
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
				return &user.User{
					ID: app.ExampleID(),
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				emailOrPwdMismatch := func(condition bool) {
					tlbx.ExitIf(condition, http.StatusNotFound, "email and/or pwd are not valid")
				}
				args := a.(*user.Login)
				validate.Str("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				validate.Str("pwd", args.Pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
				srv := service.Get(tlbx)
				user := getUser(srv, &args.Email, nil)
				emailOrPwdMismatch(user == nil)
				pwd := getPwd(srv, user.ID)
				emailOrPwdMismatch(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)))
				// if encryption params have changed re encrypt on successful login
				if len(pwd.Salt) != scryptSaltLen || len(pwd.Pwd) != scryptKeyLen || pwd.N != scryptN || pwd.R != scryptR || pwd.P != scryptP {
					setPwd(tlbx, user.ID, args.Pwd, args.Pwd)
				}
				me.Set(tlbx, user.ID)
				return &user.User
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
				me.Del(tlbx)
				return nil
			},
		},
		{
			Description:  "get me",
			Path:         (&user.Me{}).Path(),
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
				return &user.User{
					ID: app.ExampleID(),
				}
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				me := me.Get(tlbx)
				user := getUser(service.Get(tlbx), nil, &me)
				return &user.User
			},
		},
		{
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
						ID:    app.ExampleID(),
						Alias: ptr.String("Joe Bloggs"),
					},
				}
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				args := a.(*user.Get)
				if len(args.Users) == 0 {
					return nil
				}
				validate.MaxIDs(tlbx, "users", args.Users, 100)
				srv := service.Get(tlbx)
				query := bytes.NewBufferString(`SELECT id, alias FROM users WHERE id IN(?`)
				queryArgs := make([]interface{}, 0, len(args.Users))
				queryArgs = append(queryArgs, args.Users[0])
				for _, id := range args.Users[1:] {
					query.WriteString(`,?`)
					queryArgs = append(queryArgs, id)
				}
				query.WriteString(`)`)
				res := make([]*user.User, 0, len(args.Users))
				srv.User().Query(func(rows isql.Rows) {
					for rows.Next() {
						u := &user.User{}
						rows.Scan(&u.ID, &u.Alias)
						res = append(res, u)
					}
				}, query.String(), queryArgs...)
				return res
			},
		},
	}

	return eps
}

var (
	emailRegex  = regexp.MustCompile(`\A.+@.+\..+\z`)
	emailMaxLen = 250
	aliasMaxLen = 250
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

func sendActivateEmail(srv service.Layer, sendTo, from, link string) {
	srv.Email().MustSend([]string{sendTo}, from, "Activate", `<a href="`+link+`">Activate</a>`, `Activate `+link)
}

func sendConfirmChangeEmailEmail(srv service.Layer, sendTo, from, link string) {
	srv.Email().MustSend([]string{sendTo}, from, "Confirm change email", `<a href="`+link+`">Confirm change email</a>`, `Confirm change email `+link)
}

func sendResetPwdEmail(srv service.Layer, sendTo, from, newPwd string) {
	srv.Email().MustSend([]string{sendTo}, from, "Pwd Reset", `<p>New Pwd: `+newPwd+`</p>`, `New Pwd: `+newPwd)
}

type fullUser struct {
	user.User
	Email           string
	RegisteredOn    time.Time
	ActivatedOn     *time.Time
	NewEmail        *string
	ActivateCode    *string
	ChangeEmailCode *string
	LastPwdResetOn  *time.Time
}

func getUser(srv service.Layer, email *string, id *ID) *fullUser {
	PanicIf(email == nil && id == nil, "one of email or id must not be nil")
	query := `SELECT id, email, alias, registeredOn, activatedOn, newEmail, activateCode, changeEmailCode, lastPwdResetOn FROM users WHERE `
	var arg interface{}
	if email != nil {
		query += `email=?`
		arg = *email
	} else {
		query += `id=?`
		arg = *id
	}
	row := srv.User().QueryRow(query, arg)
	res := &fullUser{}
	err := row.Scan(&res.ID, &res.Email, &res.Alias, &res.RegisteredOn, &res.ActivatedOn, &res.NewEmail, &res.ActivateCode, &res.ChangeEmailCode, &res.LastPwdResetOn)
	if err == sql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func updateUser(tx service.Tx, user *fullUser) {
	_, err := tx.Exec(`UPDATE users SET email=?, alias=?, registeredOn=?, activatedOn=?, newEmail=?, activateCode=?, changeEmailCode=?, lastPwdResetOn=? WHERE id=?`, user.Email, user.Alias, user.RegisteredOn, user.ActivatedOn, user.NewEmail, user.ActivateCode, user.ChangeEmailCode, user.LastPwdResetOn, user.ID)
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

func getPwd(srv service.Layer, id ID) *pwd {
	row := srv.Pwd().QueryRow(`SELECT id, salt, pwd, n, r, p FROM pwds WHERE id=?`, id)
	res := &pwd{}
	err := row.Scan(&res.ID, &res.Salt, &res.Pwd, &res.N, &res.R, &res.P)
	if err == sql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func setPwd(tlbx app.Tlbx, id ID, pwd, confirmPwd string) {
	tlbx.BadReqIf(pwd != confirmPwd, "pwds do not match")
	validate.Str("pwd", pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
	srv := service.Get(tlbx)
	salt := crypt.Bytes(scryptSaltLen)
	pwdBs := crypt.ScryptKey([]byte(pwd), salt, scryptN, scryptR, scryptP, scryptKeyLen)
	_, err := srv.Pwd().Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE salt=VALUE(salt), pwd=VALUE(pwd), n=VALUE(n), r=VALUE(r), p=VALUE(p)`, id, salt, pwdBs, scryptN, scryptR, scryptP)
	PanicOn(err)
}
