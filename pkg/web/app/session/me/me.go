package me

import (
	"net/http"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/iredis"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/ratelimit"
	"github.com/0xor1/tlbx/pkg/web/app/session"
	"github.com/tomasen/realip"
)

func Exists(tlbx app.Tlbx) bool {
	return session.Get(tlbx).Exists()
}

func Get(tlbx app.Tlbx) ID {
	s := session.Get(tlbx)
	tlbx.ExitIf(!s.Exists(), http.StatusUnauthorized, "")
	id := &ID{}
	PanicOn(id.UnmarshalBinary(s.Get()))
	return *id
}

func Set(tlbx app.Tlbx, me ID) {
	meBs, err := me.MarshalBinary()
	PanicOn(err)
	session.Get(tlbx).Set(meBs)
}

func Del(tlbx app.Tlbx) {
	session.Get(tlbx).Del()
}

func RateLimitMware(cache iredis.Pool) func(app.Tlbx) {
	return ratelimit.Mware(func(c *ratelimit.Config) {
		c.KeyGen = func(tlbx app.Tlbx) string {
			var key string
			if Exists(tlbx) {
				key = Get(tlbx).String()
			}
			return Sprintf("rate-limiter-%s-%s", realip.RealIP(tlbx.Req()), key)
		}
		c.Pool = cache
	})
}
