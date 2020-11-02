package usereps

import (
	"bytes"
	"database/sql"
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
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/0xor1/tlbx/pkg/store"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/service"
	"github.com/0xor1/tlbx/pkg/web/app/session/me"
	"github.com/0xor1/tlbx/pkg/web/app/user"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
	"github.com/disintegration/imaging"
	"github.com/go-sql-driver/mysql"
)

func New(
	fromEmail,
	activateFmtLink,
	confirmChangeEmailFmtLink string,
	onActivate func(app.Tlbx, *user.User),
	onDelete func(app.Tlbx, ID),
	enableSocials bool,
	onSetSocials func(app.Tlbx, *user.User) error,
	avatarBucket, avatarPrefix string,
	storeClient store.Client,
) []*app.Endpoint {
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
				ex := &user.Register{
					Email:      "joe@bloggs.example",
					Pwd:        "J03-8l0-Gg5-Pwd",
					ConfirmPwd: "J03-8l0-Gg5-Pwd",
				}
				if enableSocials {
					ex.Handle = ptr.String("bloe_joggs")
					ex.Alias = ptr.String("Joe Bloggs")
				}
				return ex
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
				app.BadReqIf(me.Exists(tlbx), "already logged in")
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
				id := tlbx.NewID()
				srv := service.Get(tlbx)
				var hasAvatar *bool
				if enableSocials {
					hasAvatar = ptr.Bool(false)
				}
				_, err := srv.User().Exec("INSERT INTO users (id, email, handle, alias, hasAvatar, registeredOn, activateCode) VALUES (?, ?, ?, ?, ?, ?, ?)", id, args.Email, args.Handle, args.Alias, hasAvatar, Now(), activateCode)
				if err != nil {
					mySqlErr, ok := err.(*mysql.MySQLError)
					app.BadReqIf(ok && mySqlErr.Number == 1062, "email or handle already registered")
					PanicOn(err)
				}
				setPwd(tlbx, id, args.Pwd, args.ConfirmPwd)
				sendActivateEmail(srv, args.Email, fromEmail, Strf(activateFmtLink, args.Email, activateCode))
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
				tx := srv.User().Begin()
				defer tx.Rollback()
				fullUser := getUser(tx, &args.Email, nil)
				tx.Commit()
				if fullUser == nil || fullUser.ActivateCode == nil {
					return nil
				}
				sendActivateEmail(srv, args.Email, fromEmail, Strf(activateFmtLink, args.Email, *fullUser.ActivateCode))
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
				tx := srv.User().Begin()
				defer tx.Rollback()
				user := getUser(tx, &args.Email, nil)
				app.BadReqIf(*user.ActivateCode != args.Code, "")
				now := Now()
				user.ActivatedOn = &now
				user.ActivateCode = nil
				updateUser(tx, user)
				if onActivate != nil {
					onActivate(tlbx, &user.User)
				}
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
				me := me.Get(tlbx)
				changeEmailCode := crypt.UrlSafeString(250)
				tx := srv.User().Begin()
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
				me := me.Get(tlbx)
				tx := srv.User().Begin()
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
				tx := srv.User().Begin()
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
				tx := srv.User().Begin()
				defer tx.Rollback()
				user := getUser(tx, &args.Email, nil)
				if user != nil {
					now := Now()
					if user.LastPwdResetOn != nil {
						mustWaitDur := (10 * time.Minute) - Now().Sub(*user.LastPwdResetOn)
						app.BadReqIf(mustWaitDur > 0, "must wait %d seconds before reseting pwd again", int64(math.Ceil(mustWaitDur.Seconds())))
					}
					newPwd := `$aA1` + crypt.UrlSafeString(12)
					setPwd(tlbx, user.ID, newPwd, newPwd)
					sendResetPwdEmail(srv, args.Email, fromEmail, newPwd)
					user.LastPwdResetOn = &now
					updateUser(tx, user)
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
				app.BadReqIf(!bytes.Equal(crypt.ScryptKey([]byte(args.CurrentPwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen), pwd.Pwd), "current pwd does not match")
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
				app.BadReqIf(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)), "incorrect pwd")
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
				ex := &user.User{
					ID: app.ExampleID(),
				}
				if enableSocials {
					ex.Handle = ptr.String("bloe_joggs")
					ex.Alias = ptr.String("Joe Bloggs")
					ex.HasAvatar = ptr.Bool(true)
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
				tx := srv.User().Begin()
				defer tx.Rollback()
				user := getUser(tx, &args.Email, nil)
				emailOrPwdMismatch(user == nil)
				pwd := getPwd(srv, user.ID)
				emailOrPwdMismatch(!bytes.Equal(pwd.Pwd, crypt.ScryptKey([]byte(args.Pwd), pwd.Salt, pwd.N, pwd.R, pwd.P, scryptKeyLen)))
				// if encryption params have changed re encrypt on successful login
				if len(pwd.Salt) != scryptSaltLen || len(pwd.Pwd) != scryptKeyLen || pwd.N != scryptN || pwd.R != scryptR || pwd.P != scryptP {
					setPwd(tlbx, user.ID, args.Pwd, args.Pwd)
				}
				tx.Commit()
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
				ex := &user.User{
					ID: app.ExampleID(),
				}
				if enableSocials {
					ex.Handle = ptr.String("bloe_joggs")
					ex.Alias = ptr.String("Joe Bloggs")
					ex.HasAvatar = ptr.Bool(true)
				}
				return ex
			},
			Handler: func(tlbx app.Tlbx, _ interface{}) interface{} {
				me := me.Get(tlbx)
				tx := service.Get(tlbx).User().Begin()
				defer tx.Rollback()
				user := getUser(tx, nil, &me)
				tx.Commit()
				return &user.User
			},
		},
	}
	if enableSocials {
		storeClient.MustCreateBucket(avatarBucket, "public_read")
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
					ex := []user.User{
						{
							ID:        app.ExampleID(),
							Handle:    ptr.String("bloe_joggs"),
							Alias:     ptr.String("Joe Bloggs"),
							HasAvatar: ptr.Bool(true),
						},
					}
					return ex
				},
				Handler: func(tlbx app.Tlbx, a interface{}) interface{} {
					args := a.(*user.Get)
					if len(args.Users) == 0 {
						return nil
					}
					validate.MaxIDs(tlbx, "users", args.Users, 100)
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
					me := me.Get(tlbx)
					tx := srv.User().Begin()
					defer tx.Rollback()
					user := getUser(tx, nil, &me)
					user.Handle = &args.Handle
					updateUser(tx, user)
					if onSetSocials != nil {
						PanicOn(onSetSocials(tlbx, &user.User))
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
					me := me.Get(tlbx)
					tx := srv.User().Begin()
					defer tx.Rollback()
					user := getUser(tx, nil, &me)
					user.Alias = args.Alias
					updateUser(tx, user)
					if onSetSocials != nil {
						PanicOn(onSetSocials(tlbx, &user.User))
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
					me := me.Get(tlbx)
					srv := service.Get(tlbx)
					tx := srv.User().Begin()
					defer tx.Rollback()
					user := getUser(tx, nil, &me)
					content, err := ioutil.ReadAll(args.Content)
					PanicOn(err)
					args.Size = int64(len(content))
					if args.Size > 0 {
						if *user.HasAvatar {
							srv.Store().MustDelete(avatarBucket, store.Key(avatarPrefix, me))
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
							avatarBucket,
							store.Key(avatarPrefix, me),
							args.Name,
							"image/png",
							int64(buff.Len()),
							true,
							false,
							bytes.NewReader(buff.Bytes()))
					} else if *user.HasAvatar == true {
						srv.Store().MustDelete(avatarBucket, store.Key(avatarPrefix, me))
					}
					nowHasAvatar := args.Size > 0
					if *user.HasAvatar != nowHasAvatar {
						user.HasAvatar = ptr.Bool(nowHasAvatar)
						if onSetSocials != nil {
							PanicOn(onSetSocials(tlbx, &user.User))
						}
					}
					updateUser(tx, user)
					tx.Commit()
					return nil
				},
			},
			&app.Endpoint{
				Description:  "get avatar",
				Path:         (&user.GetAvatar{}).Path(),
				Timeout:      500,
				MaxBodyBytes: app.KB,
				IsPrivate:    false,
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
					name, mimeType, size, content := srv.Store().MustGet(avatarBucket, store.Key(avatarPrefix, args.User))
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
	return eps
}

var (
	handleRegex  = regexp.MustCompile(`[_a-z0-9]{1,15}`)
	handleMinLen = 1
	handleMaxLen = 15
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

func getUser(tx service.Tx, email *string, id *ID) *fullUser {
	PanicIf(email == nil && id == nil, "one of email or id must not be nil")
	query := `SELECT id, email, handle, alias, hasAvatar, registeredOn, activatedOn, newEmail, activateCode, changeEmailCode, lastPwdResetOn FROM users WHERE `
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
	err := row.Scan(&res.ID, &res.Email, &res.Handle, &res.Alias, &res.HasAvatar, &res.RegisteredOn, &res.ActivatedOn, &res.NewEmail, &res.ActivateCode, &res.ChangeEmailCode, &res.LastPwdResetOn)
	if err == sql.ErrNoRows {
		return nil
	}
	PanicOn(err)
	return res
}

func updateUser(tx service.Tx, user *fullUser) {
	_, err := tx.Exec(`UPDATE users SET email=?, handle=?, alias=?, hasAvatar=?, registeredOn=?, activatedOn=?, newEmail=?, activateCode=?, changeEmailCode=?, lastPwdResetOn=? WHERE id=?`, user.Email, user.Handle, user.Alias, user.HasAvatar, user.RegisteredOn, user.ActivatedOn, user.NewEmail, user.ActivateCode, user.ChangeEmailCode, user.LastPwdResetOn, user.ID)
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
	app.BadReqIf(pwd != confirmPwd, "pwds do not match")
	validate.Str("pwd", pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
	srv := service.Get(tlbx)
	salt := crypt.Bytes(scryptSaltLen)
	pwdBs := crypt.ScryptKey([]byte(pwd), salt, scryptN, scryptR, scryptP, scryptKeyLen)
	_, err := srv.Pwd().Exec(`INSERT INTO pwds (id, salt, pwd, n, r, p) VALUES (?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE salt=VALUE(salt), pwd=VALUE(pwd), n=VALUE(n), r=VALUE(r), p=VALUE(p)`, id, salt, pwdBs, scryptN, scryptR, scryptP)
	PanicOn(err)
}
