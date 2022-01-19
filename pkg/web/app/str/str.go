package str

import (
	"database/sql/driver"
	"errors"
	"regexp"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/web/app/validate"
)

var (
	keyValidRegex                 = regexp.MustCompile(`^[a-z][_a-z0-9]{0,48}[a-z0-9]?$`)
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
	// trim any leading or trailing underscores
	s = StrTrim(s, `_`)
	PanicIf(len(s) == 0, "empty str key")
	if len(s) > 50 {
		s = s[:50]
	}
	return Key(s)
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
	return []byte(s), nil
}

func (s Key) MarshalBinaryTo(dst []byte) error {
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

func invalidStrKeyErr(s string) Error {
	return Err("invalid str key detected: %q", s).(Error)
}

var (
	shortValidRegex = regexp.MustCompile(`\A\S.{0,253}\S?\z`)
)

func ToShort(s string) String {
	sh := String("")
	PanicOn(sh.UnmarshalBinary([]byte(s)))
	return sh
}

type String string

func (s String) MarshalBinary() ([]byte, error) {
	return []byte(s), nil
}

func (s String) MarshalBinaryTo(dst []byte) error {
	if len(s) > len(dst) {
		return errBufferSize
	}
	copy(dst, s)
	return nil
}

func (s *String) UnmarshalBinary(data []byte) error {
	d := string(data)
	d = StrTrimWS(d)
	*s = String(d)
	return nil
}

func (s String) MarshalText() ([]byte, error) {
	return s.MarshalBinary()
}

func (s String) MarshalTextTo(dst []byte) error {
	return s.MarshalBinaryTo(dst)
}

func (s *String) UnmarshalText(data []byte) error {
	return s.UnmarshalBinary(data)
}

func (s *String) Scan(src interface{}) error {
	switch x := src.(type) {
	case string:
		*s = String(x)
	case []byte:
		*s = String(x)
	default:
		return errScanValue
	}
	return nil
}

func (s String) Value() (driver.Value, error) {
	return s.MarshalBinary()
}

func (s *String) String() string {
	return string(*s)
}

func (s *String) MustBeValid(name string, min, max int, regexs ...*regexp.Regexp) {
	validate.Str(name, s.String(), min, max, regexs...)
}

var (
	emailValidRegex = regexp.MustCompile(`\A\S+@\S+\.\S+\z`)
)

func ToEmail(s string) Email {
	e := Email("")
	PanicOn(e.UnmarshalBinary([]byte(s)))
	return e
}

type Email string

func isValidEmail(s string) bool {
	return emailValidRegex.MatchString(s)
}

func (s Email) MarshalBinary() ([]byte, error) {
	return []byte(s), nil
}

func (s Email) MarshalBinaryTo(dst []byte) error {
	if len(s) > len(dst) {
		return errBufferSize
	}
	copy(dst, s)
	return nil
}

func (s *Email) UnmarshalBinary(data []byte) error {
	d := string(data)
	d = StrTrimWS(d)
	if !isValidEmail(d) {
		return invalidEmailErr(d)
	}
	*s = Email(d)
	return nil
}

func (s Email) MarshalText() ([]byte, error) {
	return s.MarshalBinary()
}

func (s Email) MarshalTextTo(dst []byte) error {
	return s.MarshalBinaryTo(dst)
}

func (s *Email) UnmarshalText(data []byte) error {
	return s.UnmarshalBinary(data)
}

func (s *Email) Scan(src interface{}) error {
	switch x := src.(type) {
	case string:
		if !isValidEmail(x) {
			return invalidEmailErr(string(x))
		}
		*s = Email(x)
	case []byte:
		if !isValidEmail(string(x)) {
			return invalidEmailErr(string(x))
		}
		*s = Email(x)
	default:
		return errScanValue
	}
	return nil
}

func (s Email) Value() (driver.Value, error) {
	str := string(s)
	if !isValidEmail(str) {
		return nil, invalidEmailErr(str)
	}
	return s.MarshalBinary()
}

func (s *Email) String() string {
	return string(*s)
}

func invalidEmailErr(s string) Error {
	return Err("invalid email detected: %q", s).(Error)
}

var (
	pwdRegexs = []*regexp.Regexp{
		regexp.MustCompile(`[0-9]`),
		regexp.MustCompile(`[a-z]`),
		regexp.MustCompile(`[A-Z]`),
		regexp.MustCompile(`[\w]`),
	}
	pwdMinLen = 8
	pwdMaxLen = 100
)

func ToPwd(s string) Pwd {
	p := Pwd("")
	PanicOn(p.UnmarshalBinary([]byte(s)))
	return p
}

type Pwd string

func isValidPwd(s string) bool {
	l := StrLen(s)
	if l < pwdMinLen || l > pwdMaxLen {
		return false
	}
	for _, re := range pwdRegexs {
		if !re.MatchString(s) {
			return false
		}
	}
	return true
}

func (s Pwd) MarshalBinary() ([]byte, error) {
	return []byte(s), nil
}

func (s Pwd) MarshalBinaryTo(dst []byte) error {
	if len(s) > len(dst) {
		return errBufferSize
	}
	copy(dst, s)
	return nil
}

func (s *Pwd) UnmarshalBinary(data []byte) error {
	d := string(data)
	d = StrTrimWS(d)
	if !isValidPwd(d) {
		return invalidPwdErr(d)
	}
	*s = Pwd(d)
	return nil
}

func (s Pwd) MarshalText() ([]byte, error) {
	return s.MarshalBinary()
}

func (s Pwd) MarshalTextTo(dst []byte) error {
	return s.MarshalBinaryTo(dst)
}

func (s *Pwd) UnmarshalText(data []byte) error {
	return s.UnmarshalBinary(data)
}

func (s *Pwd) Scan(src interface{}) error {
	switch x := src.(type) {
	case string:
		if !isValidPwd(x) {
			return invalidPwdErr(string(x))
		}
		*s = Pwd(x)
	case []byte:
		if !isValidPwd(string(x)) {
			return invalidPwdErr(string(x))
		}
		*s = Pwd(x)
	default:
		return errScanValue
	}
	return nil
}

func (s Pwd) Value() (driver.Value, error) {
	str := string(s)
	if !isValidPwd(str) {
		return nil, invalidPwdErr(str)
	}
	return s.MarshalBinary()
}

func (s *Pwd) String() string {
	return string(*s)
}

func invalidPwdErr(s string) Error {
	return Err("invalid pwd detected: %q", s).(Error)
}
