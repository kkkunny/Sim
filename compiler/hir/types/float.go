package types

import (
	"fmt"
	"unsafe"

	"github.com/kkkunny/Sim/compiler/hir"
)

// FloatTypeKind 浮点型类型
type FloatTypeKind uint8

const (
	FloatTypeKindHalf FloatTypeKind = 2 << iota
	FloatTypeKindFloat
	FloatTypeKindDouble
	FloatTypeKindFP128
)

var (
	F16  FloatType = &_FloatType_{kind: FloatTypeKindHalf}
	F32  FloatType = &_FloatType_{kind: FloatTypeKindFloat}
	F64  FloatType = &_FloatType_{kind: FloatTypeKindDouble}
	F128 FloatType = &_FloatType_{kind: FloatTypeKindFP128}
)

// FloatType 浮点型
type FloatType interface {
	NumType
	SignedType
	Kind() FloatTypeKind
}

type _FloatType_ struct {
	kind FloatTypeKind
}

func (self *_FloatType_) String() string {
	return fmt.Sprintf("f%d", self.kind*8)
}

func (self *_FloatType_) Equal(dst hir.Type) bool {
	t, ok := As[FloatType](dst, true)
	return ok && self.kind == t.Kind()
}

func (self *_FloatType_) EqualWithSelf(dst hir.Type, _ ...hir.Type) bool {
	t, ok := As[FloatType](dst, true)
	return ok && self.kind == t.Kind()
}

func (self *_FloatType_) Kind() FloatTypeKind {
	return self.kind
}

func (self *_FloatType_) Number() {}

func (self *_FloatType_) Signed() {}

func (self *_FloatType_) BuildIn() {}

func (self *_FloatType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
