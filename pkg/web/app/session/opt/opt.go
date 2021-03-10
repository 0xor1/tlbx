package opt

import (
	"net/http"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app"
	"github.com/0xor1/tlbx/pkg/web/app/session"
)

type Session struct {
	IsAuthed bool
	ID       ID
}

func (s Session) MarshalBinary() ([]byte, error) {
	bs := make([]byte, 17, 17)
	bs[0] = byte('t')
	if !s.IsAuthed {
		bs[0] = byte('f')
	}
	err := s.ID.MarshalBinaryTo(bs[1:])
	if err != nil {
		return nil, err
	}
	return bs, nil
}

func (s *Session) UnmarshalBinary(data []byte) error {
	s.IsAuthed = string(data[0:1]) == `t`
	id := &ID{}
	e := id.UnmarshalBinary(data[1:])
	if e != nil {
		return e
	}
	PanicIfZeroID(*id)
	s.ID = *id
	return nil
}

func Exists(tlbx app.Tlbx) bool {
	return session.Get(tlbx).Exists()
}

func Get(tlbx app.Tlbx) *Session {
	s := session.Get(tlbx)
	app.ReturnIf(!s.Exists(), http.StatusUnauthorized, "")
	ses := &Session{}
	PanicOn(ses.UnmarshalBinary(s.Get()))
	return ses
}

func Set(tlbx app.Tlbx, ses *Session) {
	sesBs, err := ses.MarshalBinary()
	PanicOn(err)
	session.Get(tlbx).Set(sesBs)
}

func Del(tlbx app.Tlbx) {
	session.Get(tlbx).Del()
}

func AuthedExists(tlbx app.Tlbx) bool {
	return Exists(tlbx) && Get(tlbx).IsAuthed
}

func AuthedGet(tlbx app.Tlbx) ID {
	ses := Get(tlbx)
	app.ReturnIf(!ses.IsAuthed, http.StatusUnauthorized, "")
	return ses.ID
}

func AuthedSet(tlbx app.Tlbx, me ID) {
	ses := &Session{
		IsAuthed: true,
		ID:       me,
	}
	sesBs, err := ses.MarshalBinary()
	PanicOn(err)
	session.Get(tlbx).Set(sesBs)
}
