package types

import (
	"unsafe"

	"github.com/kkkunny/Sim/compiler/hir"
)

var Self SelfType = new(_SelfType_)

// SelfType Self类型
type SelfType interface {
	hir.Type
	_Self_()
}

type _SelfType_ struct{}

func (self *_SelfType_) String() string {
	return "Self"
}

func (self *_SelfType_) Equal(dst hir.Type) bool {
	return Is[SelfType](dst)
}

func (self *_SelfType_) _Self_() {}

func (self *_SelfType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
