package types

// NoReturnType 无返回类型
type NoReturnType struct{}

var NoReturn = new(NoReturnType)

func (self *NoReturnType) String() string {
	return "X"
}

func (self *NoReturnType) Equal(dst Type) bool {
	_, ok := dst.(*NoReturnType)
	return ok
}

func (self *NoReturnType) nothing() {}
