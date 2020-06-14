package ptr

import (
	"time"

	"github.com/0xor1/tlbx/pkg/core"
)

func ID(v core.ID) *core.ID {
	return &v
}

func Bool(v bool) *bool {
	return &v
}

func BoolOr(v *bool, or bool) bool {
	if v == nil {
		return or
	}
	return *v
}

func String(v string) *string {
	return &v
}

func StringOr(v *string, or string) string {
	if v == nil {
		return or
	}
	return *v
}

func Time(v time.Time) *time.Time {
	return &v
}

func TimeOr(v *time.Time, or time.Time) time.Time {
	if v == nil {
		return or
	}
	return *v
}

func Int(v int) *int {
	return &v
}

func IntOr(v *int, or int) int {
	if v == nil {
		return or
	}
	return *v
}

func Int64(v int64) *int64 {
	return &v
}

func Int64Or(v *int64, or int64) int64 {
	if v == nil {
		return or
	}
	return *v
}

func Uint(v uint) *uint {
	return &v
}

func UintOr(v *uint, or uint) uint {
	if v == nil {
		return or
	}
	return *v
}

func Uint64(v uint64) *uint64 {
	return &v
}

func Uint64Or(v *uint64, or uint64) uint64 {
	if v == nil {
		return or
	}
	return *v
}

func Float64(v float64) *float64 {
	return &v
}

func Float64Or(v *float64, or float64) float64 {
	if v == nil {
		return or
	}
	return *v
}
