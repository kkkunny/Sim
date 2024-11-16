package types

import (
	"unsafe"

	"github.com/kkkunny/Sim/compiler/hir"
)

var Str StrType = new(_StrType_)

// StrType 字符串类型
type StrType interface {
	BuildInType
	Str()
}

type _StrType_ struct{}

func (self *_StrType_) String() string {
	return "str"
}

func (self *_StrType_) Equal(dst hir.Type) bool {
	return Is[StrType](dst, true)
}

func (self *_StrType_) EqualWithSelf(dst hir.Type, _ ...hir.Type) bool {
	return Is[StrType](dst, true)
}

func (self *_StrType_) Str() {}

func (self *_StrType_) BuildIn() {}

func (self *_StrType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
