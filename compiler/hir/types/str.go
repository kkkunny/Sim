package types

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

func (self *_StrType_) Equal(dst Type, _ ...Type) bool {
	_, ok := dst.(StrType)
	return ok
}

func (self *_StrType_) Str() {}

func (self *_StrType_) BuildIn() {}
