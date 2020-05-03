package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Bit(t *testing.T) {
	a := assert.New(t)

	b := Bit(1)

	b.MarshalText()
	b.UnmarshalText([]byte(`1`))

	bBs, err := b.MarshalJSON()
	a.Nil(err)
	a.Equal(`1`, string(bBs))

	b = Bit(0)
	bBs, err = b.MarshalJSON()
	a.Nil(err)
	a.Equal(`0`, string(bBs))

	a.Nil(b.UnmarshalJSON([]byte(`1`)))
	a.True(b.Bool())

	a.Nil(b.UnmarshalJSON([]byte(`0`)))
	a.False(b.Bool())

	_, err = Bit(2).MarshalJSON()
	a.Equal(`invalid value 2, Bit only accepts 0 or 1`, err.Error())
	a.Equal(`invalid value 2, Bit only accepts 0 or 1`, b.UnmarshalJSON([]byte(`2`)).Error())

	bs := Bits{0, 1, 1, 0}

	bs.MarshalText()
	bs.UnmarshalText([]byte(`0110`))

	bsBs, err := bs.MarshalText()
	a.Nil(err)
	a.Equal(`0110`, string(bsBs))

	a.Nil(bs.UnmarshalText([]byte(`0110`)))
	_, err = Bits{0, 1, 2, 1, 0}.MarshalText()
	a.Equal(`invalid value 2, Bits only accepts 0s and 1s`, err.Error())
	a.Equal(`invalid value 2, Bits only accepts 0s and 1s`, bs.UnmarshalText([]byte(`01210`)).Error())
}
