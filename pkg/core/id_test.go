package core

import (
	"testing"

	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func Test_NewIDGenPool(t *testing.T) {
	pool := NewIDGenPool(2)
	MustGoGroup(func() {
		pool.Get().MustNew()
	}, func() {
		pool.Get().MustNew()
	}, func() {
		pool.Get().MustNew()
	}, func() {
		pool.Get().MustNew()
	}, func() {
		pool.Get().MustNew()
	}, func() {
		pool.Get().MustNew()
	}, func() {
		pool.Get().MustNew()
	})
}

func Test_IDGenerator(t *testing.T) {
	NewIDGen().MustNew()
}

func Test_ID(t *testing.T) {
	a := assert.New(t)
	gen := NewIDGen()
	id1 := gen.MustNew()
	bin1, err := id1.MarshalBinary()
	a.Nil(err)

	var id2 ID
	err = id2.UnmarshalBinary(nil)
	a.NotNil(err)
	err = id2.UnmarshalBinary(bin1)
	a.Nil(err)
	a.True(id1.Equal(id2))

	bin2 := make([]byte, 16)
	err = id2.MarshalBinaryTo(bin2)
	a.Nil(err)
	a.Equal(bin1, bin2)

	id2 = gen.MustNew()
	a.False(id1.Equal(id2))
	a.InDelta(NowMilli().Unix(), id2.Time().Unix(), 1)

	str1, err := id2.MarshalText()
	a.Nil(err)

	err = id1.UnmarshalText(nil)
	a.NotNil(err)
	err = id1.UnmarshalText(str1)
	a.Nil(err)
	a.True(id1.Equal(id2))

	str2 := make([]byte, ulid.EncodedSize)
	err = id2.MarshalTextTo(str2)
	a.Nil(err)
	a.Equal(str1, str2)

	MustParseID(id1.String())
	id1.Copy()

	err = id1.Scan([]byte{1})
	a.NotNil(err)
	err = id1.Scan(bin1)
	a.Nil(err)

	val, err := id1.Value()
	a.Nil(err)
	a.NotNil(val)

	a.Equal(1, len(IDs{id1}.ToIs()))
}

func Test_zeroIDErrs(t *testing.T) {
	a := assert.New(t)
	id := &ID{}
	bs, err := id.MarshalText()
	a.Nil(err)
	err = id.UnmarshalText(bs)
	a.Equal(zeroIDErr().Message(), err.(Error).Message())
	bs, err = id.MarshalBinary()
	a.Nil(err)
	err = id.UnmarshalBinary(bs)
	a.Equal(zeroIDErr().Message(), err.(Error).Message())
	err = id.Scan(bs)
	a.Equal(zeroIDErr().Message(), err.(Error).Message())
	_, err = id.Value()
	a.Equal(zeroIDErr().Message(), err.(Error).Message())
}

func Test_IDsStrJoin(t *testing.T) {
	a := assert.New(t)
	ids := IDs{ID{}, ID{}}
	idsStr := ids.StrJoin("_")
	a.Equal("00000000000000000000000000_00000000000000000000000000", idsStr)
}

func Test_IDsMerge(t *testing.T) {
	a := assert.New(t)
	gen := NewIDGen()
	u := gen.MustNew()
	v := gen.MustNew()
	w := gen.MustNew()
	x := gen.MustNew()
	y := gen.MustNew()
	z := gen.MustNew()
	one := IDs{w, x, y, z}
	two := IDs{u, v, w, x, y, z}
	merged := IDsMerge(one, two)
	a.Equal(merged[0], w)
	a.Equal(merged[1], x)
	a.Equal(merged[2], y)
	a.Equal(merged[3], z)
	a.Equal(merged[4], u)
	a.Equal(merged[5], v)
	a.Equal(6, len(merged))
}

func Test_PanicIfZeroID(t *testing.T) {
	a := assert.New(t)
	Do(func() {
		PanicIfZeroID(ID{})
	}, func(r interface{}) {
		a.Equal(r.(Error).Message(), zeroIDErr().Message())
	})
}
