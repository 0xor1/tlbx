package crypt

import (
	"crypto/rand"
	"io"
	"math/big"

	. "github.com/0xor1/tlbx/pkg/core"
	"golang.org/x/crypto/scrypt"
)

var urlSafeRunes = []rune("0123456789_abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func Bytes(length int) []byte {
	k := make([]byte, length)
	_, err := io.ReadFull(rand.Reader, k)
	PanicOn(err)
	return k
}

func UrlSafeString(length int) string {
	buf := make([]rune, length)
	urlSafeRunesLength := big.NewInt(int64(len(urlSafeRunes)))
	for i := range buf {
		randomIdx, err := rand.Int(rand.Reader, urlSafeRunesLength)
		PanicOn(err)
		buf[i] = urlSafeRunes[int(randomIdx.Int64())]
	}
	return string(buf)
}

func ScryptKey(password, salt []byte, N, r, p, keyLen int) []byte {
	key, err := scrypt.Key(password, salt, N, r, p, keyLen)
	PanicOn(err)
	return key
}
