package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKey(t *testing.T) {
	a := assert.New(t)
	defer Recover(func(i interface{}) {
		a.Contains(i.(Error).Error(), `invalid key detected`)
	})
	MustParseKey("yolo")
	MustParseKey("_yolo_")
}
func TestKey(t *testing.T) {
	a := assert.New(t)
	v := "abcdefghijklmnopqrstuvwxyz_0123456789"
	k := Key(v)

	bs, err := k.MarshalText()
	a.Nil(err)
	a.Equal(v, string(bs))
	a.Nil(k.MarshalTextTo(bs))
	a.Nil(err)
	a.Equal(v, string(bs))

	newK := Key("")
	err = newK.UnmarshalText(bs)
	a.Nil(err)
	a.Equal(v, string(newK))

	k = MustToKey(" f   8  9  {}@#:asd   8 d  +){")
	a.Equal("f_8_9_asd_8_d", string(k))
	k = MustToKey(string(k))
	a.Equal("f_8_9_asd_8_d", string(k))

	sqlV, err := k.Value()
	a.Nil(err)
	a.NotNil(sqlV)

	a.Nil(k.Scan(sqlV))
	a.Nil(k.Scan(string(sqlV.([]byte))))
	a.Equal("f_8_9_asd_8_d", k.String())

	tooLongKey := StrRepeat("f", 101)
	k = MustToKey(tooLongKey)
	a.Len(k, 50)

	ks := Keys{k}
	a.Len(ks.ToIs(), 1)

	err = k.MarshalBinaryTo([]byte{})
	a.EqualError(err, "bad buffer size when marshaling")

	err = k.UnmarshalBinary([]byte{})
	a.Contains(err.Error(), `invalid key detected: ""`)

	err = k.Scan([]byte{})
	a.Contains(err.Error(), `invalid key detected: ""`)

	err = k.Scan("")
	a.Contains(err.Error(), `invalid key detected: ""`)

	err = k.Scan(1)
	a.EqualError(err, `source value must be a string or byte slice`)

	k = Key("")
	_, err = k.Value()
	a.Contains(err.Error(), `invalid key detected: ""`)

	defer Recover(func(i interface{}) {
		a.Contains(i.(Error).Error(), `key must not be a ulid string detected`)
	})
	MustToKey(NewIDGen().MustNew().String())
}

func Test_KeysValueAndScan(t *testing.T) {
	a := assert.New(t)
	ks := Keys{"yolo", "nolo"}
	v, e := ks.Value()
	a.Nil(e)
	a.Len(v, 9)
	scannedKs := Keys{}
	a.Nil(scannedKs.Scan(v))
	a.Equal(ks, scannedKs)
	a.Len(scannedKs, 2)
	a.Nil(scannedKs.Scan(nil))
	a.Len(scannedKs, 0)
}

func Test_ParseKeys(t *testing.T) {
	a := assert.New(t)
	ks, e := ParseKeys([]string{"yolo", " nolo"})
	a.Nil(ks)
	a.NotNil(e)

	ks = MustParseKeys([]string{"yolo", "nolo"})
	a.Len(ks, 2)

	ks, e = ToKeys([]string{"yolo", " nolo"})
	a.Len(ks, 2)
	a.Nil(e)
}
