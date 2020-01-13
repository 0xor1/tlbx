package core

import (
	"fmt"
	"unicode/utf8"
)

func StrLen(s string) int {
	return utf8.RuneCountInString(s)
}

func Errorf(f string, args ...interface{}) error {
	return fmt.Errorf(f, args...)
}

func Sprint(args ...interface{}) string {
	return fmt.Sprint(args...)
}

func Sprintf(f string, args ...interface{}) string {
	return fmt.Sprintf(f, args...)
}

func Sprintln(args ...interface{}) string {
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
