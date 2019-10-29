package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_Now(t *testing.T) {
	a := assert.New(t)
	a.Equal(time.Now().UTC().Round(time.Second).Unix(), Now().Round(time.Second).Unix())
}

func Test_NowUnixNano(t *testing.T) {
	a := assert.New(t)
	a.InDelta(time.Now().UnixNano(), NowUnixNano(), 1000)
}

func Test_NowUnixMilli(t *testing.T) {
	a := assert.New(t)
	a.Equal(time.Now().UnixNano()/1000000, NowUnixMilli())
}
