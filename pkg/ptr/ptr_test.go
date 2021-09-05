package ptr_test

import (
	"testing"
	"time"

	"github.com/0xor1/tlbx/pkg/core"
	"github.com/0xor1/tlbx/pkg/ptr"
	"github.com/stretchr/testify/assert"
)

func TestID(t *testing.T) {
	a := assert.New(t)
	idGen := core.NewIDGen()
	vA := idGen.MustNew()
	vB := idGen.MustNew()
	vPtr := ptr.ID(vA)
	a.Equal(vA, *vPtr)
	*vPtr = ptr.IDOr(vPtr, vB)
	a.Equal(vA, *vPtr)
	vPtr = nil
	vLast := ptr.IDOr(vPtr, vB)
	a.Equal(vB, vLast)
}

func TestBool(t *testing.T) {
	a := assert.New(t)
	vPtr := ptr.Bool(true)
	a.True(*vPtr)
	*vPtr = ptr.BoolOr(vPtr, false)
	a.True(*vPtr)
	vPtr = nil
	vLast := ptr.BoolOr(vPtr, false)
	a.False(vLast)
}

func TestString(t *testing.T) {
	a := assert.New(t)
	vA := "a"
	vB := "b"
	vPtr := ptr.String(vA)
	a.Equal(vA, *vPtr)
	*vPtr = ptr.StringOr(vPtr, vB)
	a.Equal(vA, *vPtr)
	vPtr = nil
	vLast := ptr.StringOr(vPtr, vB)
	a.Equal(vB, vLast)
}

func TestTime(t *testing.T) {
	a := assert.New(t)
	vA := time.Now()
	vB := vA.Add(10)
	vPtr := ptr.Time(vA)
	a.Equal(vA, *vPtr)
	*vPtr = ptr.TimeOr(vPtr, vB)
	a.Equal(vA, *vPtr)
	vPtr = nil
	vLast := ptr.TimeOr(vPtr, vB)
	a.Equal(vB, vLast)
}

func TestInt(t *testing.T) {
	a := assert.New(t)
	vA := 1
	vB := 2
	vPtr := ptr.Int(vA)
	a.Equal(vA, *vPtr)
	*vPtr = ptr.IntOr(vPtr, vB)
	a.Equal(vA, *vPtr)
	vPtr = nil
	vLast := ptr.IntOr(vPtr, vB)
	a.Equal(vB, vLast)
}

func TestInt64(t *testing.T) {
	a := assert.New(t)
	vA := int64(1)
	vB := int64(2)
	vPtr := ptr.Int64(vA)
	a.Equal(vA, *vPtr)
	*vPtr = ptr.Int64Or(vPtr, vB)
	a.Equal(vA, *vPtr)
	vPtr = nil
	vLast := ptr.Int64Or(vPtr, vB)
	a.Equal(vB, vLast)
}

func TestUint(t *testing.T) {
	a := assert.New(t)
	vA := uint(1)
	vB := uint(2)
	vPtr := ptr.Uint(vA)
	a.Equal(vA, *vPtr)
	*vPtr = ptr.UintOr(vPtr, vB)
	a.Equal(vA, *vPtr)
	vPtr = nil
	vLast := ptr.UintOr(vPtr, vB)
	a.Equal(vB, vLast)
}

func TestUint8(t *testing.T) {
	a := assert.New(t)
	vA := uint8(1)
	vB := uint8(2)
	vPtr := ptr.Uint8(vA)
	a.Equal(vA, *vPtr)
	*vPtr = ptr.Uint8Or(vPtr, vB)
	a.Equal(vA, *vPtr)
	vPtr = nil
	vLast := ptr.Uint8Or(vPtr, vB)
	a.Equal(vB, vLast)
}

func TestUint64(t *testing.T) {
	a := assert.New(t)
	vA := uint64(1)
	vB := uint64(2)
	vPtr := ptr.Uint64(vA)
	a.Equal(vA, *vPtr)
	*vPtr = ptr.Uint64Or(vPtr, vB)
	a.Equal(vA, *vPtr)
	vPtr = nil
	vLast := ptr.Uint64Or(vPtr, vB)
	a.Equal(vB, vLast)
}

func TestFloat64(t *testing.T) {
	a := assert.New(t)
	vA := float64(1)
	vB := float64(2)
	vPtr := ptr.Float64(vA)
	a.Equal(vA, *vPtr)
	*vPtr = ptr.Float64Or(vPtr, vB)
	a.Equal(vA, *vPtr)
	vPtr = nil
	vLast := ptr.Float64Or(vPtr, vB)
	a.Equal(vB, vLast)
}
