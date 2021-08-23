package core

import (
	"database/sql/driver"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/oklog/ulid/v2"
)

func StrEllipsis(s string, max uint) string {
	PanicIf(max < 4, "str ellipsis max must be greater than 3")
	runes := []rune(s)
	if len(runes) > int(max) {
		runes = runes[:max-3]
		s = string(runes) + "..."
	}
	return s
}

func StrReplaceAll(s, old, new string) string {
	return StrReplace(s, old, new, -1)
}

func StrReplace(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

func StrRepeat(s string, count int) string {
	return strings.Repeat(s, count)
}

func StrSplit(s string, sep string) []string {
	return strings.Split(s, sep)
}

func StrJoin(strs []string, sep string) string {
	return strings.Join(strs, sep)
}

func StrLower(s string) string {
	return strings.ToLower(s)
}

func StrUpper(s string) string {
	return strings.ToUpper(s)
}

func StrTrim(s, cutset string) string {
	return strings.Trim(s, cutset)
}

func StrTrimWS(s string) string {
	return strings.TrimSpace(s)
}

func StrLen(s string) int {
	return utf8.RuneCountInString(s)
}

func StrHasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func StrHasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func StrContains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func Err(f string, args ...interface{}) error {
	// return Error for stacktrace
	return ToError(fmt.Errorf(f, args...))
}

func Str(args ...interface{}) string {
	return fmt.Sprint(args...)
}

func Strf(f string, args ...interface{}) string {
	return fmt.Sprintf(f, args...)
}

func Strln(args ...interface{}) string {
	return fmt.Sprintln(args...)
}

func Print(args ...interface{}) {
	fmt.Print(args...)
}

func Printf(f string, args ...interface{}) {
	fmt.Printf(f, args...)
}

func Println(args ...interface{}) {
	fmt.Println(args...)
}

var strKeyValidRegex = regexp.MustCompile(`^[a-z0-9][_a-z0-9]{0,253}[a-z0-9]?$`)
var strKeyValidDoubleUnderscoreRegex = regexp.MustCompile(`__`)
var strKeyWhiteSpaceOrUnderscores = regexp.MustCompile(`[\s_]+`)
var strKeyInvalidChar = regexp.MustCompile(`[^a-z0-9_]+`)

func StrKeyMustConvert(s string) StrKey {
	// lower all chars
	s = StrLower(s)
	// replace all ws or underscore chars with a single _
	s = strKeyWhiteSpaceOrUnderscores.ReplaceAllString(s, `_`)
	// remove all invalid chars
	s = strKeyInvalidChar.ReplaceAllString(s, ``)
	// replace all ws or underscore chars with a single _ again incase the
	// removal of invalid chars created any double underscores
	s = strKeyWhiteSpaceOrUnderscores.ReplaceAllString(s, `_`)
	// trim any leading or trailing underscores
	s = StrTrim(s, `_`)
	PanicIf(len(s) == 0, "empty str key")
	if len(s) > 255 {
		s = s[:256]
	}
	return StrKey(s)
}

func isValidStrKey(s string) bool {
	return strKeyValidRegex.MatchString(s) &&
		!strKeyValidDoubleUnderscoreRegex.MatchString(s)
}

// string keys are user defined ids
type StrKey string

func (s StrKey) MarshalBinary() ([]byte, error) {
	return []byte(s), nil
}

func (s StrKey) MarshalBinaryTo(dst []byte) error {
	if len(s) > len(dst) {
		return ulid.ErrBufferSize
	}
	copy(dst, s)
	return nil
}

func (s *StrKey) UnmarshalBinary(data []byte) error {
	if !isValidStrKey(string(data)) {
		return invalidStrKeyErr(string(data))
	}
	*s = StrKey(data)
	return nil
}

func (s StrKey) MarshalText() ([]byte, error) {
	return s.MarshalBinary()
}

func (s StrKey) MarshalTextTo(dst []byte) error {
	return s.MarshalBinaryTo(dst)
}

func (s *StrKey) UnmarshalText(data []byte) error {
	return s.UnmarshalBinary(data)
}

func (s *StrKey) Scan(src interface{}) error {
	switch x := src.(type) {
	case string:
		if !isValidStrKey(x) {
			return invalidStrKeyErr(string(x))
		}
		*s = StrKey(x)
	case []byte:
		if !isValidStrKey(string(x)) {
			return invalidStrKeyErr(string(x))
		}
		*s = StrKey(x)
	default:
		return ulid.ErrScanValue
	}
	return nil
}

func (s StrKey) Value() (driver.Value, error) {
	str := string(s)
	if !isValidStrKey(str) {
		return nil, invalidStrKeyErr(str)
	}
	return s.MarshalBinary()
}

func (s *StrKey) String() string {
	return string(*s)
}

func invalidStrKeyErr(s string) Error {
	return Err("invalid str key detected: %q", s).(Error)
}
