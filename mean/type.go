package mean

var (
	Empty = &EmptyType{}

	Isize = &SintType{Bits: 64}
)

// Type 类型
type Type interface {
	Equal(dst Type) bool
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
