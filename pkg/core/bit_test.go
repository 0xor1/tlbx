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

	a.Equal(`invalid value yolo, Bit only accepts 0 or 1`, b.UnmarshalJSON([]byte(`yolo`)).Error())
}
