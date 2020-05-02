package main

import (
	. "github.com/0xor1/wtf/pkg/core"
)

func main() {
	id := NewIDGen().MustNew()
	bs, _ := id.MarshalBinary()
	Println(string(bs))
	bs, _ = NewIDGen().MustNew().MarshalText()
	Println(string(bs))
}
