package main

import (
	"encoding/base64"
	"flag"
	"os"

	. "github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/crypt"
)

func main() {
	fs := flag.NewFlagSet("tlbxcrypt", flag.ExitOnError)
	var t string
	fs.StringVar(&t, "t", "b", "b for url base64 encoded bytes array or s for ASCII string")
	var nTmp uint
	fs.UintVar(&nTmp, "n", 1, "number of crypt bytes or ASCII characters to generate")
	var lTmp uint
	fs.UintVar(&lTmp, "l", 64, "length of each crypt byte array or ASCII string")
	PanicOn(fs.Parse(os.Args[1:]))
	n := int(nTmp)
	l := int(lTmp)
	if t == "s" {
		for i := 0; i < n; i++ {
			Println(crypt.UrlSafeString(l))
		}
	} else {
		for i := 0; i < n; i++ {
			Println(Sprintf("%s", base64.RawURLEncoding.EncodeToString(crypt.Bytes(l))))
		}
	}
}
