package types

import "unsafe"

var Self SelfType = new(_SelfType_)

// SelfType Self类型
type SelfType interface {
	Type
	Self()
}

type _SelfType_ struct{}

func (self *_SelfType_) String() string {
	return "Self"
}

func (self *_SelfType_) Equal(dst Type) bool {
	return Is[SelfType](dst)
}

func (self *_SelfType_) EqualWithSelf(dst Type, _ ...Type) bool {
	return Is[SelfType](dst)
}

func (self *_SelfType_) Self() {}

func (self *_SelfType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
