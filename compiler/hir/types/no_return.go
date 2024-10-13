package types

// NoReturnType 无返回类型
type NoReturnType interface {
	Type
	_NoReturnType_()
}

type _NoReturnType_ struct{}

var NoReturn NoReturnType = new(_NoReturnType_)

func (self *_NoReturnType_) String() string {
	return "X"
}

func (self *_NoReturnType_) Equal(dst Type) bool {
	_, ok := dst.(NoReturnType)
	return ok
}

func (self *_NoReturnType_) _NoReturnType_() {}
