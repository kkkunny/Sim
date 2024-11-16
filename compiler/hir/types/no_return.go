package types

import (
	"unsafe"

	"github.com/kkkunny/Sim/compiler/hir"
)

// NoReturnType 无返回类型
type NoReturnType interface {
	BuildInType
	_NoReturnType_()
}

type _NoReturnType_ struct{}

var NoReturn NoReturnType = new(_NoReturnType_)

func (self *_NoReturnType_) String() string {
	return "X"
}

func (self *_NoReturnType_) Equal(dst hir.Type) bool {
	return Is[NoReturnType](dst, true)
}

func (self *_NoReturnType_) _NoReturnType_() {}

func (self *_NoReturnType_) BuildIn() {}

func (self *_NoReturnType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
