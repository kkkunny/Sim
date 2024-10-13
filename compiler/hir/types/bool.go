package types

var Bool BoolType = new(_BoolType_)

// BoolType 布尔类型
type BoolType interface {
	Type
	Bool()
}

type _BoolType_ struct{}

func (self *_BoolType_) String() string {
	return "bool"
}

func (self *_BoolType_) Equal(dst Type) bool {
	_, ok := dst.(BoolType)
	return ok
}

func (self *_BoolType_) Bool() {}
