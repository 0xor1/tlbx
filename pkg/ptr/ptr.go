package ptr

import (
	"github.com/0xor1/wtf/pkg/core"
	"time"
)

func ID(v core.ID) *core.ID {
	return &v
}

func Bool(v bool) *bool {
	return &v
}

func String(v string) *string {
	return &v
}

func Time(v time.Time) *time.Time {
	return &v
}

func Int(v int) *int {
	return &v
}

func Int64(v int64) *int64 {
	return &v
}

func Uint(v uint) *uint {
	return &v
}

func Uint64(v uint64) *uint64 {
	return &v
}
