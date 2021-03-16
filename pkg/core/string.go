package core

import (
	"fmt"
	"strings"
	"unicode/utf8"
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
