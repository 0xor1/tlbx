package core

import (
	"database/sql/driver"
	"errors"
	"regexp"
)

var (
	keyValidRegex                 = regexp.MustCompile(`^[a-z][_a-z0-9]{0,48}[a-z0-9]?$`)
	invalidPrefixRegex            = regexp.MustCompile(`^[_0-9]+`)
	keyValidDoubleUnderscoreRegex = regexp.MustCompile(`__`)
	keyWhiteSpaceOrUnderscores    = regexp.MustCompile(`[\s_]+`)
	keyInvalidChar                = regexp.MustCompile(`[^a-z0-9_]+`)
	errUlidString                 = errors.New("key must not be a ulid string detected")
	errBufferSize                 = errors.New("bad buffer size when marshaling")
	errScanValue                  = errors.New("source value must be a string or byte slice")
)

func ToKey(s string) Key {
	if _, err := ParseID(s); err == nil {
		// not allowed to be a ulid string
		PanicOn(errUlidString)
	}

	// lower all chars
	s = StrLower(s)
	// replace all ws or underscore chars with a single _
	s = keyWhiteSpaceOrUnderscores.ReplaceAllString(s, `_`)
	// remove all invalid chars
	s = keyInvalidChar.ReplaceAllString(s, ``)
	// replace all ws or underscore chars with a single _ again incase the
	// removal of invalid chars created any double underscores
	s = keyWhiteSpaceOrUnderscores.ReplaceAllString(s, `_`)
	// cut invalid prefix
	s = invalidPrefixRegex.ReplaceAllString(s, ``)
	// trim any leading or trailing underscores
	s = StrTrim(s, `_`)
	PanicIf(len(s) == 0, "empty str key")
	if len(s) > 50 {
		s = s[:50]
	}
	return Key(s)
}

func ToKeyPtr(s string) *Key {
	k := ToKey(s)
	return &k
}

type Keys []Key

func (s Keys) ToIs() []interface{} {
	res := make([]interface{}, len(s))
	for i, k := range s {
		res[i] = k
	}
	return res
}

// keys are user defined ids, max chars 50
type Key string

func isValidKey(s string) bool {
	return keyValidRegex.MatchString(s) &&
		!keyValidDoubleUnderscoreRegex.MatchString(s)
}

func (s Key) MarshalBinary() ([]byte, error) {
	if !isValidKey(s.String()) {
		return nil, invalidStrKeyErr(s.String())
	}
	return []byte(s), nil
}

func (s Key) MarshalBinaryTo(dst []byte) error {
	if !isValidKey(s.String()) {
		return invalidStrKeyErr(s.String())
	}
	if len(s) > len(dst) {
		return errBufferSize
	}
	copy(dst, s)
	return nil
}

func (s *Key) UnmarshalBinary(data []byte) error {
	d := string(data)
	d = StrTrimWS(d)
	if !isValidKey(d) {
		return invalidStrKeyErr(d)
	}
	*s = Key(d)
	return nil
}

func (s Key) MarshalText() ([]byte, error) {
	return s.MarshalBinary()
}

func (s Key) MarshalTextTo(dst []byte) error {
	return s.MarshalBinaryTo(dst)
}

func (s *Key) UnmarshalText(data []byte) error {
	return s.UnmarshalBinary(data)
}

func (s *Key) Scan(src interface{}) error {
	switch x := src.(type) {
	case string:
		if !isValidKey(x) {
			return invalidStrKeyErr(string(x))
		}
		*s = Key(x)
	case []byte:
		if !isValidKey(string(x)) {
			return invalidStrKeyErr(string(x))
		}
		*s = Key(x)
	default:
		return errScanValue
	}
	return nil
}

func (s Key) Value() (driver.Value, error) {
	str := string(s)
	if !isValidKey(str) {
		return nil, invalidStrKeyErr(str)
	}
	return s.MarshalBinary()
}

func (s *Key) String() string {
	return string(*s)
}

func invalidStrKeyErr(s string) error {
	return Err("invalid str key detected: %q", s)
}
