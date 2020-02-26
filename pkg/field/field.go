package field

import (
	. "github.com/0xor1/wtf/pkg/core"
	"time"
)

type Id struct {
	Val ID `json:"val"`
}

type IdPtr struct {
	Val *ID `json:"val"`
}

type Bool struct {
	Val bool `json:"val"`
}

type BoolPtr struct {
	Val *bool `json:"val"`
}

type String struct {
	Val string `json:"val"`
}

type StringPtr struct {
	Val *string `json:"val"`
}

type Time struct {
	Val time.Time `json:"val"`
}

type TimePtr struct {
	Val *time.Time `json:"val"`
}

type Int struct {
	Val int `json:"val"`
}

type IntPtr struct {
	Val *int `json:"val"`
}

type Int8 struct {
	Val int8 `json:"val"`
}

type Int8Ptr struct {
	Val *int8 `json:"val"`
}

type Int16 struct {
	Val int16 `json:"val"`
}

type Int16Ptr struct {
	Val *int16 `json:"val"`
}

type Int32 struct {
	Val int32 `json:"val"`
}

type Int32Ptr struct {
	Val *int32 `json:"val"`
}

type Int64 struct {
	Val int64 `json:"val"`
}

type Int64Ptr struct {
	Val *int64 `json:"val"`
}

type UInt struct {
	Val uint `json:"val"`
}

type UIntPtr struct {
	Val *uint `json:"val"`
}

type UInt8 struct {
	Val uint8 `json:"val"`
}

type UInt8Ptr struct {
	Val *uint8 `json:"val"`
}

type UInt16 struct {
	Val uint16 `json:"val"`
}

type UInt16Ptr struct {
	Val *uint16 `json:"val"`
}

type UInt32 struct {
	Val uint32 `json:"val"`
}

type UInt32Ptr struct {
	Val *uint32 `json:"val"`
}

type UInt64 struct {
	Val uint64 `json:"val"`
}

type UInt64Ptr struct {
	Val *uint64 `json:"val"`
}

type Float32 struct {
	Val float32 `json:"val"`
}

type Float32Ptr struct {
	Val *float32 `json:"val"`
}

type Float64 struct {
	Val float64 `json:"val"`
}

type Float64Ptr struct {
	Val *float64 `json:"val"`
}
