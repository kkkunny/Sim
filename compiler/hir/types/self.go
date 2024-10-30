package types

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
	_, ok := dst.(SelfType)
	return ok
}

func (self *_SelfType_) Self() {}
