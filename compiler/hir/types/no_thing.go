package types

import "unsafe"

// NoThingType 无返回值类型
type NoThingType interface {
	BuildInType
	_NoThingType_()
}

type _NoThingType_ struct{}

var NoThing = new(_NoThingType_)

func (self *_NoThingType_) String() string {
	return ""
}

func (self *_NoThingType_) Equal(dst Type) bool {
	return Is[NoThingType](dst, true)
}

func (self *_NoThingType_) EqualWithSelf(dst Type, _ ...Type) bool {
	return Is[NoThingType](dst, true)
}

func (self *_NoThingType_) _NoThingType_() {}

func (self *_NoThingType_) BuildIn() {}

func (self *_NoThingType_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}
