package types

import (
	"fmt"
	"unsafe"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

// ArrayType 数组类型
type ArrayType interface {
	BuildInType
	Elem() hir.Type
	Size() uint
}

type _ArrayType_ struct {
	elem hir.Type
	size uint
}

func NewArrayType(e hir.Type, s uint) ArrayType {
	return &_ArrayType_{
		elem: e,
		size: s,
	}
}

func (self *_ArrayType_) String() string {
	return fmt.Sprintf("[%d]%s", self.size, self.elem.String())
}

func (self *_ArrayType_) Equal(dst hir.Type) bool {
	t, ok := As[ArrayType](dst, true)
	return ok && self.size == t.Size() && self.elem.Equal(t.Elem())
}

func (self *_ArrayType_) EqualWithSelf(dst hir.Type, selfs ...hir.Type) bool {
	if dst.Equal(Self) && len(selfs) > 0 {
		dst = stlslices.Last(selfs)
	}

	t, ok := As[ArrayType](dst, true)
	return ok && self.size == t.Size() && self.elem.EqualWithSelf(t.Elem(), selfs...)
}

func (self *_ArrayType_) Elem() hir.Type {
	return self.elem
}

func (self *_ArrayType_) Size() uint {
	return self.size
}

func (self *_ArrayType_) BuildIn() {}
func (self *_ArrayType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
