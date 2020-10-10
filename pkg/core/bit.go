package core

import "bytes"

type Bit uint8

func (b Bit) Bool() bool {
	PanicIf(b > 1, "Invalid bit value %d", b)
	return b == 1
}

func (b Bit) MarshalBinary() ([]byte, error) {
	switch b {
	case 0:
		return []byte(`0`), nil
	case 1:
		return []byte(`1`), nil
	default:
		return nil, Err("invalid value %d, Bit only accepts 0 or 1", b)
	}
}

func (b *Bit) UnmarshalBinary(d []byte) error {
	strD := string(d)
	switch strD {
	case `0`:
		*b = 0
	case `1`:
		*b = 1
	default:
		return Err("invalid value %s, Bit only accepts 0 or 1", strD)
	}
	return nil
}

func (b Bit) MarshalText() ([]byte, error) {
	return b.MarshalBinary()
}

func (b *Bit) UnmarshalText(d []byte) error {
	return b.UnmarshalBinary(d)
}

func (b Bit) MarshalJSON() ([]byte, error) {
	return b.MarshalBinary()
}

func (b *Bit) UnmarshalJSON(d []byte) error {
	return b.UnmarshalBinary(d)
}

type Bits []Bit

func (bs Bits) MarshalBinary() ([]byte, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(bs)))
	for _, b := range bs {
		switch b {
		case 0:
			buf.WriteString(`0`)
		case 1:
			buf.WriteString(`1`)
		default:
			return nil, Err("invalid value %d, Bits only accepts 0s and 1s", b)
		}
	}
	return buf.Bytes(), nil
}

func (bs *Bits) UnmarshalBinary(ds []byte) error {
	res := make([]Bit, 0, len(ds))
	for _, d := range ds {
		dStr := string(d)
		switch dStr {
		case `0`:
			res = append(res, 0)
		case `1`:
			res = append(res, 1)
		default:
			return Err("invalid value %s, Bits only accepts 0s and 1s", dStr)
		}
	}
	*bs = res
	return nil
}

func (bs Bits) MarshalText() ([]byte, error) {
	return bs.MarshalBinary()
}

func (bs *Bits) UnmarshalText(d []byte) error {
	return bs.UnmarshalBinary(d)
}
