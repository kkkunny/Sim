package types

import (
	"fmt"
	"strings"
	"unsafe"

	stlslices "github.com/kkkunny/stl/container/slices"
)

// TupleType 元组类型
type TupleType interface {
	BuildInType
	Elems() []Type
}

func NewTupleType(es ...Type) TupleType {
	return &_TupleType_{
		elems: es,
	}
}

type _TupleType_ struct {
	elems []Type
}

func (self *_TupleType_) String() string {
	elems := stlslices.Map(self.elems, func(_ int, e Type) string { return e.String() })
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *_TupleType_) Equal(dst Type) bool {
	t, ok := As[TupleType](dst, true)
	if !ok || len(self.elems) != len(t.Elems()) {
		return false
	}
	return stlslices.All(self.elems, func(i int, e Type) bool {
		return e.Equal(t.Elems()[i])
	})
}

func (self *_TupleType_) EqualWithSelf(dst Type, selfs ...Type) bool {
	if dst.Equal(Self) && len(selfs) > 0 {
		dst = stlslices.Last(selfs)
	}

	t, ok := As[TupleType](dst, true)
	if !ok || len(self.elems) != len(t.Elems()) {
		return false
	}
	return stlslices.All(self.elems, func(i int, e Type) bool {
		return e.EqualWithSelf(t.Elems()[i], selfs...)
	})
}

func (self *_TupleType_) Elems() []Type {
	return self.elems
}

func (self *_TupleType_) BuildIn() {}

func (self *_TupleType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
