package me

import (
	"net/http"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/session"
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
