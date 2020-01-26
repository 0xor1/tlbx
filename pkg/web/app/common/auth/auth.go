package auth

import (
	"database/sql"
	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/crypt"
	"github.com/0xor1/wtf/pkg/web/app"
	"github.com/0xor1/wtf/pkg/web/app/common/service"
	"github.com/go-sql-driver/mysql"
	"net/http"
	"net/url"
	"regexp"
)

func Endpoints(onActivate, onDelete func(ID), fromEmail, baseHref string, saltLen, scryptN, scryptR, scryptP, scryptKeyLen int) []*app.Endpoint {
	return []*app.Endpoint{
		{
			Description:  "register a new account (requires email link)",
			Path:         "/api/me/register",
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &register{}
			},
			GetExampleArgs: func() interface{} {
				return &register{
					Name:       "Joe Bloggs",
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

				args := a.(*register)

				validateStr("email", args.Email, tlbx, 0, emailMaxLen, emailRegex)
				validateStr("pwd", args.Pwd, tlbx, pwdMinLen, pwdMaxLen, pwdRegexs...)
				validateStr("name", args.Name, tlbx, 0, nameMaxLen)

				tlbx.ReturnMsgIf(args.Pwd != args.ConfirmPwd, http.StatusBadRequest, "pwds do not match")

				activateCode := crypt.UrlSafeString(250)
				id := tlbx.NewID()

				serv := service.Get(tlbx)

				_, err := serv.User().Exec("INSERT INTO users (id, email, name, registeredOn, activateCode) VALUES (?, ?, ?, ?, ?)", id, args.Email, args.Name, Now(), activateCode)
				if err != nil {
					mySqlErr, ok := err.(*mysql.MySQLError)
					tlbx.ReturnMsgIf(ok && mySqlErr.Number == 1062, http.StatusBadRequest, "email already registered")
					if !ok {
						PanicOn(err)
					}
				}

				salt := crypt.Bytes(saltLen)
				pwd := crypt.ScryptKey([]byte(args.Pwd), salt, scryptN, scryptR, scryptP, scryptKeyLen)
				_, err = serv.Pwd().Exec("INSERT INTO pwds (id, pwd, salt, n, r, p, keyLen) VALUES (?, ?, ?, ?, ?, ?, ?)", id, pwd, salt, scryptN, scryptR, scryptP, scryptKeyLen)
				PanicOn(err)

				sendActivateEmail(serv, args.Email, fromEmail, baseHref, activateCode)

				return nil
			},
		},
		{
			Description:  "resend activate link",
			Path:         "/api/me/resendActivateLink",
			Timeout:      500,
			MaxBodyBytes: app.KB,
			IsPrivate:    false,
			GetDefaultArgs: func() interface{} {
				return &resendActivateLink{}
			},
			GetExampleArgs: func() interface{} {
				return &resendActivateLink{
					Email: "joe@bloggs.example",
				}
			},
			GetExampleResponse: func() interface{} {
				return nil
			},
			Handler: func(tlbx app.Toolbox, a interface{}) interface{} {
				args := a.(*resendActivateLink)
				serv := service.Get(tlbx)
				row := serv.User().QueryRow("SELECT activateCode FROM users WHERE email=? AND activatedOn IS NULL", args.Email)
				var activateCode string
				err := row.Scan(&activateCode)
				if err == sql.ErrNoRows {
					return nil
				}
				PanicOn(err)
				sendActivateEmail(serv, args.Email, fromEmail, baseHref, activateCode)
				return nil
			},
		},
		// {
		// 	Description: "activate a new account",
		// 	Path:        "/api/me/activate",
		// },
		// {
		// 	Description: "change email address (requires email link)",
		// 	Path:        "/api/me/changeEmail",
		// },
		// {
		// 	Description: "resend change email link",
		// 	Path:        "/api/me/resendChangeEmailLink",
		// },
		// {
		// 	Description: "confirm change email",
		// 	Path:        "/api/me/confirmChangeEmail",
		// },
		// {
		// 	Description: "reset password (requires email link)",
		// 	Path:        "/api/me/resetPwd",
		// },
		// {
		// 	Description: "set new password from password reset",
		// 	Path:        "/api/me/setNewPwdFromPwdReset",
		// },
		// {
		// 	Description: "set new password",
		// 	Path:        "/api/me/setPwd",
		// },
		// {
		// 	Description: "set name",
		// 	Path:        "/api/me/setName",
		// },
		// {
		// 	Description: "delete account",
		// 	Path:        "/api/me/delete",
		// },
		// {
		// 	Path: "/api/me/login",
		// },
		// {
		// 	Path: "/api/me/logout",
		// },
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
	pwdMinLen  = 8
	pwdMaxLen  = 100
	nameMaxLen = 100
)

func validateStr(name, str string, tlbx app.Toolbox, minLen, maxLen int, regexs ...*regexp.Regexp) {
	tlbx.ReturnMsgIf(minLen > 0 && StrLen(str) < minLen, http.StatusBadRequest, "%s does not satisfy min len %d", name, minLen)
	tlbx.ReturnMsgIf(maxLen > 0 && StrLen(str) > maxLen, http.StatusBadRequest, "%s does not satisfy max len %d", name, maxLen)
	for _, re := range regexs {
		tlbx.ReturnMsgIf(!re.MatchString(str), http.StatusBadRequest, "%s does not satisfy regexp %s", name, re)
	}
}

type register struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Pwd        string `json:"pwd"`
	ConfirmPwd string `json:"confirmPwd"`
}

type resendActivateLink struct {
	Email string `json:"email"`
}

func sendActivateEmail(serv service.Layer, sendTo, from, baseHref, code string) {
	link := baseHref + `/api/me/activate?args=` + url.QueryEscape(`{"code":"`+code+`"}`)
	serv.Email().MustSend([]string{sendTo}, from, "Activate", `<a href="`+link+`">Activate</a>`, `Activate `+link)
}
