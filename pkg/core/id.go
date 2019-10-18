package core

import (
	"database/sql/driver"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	now    = Now().UnixNano()
	nowMtx = &sync.Mutex{}
)

type IDGenerator interface {
	New() (ID, error)
	MustNew() ID
}

func NewIDGenerator() IDGenerator {
	nowMtx.Lock()
	t := now
	now++
	nowMtx.Unlock()
	return &idGenerator{
		entropyMtx: &sync.Mutex{},
		entropy:    rand.New(rand.NewSource(t)),
	}
}

type idGenerator struct {
	entropyMtx *sync.Mutex
	entropy    io.Reader
}

func (g *idGenerator) New() (ID, error) {
	g.entropyMtx.Lock()
	defer g.entropyMtx.Unlock()
	id, e := ulid.New(ulid.Now(), g.entropy)
	return ID(id), e
}

func (g *idGenerator) MustNew() ID {
	id, err := g.New()
	PanicOn(err)
	return id
}

type ID ulid.ULID

func (id ID) MarshalBinary() ([]byte, error) {
	return ulid.ULID(id).MarshalBinary()
}

func (id ID) MarshalBinaryTo(dst []byte) error {
	return ulid.ULID(id).MarshalBinaryTo(dst)
}

func (id *ID) UnmarshalBinary(data []byte) error {
	ulid := &ulid.ULID{}
	e := ulid.UnmarshalBinary(data)
	if e != nil {
		return e
	}
	*id = ID(*ulid)
	return nil
}

func (id ID) MarshalText() ([]byte, error) {
	return ulid.ULID(id).MarshalText()
}

func (id ID) MarshalTextTo(dst []byte) error {
	return ulid.ULID(id).MarshalTextTo(dst)
}

func (id *ID) UnmarshalText(data []byte) error {
	ulid := &ulid.ULID{}
	e := ulid.UnmarshalText(data)
	if e != nil {
		return e
	}
	*id = ID(*ulid)
	return nil
}

func (id *ID) Scan(src interface{}) error {
	ulid := &ulid.ULID{}
	e := ulid.Scan(src)
	if e != nil {
		return e
	}
	*id = ID(*ulid)
	return nil
}

func (id ID) Value() (driver.Value, error) {
	return ulid.ULID(id).Value()
}

func (id ID) Compare(other ID) int {
	return ulid.ULID(id).Compare(ulid.ULID(other))
}

func (id ID) Equal(other ID) bool {
	return id.Compare(other) == 0
}

func (id ID) String() string {
	return ulid.ULID(id).String()
}

func (id ID) Copy() ID {
	copy := ulid.ULID{}
	for i, b := range id {
		copy[i] = b
	}
	return ID(copy)
}

func Now() time.Time {
	return time.Now().UTC()
}
