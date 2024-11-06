package types

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

func (self *_NoThingType_) Equal(dst Type, _ ...Type) bool {
	_, ok := dst.(NoThingType)
	return ok
}

func (self *_NoThingType_) _NoThingType_() {}

func (self *_NoThingType_) BuildIn() {}
