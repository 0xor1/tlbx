package pbit

import (
	"bytes"

	. "github.com/0xor1/tlbx/pkg/core"
)

type Pbit uint8

func (p Pbit) MarshalBinary() ([]byte, error) {
	switch p {
	case 0:
		return []byte(`0`), nil
	case 1:
		return []byte(`1`), nil
	case 2:
		return []byte(`2`), nil
	case 3:
		return []byte(`3`), nil
	case 4:
		return []byte(`4`), nil
	default:
		return nil, Errorf("invalid value %d, Pbit only accepts 0, 1, 2, 3 or 4", p)
	}
}

func (p *Pbit) UnmarshalBinary(d []byte) error {
	strD := string(d)
	switch strD {
	case `0`:
		*p = 0
	case `1`:
		*p = 1
	case `2`:
		*p = 2
	case `3`:
		*p = 3
	case `4`:
		*p = 4
	default:
		return Errorf("invalid value %s, Pbit only accepts 0, 1, 2, 3 or 4", strD)
	}
	return nil
}

func (q Pbit) MarshalText() ([]byte, error) {
	return q.MarshalBinary()
}

func (q *Pbit) UnmarshalText(d []byte) error {
	return q.UnmarshalBinary(d)
}

func (q Pbit) MarshalJSON() ([]byte, error) {
	return q.MarshalBinary()
}

func (q *Pbit) UnmarshalJSON(d []byte) error {
	return q.UnmarshalBinary(d)
}

type Pbits []Pbit

func (ps Pbits) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(ps)))
	for _, p := range ps {
		switch p {
		case 0:
			buf.WriteString(`0`)
		case 1:
			buf.WriteString(`1`)
		case 2:
			buf.WriteString(`2`)
		case 3:
			buf.WriteString(`3`)
		case 4:
			buf.WriteString(`4`)
		default:
			return nil, Errorf("invalid value %d, Pbits only accepts 0s, 1s, 2s, 3s and 4s", p)
		}
	}
	return buf.Bytes(), nil
}

func (ps *Pbits) UnmarshalBinary(ds []byte) error {
	res := make([]Pbit, 0, len(ds))
	for _, d := range ds {
		dStr := string(d)
		switch dStr {
		case `0`:
			res = append(res, 0)
		case `1`:
			res = append(res, 1)
		case `2`:
			res = append(res, 2)
		case `3`:
			res = append(res, 3)
		case `4`:
			res = append(res, 4)
		default:
			return Errorf("invalid value %s, Pbits only accepts 0s, 1s, 2s, 3s and 4s", dStr)
		}
	}
	*ps = res
	return nil
}

func (ps Pbits) MarshalText() ([]byte, error) {
	return ps.MarshalBinary()
}

func (ps *Pbits) UnmarshalText(d []byte) error {
	return ps.UnmarshalBinary(d)
}
