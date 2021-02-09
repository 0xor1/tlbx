package core

import (
	"database/sql/driver"
	"fmt"
	"io"
	"math/rand"
	"sync"

	"github.com/oklog/ulid/v2"
)

var (
	seed    = NowUnixNano()
	seedMtx = &sync.Mutex{}
)

type IDGenPool interface {
	Get() IDGen
}

// n.b. each IDGen allocates approx 5kb of memory
// so use a fixed pool to save on lots of big memory
// de/allocations
func NewIDGenPool(size int) IDGenPool {
	PanicIf(size <= 1, "pool size must be greater than 1")
	pool := make([]IDGen, size)
	for i := 0; i < size; i++ {
		pool[i] = NewIDGen()
	}
	return &idGenPool{
		mtx:  &sync.Mutex{},
		i:    0,
		sm1:  size - 1,
		pool: pool,
	}
}

type idGenPool struct {
	mtx  *sync.Mutex
	i    int
	sm1  int
	pool []IDGen
}

func (p *idGenPool) Get() IDGen {
	p.mtx.Lock()
	if p.i == p.sm1 {
		p.i = 0
	} else {
		p.i++
	}
	p.mtx.Unlock()
	return p.pool[p.i]
}

type IDGen interface {
	New() (ID, error)
	MustNew() ID
}

func NewIDGen() IDGen {
	seedMtx.Lock()
	t := seed
	seed++
	seedMtx.Unlock()
	return &idGen{
		entropyMtx: &sync.Mutex{},
		entropy:    rand.New(rand.NewSource(t)),
	}
}

type idGen struct {
	entropyMtx *sync.Mutex
	entropy    io.Reader
}

func (g *idGen) New() (ID, error) {
	g.entropyMtx.Lock()
	defer g.entropyMtx.Unlock()
	id, e := ulid.New(ulid.Now(), g.entropy)
	return ID(id), e
}

func (g *idGen) MustNew() ID {
	id, err := g.New()
	PanicOn(err)
	return id
}

type ID ulid.ULID

func (id ID) IsZero() bool {
	return id.Equal(ID{})
}

func ParseID(id string) (ID, error) {
	i := &ID{}
	err := i.UnmarshalText([]byte(id))
	return *i, err
}

func MustParseID(id string) ID {
	i, err := ParseID(id)
	PanicOn(err)
	return i
}

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
	if id.IsZero() {
		return zeroIDErr
	}
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
	if id.IsZero() {
		return zeroIDErr
	}
	return nil
}

func (id *ID) Scan(src interface{}) error {
	ulid := &ulid.ULID{}
	e := ulid.Scan(src)
	if e != nil {
		return e
	}
	*id = ID(*ulid)
	if id.IsZero() {
		return zeroIDErr
	}
	return nil
}

func (id ID) Value() (driver.Value, error) {
	if id.IsZero() {
		return nil, zeroIDErr
	}
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

type IDs []ID

func (ids IDs) ToIs() []interface{} {
	res := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		res = append(res, id)
	}
	return res
}

func (ids IDs) StrJoin(sep string) string {
	strs := make([]string, 0, len(ids))
	for _, id := range ids {
		strs = append(strs, id.String())
	}
	return StrJoin(strs, sep)
}

// use fmt.Errorf as no stack trace here.
var zeroIDErr = fmt.Errorf("zero id detected")

func PanicIfZeroID(id ID) {
	// I cant think of a good reason why a nil value would ever
	// be the right thing to pass to an endpoint, it always means
	// the users has forgotten to pass a value.
	if id.IsZero() {
		PanicOn(zeroIDErr)
	}
}
