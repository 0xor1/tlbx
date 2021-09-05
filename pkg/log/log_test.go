package log

import (
	"testing"

	. "github.com/0xor1/tlbx/pkg/core"
)

func TestLog(t *testing.T) {
	l := New(func(e *Entry) {
		Println(e)
	})
	l.Debug("yolo")
	l.Stats(1)
	l.FatalOn(nil)
}
