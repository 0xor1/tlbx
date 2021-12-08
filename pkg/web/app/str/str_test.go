package str_test

import (
	"testing"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app/str"
	"github.com/stretchr/testify/assert"
)

func TestKey(t *testing.T) {
	a := assert.New(t)
	v := "0123456789_abcdefghijklmnopqrstuvwxyz"
	k := str.Key(v)

	bs, err := k.MarshalText()
	a.Nil(err)
	a.Equal(v, string(bs))
	a.Nil(k.MarshalTextTo(bs))
	a.Nil(err)
	a.Equal(v, string(bs))

	newK := str.Key("")
	err = newK.UnmarshalText(bs)
	a.Nil(err)
	a.Equal(v, string(newK))

	k = str.ToKey("   8  9  {}@#:asd   8 d  +){")
	a.Equal("8_9_asd_8_d", string(k))
	k = str.ToKey(string(k))
	a.Equal("8_9_asd_8_d", string(k))

	sqlV, err := k.Value()
	a.Nil(err)
	a.NotNil(sqlV)

	a.Nil(k.Scan(sqlV))
	a.Nil(k.Scan(string(sqlV.([]byte))))
	a.Equal("8_9_asd_8_d", k.String())

	tooLongKey := StrRepeat("1", 300)
	k = str.ToKey(tooLongKey)
	a.Len(k, 255)

	ks := str.Keys{k}
	a.Len(ks.ToIs(), 1)

	err = k.MarshalBinaryTo([]byte{})
	a.EqualError(err, "bad buffer size when marshaling")

	err = k.UnmarshalBinary([]byte{})
	a.Contains(err.Error(), `invalid str key detected: ""`)

	err = k.Scan([]byte{})
	a.Contains(err.Error(), `invalid str key detected: ""`)

	err = k.Scan("")
	a.Contains(err.Error(), `invalid str key detected: ""`)

	err = k.Scan(1)
	a.EqualError(err, `source value must be a string or byte slice`)

	k = str.Key("")
	_, err = k.Value()
	a.Contains(err.Error(), `invalid str key detected: ""`)

	defer Recover(func(i interface{}) {
		a.Contains(i.(Error).Error(), `key must not be a ulid string detected`)
	})
	str.ToKey(NewIDGen().MustNew().String())
}

func TestShort(t *testing.T) {
	a := assert.New(t)
	tooLong := StrRepeat("1", 300)
	v := "hi no"
	k := str.ToShort(v)

	bs, err := k.MarshalText()
	a.Nil(err)
	a.Equal(v, string(bs))
	a.Nil(k.MarshalTextTo(bs))
	a.Nil(err)
	a.Equal(v, string(bs))

	newK := str.Short("")
	err = newK.UnmarshalText(bs)
	a.Nil(err)
	a.Equal(v, string(newK))

	ss := str.Short(tooLong)
	sqlV, err := ss.Value()
	a.NotNil(err)
	a.Nil(sqlV)

	sqlV, err = k.Value()
	a.Nil(err)
	a.NotNil(sqlV)

	a.Nil(k.Scan(sqlV))
	a.Nil(k.Scan(string(sqlV.([]byte))))
	a.Equal("hi no", k.String())

	err = k.UnmarshalBinary([]byte(tooLong))
	a.Contains(err.Error(), Strf("invalid short string detected: %q", tooLong))

	err = k.MarshalBinaryTo([]byte{})
	a.EqualError(err, "bad buffer size when marshaling")

	err = k.Scan([]byte(tooLong))
	a.Contains(err.Error(), Strf(`invalid short string detected: %q`, tooLong))

	err = k.Scan(tooLong)
	a.Contains(err.Error(), Strf(`invalid short string detected: %q`, tooLong))

	err = k.Scan(1)
	a.EqualError(err, `source value must be a string or byte slice`)
}

func TestLong(t *testing.T) {
	a := assert.New(t)
	v := "hi no"
	k := str.ToLong(v)

	bs, err := k.MarshalText()
	a.Nil(err)
	a.Equal(v, string(bs))
	a.Nil(k.MarshalTextTo(bs))
	a.Nil(err)
	a.Equal(v, string(bs))

	newK := str.Long("")
	err = newK.UnmarshalText(bs)
	a.Nil(err)
	a.Equal(v, string(newK))

	sqlV, err := k.Value()
	a.Nil(err)
	a.NotNil(sqlV)

	a.Nil(k.Scan(sqlV))
	a.Nil(k.Scan(string(sqlV.([]byte))))
	a.Equal("hi no", k.String())

	err = k.MarshalBinaryTo([]byte{})
	a.EqualError(err, "bad buffer size when marshaling")

	err = k.Scan(1)
	a.EqualError(err, `source value must be a string or byte slice`)
}
