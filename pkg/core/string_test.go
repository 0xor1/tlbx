package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrEllipsis(t *testing.T) {
	a := assert.New(t)
	s := `0123456789`
	Do(func() {
		StrEllipsis(s, 3)
	}, func(r interface{}) {
		a.Contains(r.(Error).Message(), "str ellipsis max must be greater than 3")
	})
	StrEllipsis(s, 4)
	a.Equal("0...", StrEllipsis(s, 4))
}

func TestStrRepeat(t *testing.T) {
	a := assert.New(t)
	s := `ABC`
	a.Equal(`ABCABCABC`, StrRepeat(s, 3))
}

func TestStrReplaceAll(t *testing.T) {
	a := assert.New(t)
	s := `ABCB`
	a.Equal(`ADCD`, StrReplaceAll(s, `B`, `D`))
}

func TestStrReplace(t *testing.T) {
	a := assert.New(t)
	s := `ABCB`
	a.Equal(`ADCB`, StrReplace(s, `B`, `D`, 1))
}

func TestStrSplit(t *testing.T) {
	a := assert.New(t)
	s := `A,B,C`
	a.Equal([]string{`A`, `B`, `C`}, StrSplit(s, ","))
}

func TestStrLower(t *testing.T) {
	a := assert.New(t)
	s := `ABC`
	a.Equal(`abc`, StrLower(s))
}

func TestStrUpper(t *testing.T) {
	a := assert.New(t)
	s := `abc`
	a.Equal(`ABC`, StrUpper(s))
}

func TestStrTrim(t *testing.T) {
	a := assert.New(t)
	s := `$$abc$$`
	a.Equal(`abc`, StrTrim(s, "$$"))
}

func TestStrTrimWS(t *testing.T) {
	a := assert.New(t)
	s := ` abc     `
	a.Equal(`abc`, StrTrimWS(s))
}

func TestStrLen(t *testing.T) {
	a := assert.New(t)
	s := `平仮名, ひらがな`
	a.NotEqual(9, len(s))
	a.Equal(9, StrLen(s))
}

func TestErrorf(t *testing.T) {
	a := assert.New(t)
	a.Contains(Err("1 %d %q", 1, "1").Error(), "message: 1 1 \"1\"\nstackTrace")
	a.Contains(Err("1").Error(), "message: 1\nstackTrace")
}

func TestSprint(t *testing.T) {
	a := assert.New(t)
	a.Equal(`1`, Str("1"))
}

func TestSprintf(t *testing.T) {
	a := assert.New(t)
	a.Equal(`1 1 "1"`, Strf("1 %d %q", 1, "1"))
	a.Equal(`1`, Strf("1"))
}

func TestSprintln(t *testing.T) {
	a := assert.New(t)
	a.Equal("1\n", Strln("1"))
}

func TestPrintFuncs(t *testing.T) {
	Print("a")
	Printf("a")
	Println("a")
}
