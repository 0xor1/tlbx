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

func TestSprintf(t *testing.T) {
	a := assert.New(t)
	a.Equal(`1 1 "1"`, Sprintf("1 %d %q", 1, "1"))
	a.Equal(`1`, Sprintf("1"))
}
