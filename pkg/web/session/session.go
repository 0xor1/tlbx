package session

import (
	"encoding/gob"
	"net/http"
	"time"

	. "github.com/0xor1/wtf/pkg/core"
	"github.com/0xor1/wtf/pkg/crypt"
	"github.com/0xor1/wtf/pkg/web/toolbox"
	"github.com/gorilla/context"
	"github.com/gorilla/sessions"
)

func Mware(next http.HandlerFunc, configs ...func(*Config)) http.HandlerFunc {
	c := &Config{
		AuthKey64s: [][]byte{crypt.Bytes(64)},
		EncrKey32s: [][]byte{crypt.Bytes(32)},
		Name:       "s",
		Path:       "",
		Domain:     "",
		MaxAge:     0,
		Secure:     true,
		HttpOnly:   true,
		SameSite:   http.SameSiteDefaultMode,
	}
	for _, config := range configs {
		config(c)
	}

	sessionAuthEncrKeyPairs := make([][]byte, 0, len(c.AuthKey64s)*2)
	for i := range c.AuthKey64s {
		PanicIf(len(c.AuthKey64s[i]) != 64, "authKey64s length is not 64")
		PanicIf(len(c.EncrKey32s[i]) != 32, "encrKey32s length is not 32")
		sessionAuthEncrKeyPairs = append(sessionAuthEncrKeyPairs, c.AuthKey64s[i], c.EncrKey32s[i])
	}
	sessionStore := sessions.NewCookieStore(sessionAuthEncrKeyPairs...)
	sessionStore.Options.Path = c.Path
	sessionStore.Options.Domain = c.Domain
	sessionStore.Options.MaxAge = c.MaxAge
	sessionStore.Options.Secure = c.Secure
	sessionStore.Options.HttpOnly = c.HttpOnly
	sessionStore.Options.SameSite = c.SameSite
	// register types for sessionCookie
	gob.Register(NewIDGen().MustNew())
	gob.Register(time.Time{})
	gob.Register(&sessionCore{})
	return func(w http.ResponseWriter, r *http.Request) {
		defer context.Clear(r)
		gses, err := sessionStore.Get(r, c.Name)
		PanicOn(err)
		sesW := &sessionWrapper{
			r:       r,
			w:       w,
			gorilla: gses,
		}
		i, ok := gses.Values["s"]
		if ok {
			sesW.session = i.(*sessionCore)
		}
		toolbox.Get(r).Set(tlbxKey{}, sesW)
		next(w, r)
	}
}

func Get(r *http.Request) Session {
	return toolbox.Get(r).Get(tlbxKey{}).(Session)
}

type Session interface {
	Me() *ID
	AuthedOn() time.Time
	Login(ID)
	Logout()
}

type Config struct {
	AuthKey64s [][]byte
	EncrKey32s [][]byte
	Name       string
	Path       string
	Domain     string
	MaxAge     int
	Secure     bool
	HttpOnly   bool
	SameSite   http.SameSite
}

type sessionCore struct {
	Me_       ID
	AuthedOn_ time.Time
}

type sessionWrapper struct {
	w       http.ResponseWriter
	r       *http.Request
	session *sessionCore
	gorilla *sessions.Session
}

func (s *sessionWrapper) Me() *ID {
	if s.session != nil {
		return &s.session.Me_
	}
	return nil
}

func (s *sessionWrapper) AuthedOn() time.Time {
	if s.session != nil {
		return s.session.AuthedOn_
	}
	return time.Time{}
}

func (s *sessionWrapper) Login(me ID) {
	s.session = &sessionCore{
		Me_:       me,
		AuthedOn_: Now(),
	}
	s.gorilla.Values = map[interface{}]interface{}{
		"s": s.session,
	}
	s.gorilla.Save(s.r, s.w)
}

func (s *sessionWrapper) Logout(w http.ResponseWriter, r *http.Request) {
	s.session = nil
	s.gorilla.Options.MaxAge = -1
	s.gorilla.Values = map[interface{}]interface{}{}
	s.gorilla.Save(r, w)
}

type tlbxKey struct{}
