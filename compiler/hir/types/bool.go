package types

import (
	"unsafe"

	"github.com/kkkunny/Sim/compiler/hir"
)

var Bool BoolType = new(_BoolType_)

// BoolType 布尔类型
type BoolType interface {
	BuildInType
	Bool()
}

type _BoolType_ struct{}

func (self *_BoolType_) String() string {
	return "bool"
}

func (self *_BoolType_) Equal(dst hir.Type) bool {
	return Is[BoolType](dst, true)
}

func (self *_BoolType_) Bool() {}

func (self *_BoolType_) BuildIn() {}

func (self *_BoolType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
