package types

import (
	"unsafe"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

// GenericParamType 泛型参数类型
type GenericParamType interface {
	VirtualType
	Restraint() (hir.GenericRestraint, bool)
}

type _GenericParamType_ struct {
	name      string
	restraint hir.GenericRestraint
}

func NewGenericParam(name string, restraint ...hir.GenericRestraint) GenericParamType {
	return &_GenericParamType_{
		name:      name,
		restraint: stlslices.Last(restraint),
	}
}

func (self *_GenericParamType_) String() string {
	return self.name
}

func (self *_GenericParamType_) Equal(dst hir.Type) bool {
	return self == dst
}

func (self *_GenericParamType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}

func (self *_GenericParamType_) virtual() {}

func (self *_GenericParamType_) Restraint() (hir.GenericRestraint, bool) {
	return self.restraint, self.restraint != nil
}
