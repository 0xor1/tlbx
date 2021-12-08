package str

import (
	"database/sql/driver"
	"errors"
	"regexp"

	. "github.com/0xor1/tlbx/pkg/core"
)

var (
	keyValidRegex                 = regexp.MustCompile(`^[a-z0-9][_a-z0-9]{0,253}[a-z0-9]?$`)
	keyValidDoubleUnderscoreRegex = regexp.MustCompile(`__`)
	keyWhiteSpaceOrUnderscores    = regexp.MustCompile(`[\s_]+`)
	keyInvalidChar                = regexp.MustCompile(`[^a-z0-9_]+`)
	errBufferSize                 = errors.New("bad buffer size when marshaling")
	errScanValue                  = errors.New("source value must be a string or byte slice")
)

func ToKey(s string) Key {
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
	if len(s) > 255 {
		s = s[:255]
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

// keys are user defined ids
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

func ToShort(s string) Short {
	sh := Short("")
	PanicOn(sh.UnmarshalBinary([]byte(s)))
	return sh
}

type Short string

func isValidShort(s string) bool {
	return s == "" || shortValidRegex.MatchString(s)
}

func (s Short) MarshalBinary() ([]byte, error) {
	return []byte(s), nil
}

func (s Short) MarshalBinaryTo(dst []byte) error {
	if len(s) > len(dst) {
		return errBufferSize
	}
	copy(dst, s)
	return nil
}

func (s *Short) UnmarshalBinary(data []byte) error {
	d := string(data)
	d = StrTrimWS(d)
	if !isValidShort(d) {
		return invalidShortErr(d)
	}
	*s = Short(d)
	return nil
}

func (s Short) MarshalText() ([]byte, error) {
	return s.MarshalBinary()
}

func (s Short) MarshalTextTo(dst []byte) error {
	return s.MarshalBinaryTo(dst)
}

func (s *Short) UnmarshalText(data []byte) error {
	return s.UnmarshalBinary(data)
}

func (s *Short) Scan(src interface{}) error {
	switch x := src.(type) {
	case string:
		if !isValidShort(x) {
			return invalidShortErr(x)
		}
		*s = Short(x)
	case []byte:
		if !isValidShort(string(x)) {
			return invalidShortErr(string(x))
		}
		*s = Short(x)
	default:
		return errScanValue
	}
	return nil
}

func (s Short) Value() (driver.Value, error) {
	str := string(s)
	if !isValidShort(str) {
		return nil, invalidShortErr(str)
	}
	return s.MarshalBinary()
}

func (s *Short) String() string {
	return string(*s)
}

func invalidShortErr(s string) Error {
	return Err("invalid short string detected: %q", s).(Error)
}

func ToLong(s string) Long {
	l := Long("")
	PanicOn(l.UnmarshalBinary([]byte(s)))
	return l
}

type Long string

func (s Long) MarshalBinary() ([]byte, error) {
	return []byte(s), nil
}

func (s Long) MarshalBinaryTo(dst []byte) error {
	if len(s) > len(dst) {
		return errBufferSize
	}
	copy(dst, s)
	return nil
}

func (s *Long) UnmarshalBinary(data []byte) error {
	d := string(data)
	d = StrTrimWS(d)
	*s = Long(d)
	return nil
}

func (s Long) MarshalText() ([]byte, error) {
	return s.MarshalBinary()
}

func (s Long) MarshalTextTo(dst []byte) error {
	return s.MarshalBinaryTo(dst)
}

func (s *Long) UnmarshalText(data []byte) error {
	return s.UnmarshalBinary(data)
}

func (s *Long) Scan(src interface{}) error {
	switch x := src.(type) {
	case string:
		*s = Long(x)
	case []byte:
		*s = Long(x)
	default:
		return errScanValue
	}
	return nil
}

func (s Long) Value() (driver.Value, error) {
	return s.MarshalBinary()
}

func (s *Long) String() string {
	return string(*s)
}

var (
	emailValidRegex = regexp.MustCompile(`\A\S+@\S+\.\S+\z`)
)

func ToEmail(s string) Email {
	e := Email("")
	PanicOn(e.UnmarshalBinary([]byte(s)))
	return e
}

type EmailField struct {
	V Email `json:"v"`
}

type EmailPtrField struct {
	V *Email `json:"v"`
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

type PwdField struct {
	V Pwd `json:"v"`
}

type PwdPtrField struct {
	V *Pwd `json:"v"`
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
