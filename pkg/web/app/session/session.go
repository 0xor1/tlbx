package session

import (
	"net/http"
	"sync"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/gorilla/sessions"
)

type tlbxKey struct{}

func BasicMware(authKey64s, encrKey32s [][]byte, isLocal bool) func(app.Tlbx) {
	return Mware(func(c *Config) {
		c.AuthKey64s = authKey64s
		c.EncrKey32s = encrKey32s
		c.Secure = !isLocal
	})
}

func Mware(configs ...func(*Config)) func(app.Tlbx) {
	c := config(configs...)
	AuthEncrKeyPairs := make([][]byte, 0, len(c.AuthKey64s)*2)
	for i := range c.AuthKey64s {
		PanicIf(len(c.AuthKey64s[i]) != 64, "authKey64s length is not 64")
		PanicIf(len(c.EncrKey32s[i]) != 32, "encrKey32s length is not 32")
		AuthEncrKeyPairs = append(AuthEncrKeyPairs, c.AuthKey64s[i], c.EncrKey32s[i])
	}
	store := sessions.NewCookieStore(AuthEncrKeyPairs...)
	store.Options.Path = c.Path
	store.Options.Domain = c.Domain
	store.Options.MaxAge = c.MaxAge
	store.Options.Secure = c.Secure
	store.Options.HttpOnly = c.HttpOnly
	store.Options.SameSite = c.SameSite
	return func(tlbx app.Tlbx) {
		gorilla, err := store.Get(tlbx.Req(), c.Name)
		PanicOn(err)
		s := &session{
			tlbx:    tlbx,
			gorilla: gorilla,
			mtx:     &sync.RWMutex{},
		}
		if !s.gorilla.IsNew {
			i, ok := s.gorilla.Values["v"]
			if ok {
				s.v = i.([]byte)
			}
		}
		tlbx.Set(tlbxKey{}, s)
	}
}

func Get(tlbx app.Tlbx) Session {
	return tlbx.Get(tlbxKey{}).(Session)
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

type Session interface {
	Exists() bool
	Get() []byte
	Set([]byte)
	Del()
}

type session struct {
	tlbx    app.Tlbx
	v       []byte
	gorilla *sessions.Session
	mtx     *sync.RWMutex
}

func (s *session) Exists() bool {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return len(s.v) > 0
}

func (s *session) Get() []byte {
	s.mtx.RLock()
	defer s.mtx.RUnlock()
	return s.v
}

func (s *session) Set(v []byte) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.v = v
	s.gorilla.Values = map[interface{}]interface{}{
		"v": v,
	}

	PanicOn(s.gorilla.Save(s.tlbx.Req(), s.tlbx.Resp()))
}

func (s *session) Del() {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	s.v = nil
	s.gorilla.Options.MaxAge = -1
	s.gorilla.Values = map[interface{}]interface{}{}
	PanicOn(s.gorilla.Save(s.tlbx.Req(), s.tlbx.Resp()))
}

func config(configs ...func(*Config)) *Config {
	c := &Config{
		AuthKey64s: [][]byte{},
		EncrKey32s: [][]byte{},
		Name:       "s",
		Path:       "/",
		Domain:     "",
		MaxAge:     0,
		Secure:     false,
		HttpOnly:   true,
		SameSite:   http.SameSiteDefaultMode,
	}
	for _, config := range configs {
		config(c)
	}
	return c
}
