package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrLen(t *testing.T) {
	a := assert.New(t)
	s := `平仮名, ひらがな`
	a.NotEqual(9, len(s))
	a.Equal(9, StrLen(s))
}

func TestErrorf(t *testing.T) {
	a := assert.New(t)
	a.Contains(Errorf("1 %d %q", 1, "1").Error(), "message: 1 1 \"1\"\nstackTrace")
	a.Contains(Errorf("1").Error(), "message: 1\nstackTrace")
}

func TestSprint(t *testing.T) {
	a := assert.New(t)
	a.Equal(`1`, Sprint("1"))
}

func TestSprintf(t *testing.T) {
	a := assert.New(t)
	a.Equal(`1 1 "1"`, Sprintf("1 %d %q", 1, "1"))
	a.Equal(`1`, Sprintf("1"))
}

func TestSprintln(t *testing.T) {
	a := assert.New(t)
	a.Equal("1\n", Sprintln("1"))
}

func TestPrintFuncs(t *testing.T) {
	Print("a")
	Printf("a")
	Println("a")
}
