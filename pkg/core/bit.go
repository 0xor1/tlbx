package core

type Bit uint8

func (b Bit) Bool() bool {
	return b > 0
}

func (b Bit) MarshalBinary() ([]byte, error) {
	if b > 0 {
		return []byte(`1`), nil
	}
	return []byte(`0`), nil
}

func (b *Bit) UnmarshalBinary(d []byte) error {
	strD := string(d)
	switch strD {
	case `1`:
		*b = 1
	case `0`:
		*b = 0
	default:
		return Errorf("invalid value %s, Bit only accepts 0 or 1", strD)
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
