package crypt

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Bytes(t *testing.T) {
	l := 3
	bs := Bytes(l)
	assert.Equal(t, l, len(bs))
	l = 5
	bs = Bytes(l)
	assert.Equal(t, l, len(bs))
}

func Test_UrlSafeString(t *testing.T) {
	l := 3
	bs := UrlSafeString(l)
	assert.Equal(t, l, len(bs))
	l = 5
	bs = UrlSafeString(l)
	assert.Equal(t, l, len(bs))
}

func Test_ScryptKey(t *testing.T) {
	l := 4
	pwd := Bytes(l)
	salt := Bytes(l)
	scryptPwd := ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
	l = 8
	pwd = Bytes(l)
	salt = Bytes(l)
	scryptPwd = ScryptKey(pwd, salt, l, l, l, l)
	assert.Equal(t, l, len(scryptPwd))
}
