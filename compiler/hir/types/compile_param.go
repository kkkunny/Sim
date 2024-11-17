package types

import (
	"unsafe"

	"github.com/kkkunny/Sim/compiler/hir"
)

// CompileParamType 编译参数
type CompileParamType interface {
	hir.Type
	_GenericParam_()
}

type _CompileParamType_ struct {
	name string
}

func NewGenericParam(name string) CompileParamType {
	return &_CompileParamType_{
		name: name,
	}
}

func (self *_CompileParamType_) String() string {
	return self.name
}

func (self *_CompileParamType_) Equal(dst hir.Type) bool {
	return self == dst
}

func (self *_CompileParamType_) _GenericParam_() {}

func (self *_CompileParamType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
