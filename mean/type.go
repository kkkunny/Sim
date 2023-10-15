package mean

var (
	Empty = &EmptyType{}

	Isize = &SintType{Bits: 64}

	F64 = &FloatType{Bits: 64}
)

// Type 类型
type Type interface {
	Equal(dst Type) bool
}

// TypeIs 类型是否是
func TypeIs[T Type](v Type) bool {
	_, ok := v.(T)
	return ok
}

// EmptyType 空类型
type EmptyType struct{}

func (self *EmptyType) Equal(dst Type) bool {
	_, ok := dst.(*EmptyType)
	return ok
}

// NumberType 数字型
type NumberType interface {
	Type
	GetBits() uint
}

// IntType 整型
type IntType interface {
	NumberType
	HasSign() bool
}

// SintType 有符号整型
type SintType struct {
	Bits uint
}

func (self *SintType) Equal(dst Type) bool {
	t, ok := dst.(*SintType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}

func (self *SintType) HasSign() bool {
	return true
}

func (self *SintType) GetBits() uint {
	return self.Bits
}

// FloatType 浮点型
type FloatType struct {
	Bits uint
}

func (self *FloatType) Equal(dst Type) bool {
	t, ok := dst.(*FloatType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}

func (self *FloatType) GetBits() uint {
	return self.Bits
}

// FuncType 函数类型
type FuncType struct {
	Ret Type
}

func (self *FuncType) Equal(dst Type) bool {
	t, ok := dst.(*FuncType)
	if !ok {
		return false
	}
	return self.Ret.Equal(t.Ret)
}
