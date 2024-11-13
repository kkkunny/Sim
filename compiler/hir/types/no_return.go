package types

// NoReturnType 无返回类型
type NoReturnType interface {
	BuildInType
	_NoReturnType_()
}

type _NoReturnType_ struct{}

var NoReturn NoReturnType = new(_NoReturnType_)

func (self *_NoReturnType_) String() string {
	return "X"
}

func (self *_NoReturnType_) Equal(dst Type, _ ...Type) bool {
	return Is[NoReturnType](dst, true)
}

func (self *_NoReturnType_) _NoReturnType_() {}

func (self *_NoReturnType_) BuildIn() {}
