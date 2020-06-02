package blockers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Bit(t *testing.T) {
	a := assert.New(t)

	p := Pbit(0)
	p.MarshalText()
	p = Pbit(1)
	p.MarshalText()
	p = Pbit(2)
	p.MarshalText()
	p = Pbit(3)
	p.MarshalText()
	p = Pbit(4)
	p.MarshalText()
	p.UnmarshalText([]byte(`1`))

	pBs, err := p.MarshalJSON()
	a.Nil(err)
	a.Equal(`1`, string(pBs))

	p = Pbit(0)
	pBs, err = p.MarshalJSON()
	a.Nil(err)
	a.Equal(`0`, string(pBs))

	a.Nil(p.UnmarshalJSON([]byte(`0`)))
	a.Nil(p.UnmarshalJSON([]byte(`1`)))
	a.Nil(p.UnmarshalJSON([]byte(`2`)))
	a.Nil(p.UnmarshalJSON([]byte(`3`)))
	a.Nil(p.UnmarshalJSON([]byte(`4`)))

	_, err = Pbit(5).MarshalJSON()
	a.Equal(`invalid value 5, Pbit only accepts 0, 1, 2, 3 or 4`, err.Error())
	a.Equal(`invalid value 5, Pbit only accepts 0, 1, 2, 3 or 4`, p.UnmarshalJSON([]byte(`5`)).Error())

	ps := Pbits{0, 1, 2, 3, 4}

	ps.MarshalText()
	ps.UnmarshalText([]byte(`01234`))

	psBs, err := ps.MarshalText()
	a.Nil(err)
	a.Equal(`01234`, string(psBs))

	a.Nil(ps.UnmarshalText([]byte(`012343210`)))

	_, err = Pbits{0, 1, 2, 3, 4, 5, 4, 3, 2, 1, 0}.MarshalText()
	a.Equal(`invalid value 5, Pbits only accepts 0s, 1s, 2s, 3s and 4s`, err.Error())
	a.Equal(`invalid value 5, Pbits only accepts 0s, 1s, 2s, 3s and 4s`, ps.UnmarshalText([]byte(`01234543210`)).Error())
}
