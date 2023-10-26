package mean

var (
	Empty = &EmptyType{}

	Isize = &SintType{Bits: 0}
	I8    = &SintType{Bits: 8}
	I16   = &SintType{Bits: 16}
	I32   = &SintType{Bits: 32}
	I64   = &SintType{Bits: 64}
	I128  = &SintType{Bits: 128}

	Usize = &UintType{Bits: 0}
	U8    = &UintType{Bits: 8}
	U16   = &UintType{Bits: 16}
	U32   = &UintType{Bits: 32}
	U64   = &UintType{Bits: 64}
	U128  = &UintType{Bits: 128}

	F32 = &FloatType{Bits: 32}
	F64 = &FloatType{Bits: 64}

	Bool = &BoolType{}
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

// UintType 无符号整型
type UintType struct {
	Bits uint
}

func (self *UintType) Equal(dst Type) bool {
	t, ok := dst.(*UintType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}

func (self *UintType) HasSign() bool {
	return false
}

func (self *UintType) GetBits() uint {
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
	Ret    Type
	Params []Type
}

func (self *FuncType) Equal(dst Type) bool {
	t, ok := dst.(*FuncType)
	if !ok || len(t.Params) != len(self.Params) {
		return false
	}
	for i, p := range self.Params {
		if !p.Equal(t.Params[i]) {
			return false
		}
	}
	return true
}

// BoolType 布尔型
type BoolType struct{}

func (self *BoolType) Equal(dst Type) bool {
	_, ok := dst.(*BoolType)
	return ok
}

// ArrayType 数组型
type ArrayType struct {
	Size uint
	Elem Type
}

func (self *ArrayType) Equal(dst Type) bool {
	t, ok := dst.(*ArrayType)
	if !ok {
		return false
	}
	return self.Size == t.Size && self.Elem.Equal(t.Elem)
}

// TupleType 元组型
type TupleType struct {
	Elems []Type
}

func (self *TupleType) Equal(dst Type) bool {
	t, ok := dst.(*TupleType)
	if !ok || len(self.Elems) != len(t.Elems) {
		return false
	}
	for i, e := range self.Elems {
		if !e.Equal(t.Elems[i]) {
			return false
		}
	}
	return true
}

type StructType = StructDef
