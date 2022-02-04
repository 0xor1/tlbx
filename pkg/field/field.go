package field

import (
	"time"

	"github.com/0xor1/tlbx/pkg/core"
)

type ID struct {
	V core.ID `json:"v"`
}

type IDPtr struct {
	V *core.ID `json:"v"`
}

type Bool struct {
	V bool `json:"v"`
}

type BoolPtr struct {
	V *bool `json:"v"`
}

type String struct {
	V string `json:"v"`
}

type StringPtr struct {
	V *string `json:"v"`
}

type Time struct {
	V time.Time `json:"v"`
}

type TimePtr struct {
	V *time.Time `json:"v"`
}

type Int struct {
	V int `json:"v"`
}

type IntPtr struct {
	V *int `json:"v"`
}

type Int8 struct {
	V int8 `json:"v"`
}

type Int8Ptr struct {
	V *int8 `json:"v"`
}

type Int16 struct {
	V int16 `json:"v"`
}

type Int16Ptr struct {
	V *int16 `json:"v"`
}

type Int32 struct {
	V int32 `json:"v"`
}

type Int32Ptr struct {
	V *int32 `json:"v"`
}

type Int64 struct {
	V int64 `json:"v"`
}

type Int64Ptr struct {
	V *int64 `json:"v"`
}

type UInt struct {
	V uint `json:"v"`
}

type UIntPtr struct {
	V *uint `json:"v"`
}

type UInt8 struct {
	V uint8 `json:"v"`
}

type UInt8Ptr struct {
	V *uint8 `json:"v"`
}

type UInt16 struct {
	V uint16 `json:"v"`
}

type UInt16Ptr struct {
	V *uint16 `json:"v"`
}

type UInt32 struct {
	V uint32 `json:"v"`
}

type UInt32Ptr struct {
	V *uint32 `json:"v"`
}

type UInt64 struct {
	V uint64 `json:"v"`
}

type UInt64Ptr struct {
	V *uint64 `json:"v"`
}

type Float32 struct {
	V float32 `json:"v"`
}

type Float32Ptr struct {
	V *float32 `json:"v"`
}

type Float64 struct {
	V float64 `json:"v"`
}

type Float64Ptr struct {
	V *float64 `json:"v"`
}

type Key struct {
	V core.Key `json:"v"`
}

type KeyPtr struct {
	V *core.Key `json:"v"`
}
