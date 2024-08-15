package types

// NoThingType 无返回值类型
type NoThingType struct{}

var NoThing = new(NoThingType)

func (self *NoThingType) String() string {
	return ""
}

func (self *NoThingType) Equal(dst Type) bool {
	_, ok := dst.(*NoThingType)
	return ok
}

func (self *NoThingType) nothing() {}
