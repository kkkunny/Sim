package types

import (
	"unsafe"

	"github.com/kkkunny/Sim/compiler/hir"
)

// GenericParamType 泛型参数类型
type GenericParamType interface {
	VirtualType
	_GenericParamType_()
}

type _GenericParamType_ struct {
	name string
}

func NewGenericParam(name string) GenericParamType {
	return &_GenericParamType_{
		name: name,
	}
}

func (self *_GenericParamType_) String() string {
	return self.name
}

func (self *_GenericParamType_) Equal(dst hir.Type) bool {
	return self == dst
}

func (self *_GenericParamType_) _GenericParamType_() {}

func (self *_GenericParamType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}

func (self *_GenericParamType_) virtual() {}
