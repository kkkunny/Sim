package hir

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/util"
)

var (
	Empty = &EmptyType{}

	Isize = &SintType{Bits: 0}
	I8    = &SintType{Bits: 8}
	I16   = &SintType{Bits: 16}
	I32   = &SintType{Bits: 32}
	I64   = &SintType{Bits: 64}

	Usize = &UintType{Bits: 0}
	U8    = &UintType{Bits: 8}
	U16   = &UintType{Bits: 16}
	U32   = &UintType{Bits: 32}
	U64   = &UintType{Bits: 64}

	F32 = &FloatType{Bits: 32}
	F64 = &FloatType{Bits: 64}

	Bool = &BoolType{}

	Str = &StringType{}
)

// Type 类型
type Type interface {
	fmt.Stringer
	ID() string
	stlbasic.Hashable
	stlbasic.Comparable[Type]
}

func FlattenType(t Type)Type{
	switch tt := t.(type) {
	case *SelfType:
		return FlattenType(tt.Self.MustValue())
	case *AliasType:
		return FlattenType(tt.Target)
	default:
		return tt
	}
}

func IsEmptyType(t Type)bool{
	return stlbasic.Is[*EmptyType](FlattenType(t))
}

// EmptyType 空类型
type EmptyType struct{}

func (*EmptyType) String() string {
	return "empty"
}

func (self *EmptyType) ID() string          { return self.String() }
func (self *EmptyType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *EmptyType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsNumberType(t Type)bool{
	t = FlattenType(t)
	return IsIntType(t) || IsFloatType(t)
}

func IsIntType(t Type)bool{
	t = FlattenType(t)
	return IsSintType(t) || IsUintType(t)
}

func IsSintType(t Type)bool{
	return stlbasic.Is[*SintType](FlattenType(t))
}

func AsSintType(t Type)*SintType{
	return FlattenType(t).(*SintType)
}

// SintType 有符号整型
type SintType struct {
	Bits uint
}

func (self *SintType) String() string {
	return stlbasic.Ternary(self.Bits==0, "isize", fmt.Sprintf("i%d", self.Bits))
}

func (self *SintType) ID() string          { return self.String() }
func (self *SintType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *SintType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsUintType(t Type)bool{
	return stlbasic.Is[*UintType](FlattenType(t))
}

func AsUintType(t Type)*UintType{
	return FlattenType(t).(*UintType)
}

// UintType 无符号整型
type UintType struct {
	Bits uint
}

func (self *UintType) String() string {
	return stlbasic.Ternary(self.Bits==0, "usize", fmt.Sprintf("u%d", self.Bits))
}

func (self *UintType) ID() string          { return self.String() }
func (self *UintType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *UintType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsFloatType(t Type)bool{
	return stlbasic.Is[*FloatType](FlattenType(t))
}

func AsFloatType(t Type)*FloatType{
	return FlattenType(t).(*FloatType)
}

// FloatType 浮点型
type FloatType struct {
	Bits uint
}

func (self *FloatType) String() string {
	return fmt.Sprintf("f%d", self.Bits)
}

func (self *FloatType) ID() string          { return self.String() }
func (self *FloatType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *FloatType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsFuncType(t Type)bool{
	return stlbasic.Is[*FuncType](FlattenType(t))
}

func AsFuncType(t Type)*FuncType{
	return FlattenType(t).(*FuncType)
}

// FuncType 函数类型
type FuncType struct {
	Ret    Type
	Params []Type
}

func (self *FuncType) String() string {
	params := stlslices.Map(self.Params, func(_ int, e Type) string { return e.String() })
	ret := stlbasic.Ternary(self.Ret.Equal(Empty), "", self.Ret.String())
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), ret)
}

func (self *FuncType) ID() string          { return self.String() }
func (self *FuncType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *FuncType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsBoolType(t Type)bool{
	return stlbasic.Is[*BoolType](FlattenType(t))
}

// BoolType 布尔型
type BoolType struct{}

func (*BoolType) String() string {
	return "bool"
}

func (self *BoolType) ID() string          { return self.String() }
func (self *BoolType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *BoolType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsArrayType(t Type)bool{
	return stlbasic.Is[*ArrayType](FlattenType(t))
}

func AsArrayType(t Type)*ArrayType{
	return FlattenType(t).(*ArrayType)
}

// ArrayType 数组型
type ArrayType struct {
	Size uint
	Elem Type
}

func (self *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", self.Size, self.Elem)
}

func (self *ArrayType) ID() string          { return self.String() }
func (self *ArrayType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *ArrayType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsTupleType(t Type)bool{
	return stlbasic.Is[*TupleType](FlattenType(t))
}

func AsTupleType(t Type)*TupleType{
	return FlattenType(t).(*TupleType)
}

// TupleType 元组型
type TupleType struct {
	Elems []Type
}

func (self *TupleType) String() string {
	elems := stlslices.Map(self.Elems, func(_ int, e Type) string {return e.String()})
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *TupleType) ID() string          { return self.String() }
func (self *TupleType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *TupleType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsStructType(t Type)bool{
	return stlbasic.Is[*StructType](FlattenType(t))
}

func AsStructType(t Type)*StructType{
	return FlattenType(t).(*StructType)
}

type StructType = StructDef

func (self *StructType) String() string {
	return stlbasic.Ternary(self.Pkg.Equal(BuildInPackage), self.Name, fmt.Sprintf("%s::%s", self.Pkg, self.Name))
}

func (self *StructType) ID() string          { return self.String() }
func (self *StructType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *StructType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsStringType(t Type)bool{
	return stlbasic.Is[*StringType](FlattenType(t))
}

func AsStringType(t Type)*StringType{
	return FlattenType(t).(*StringType)
}

// StringType 字符串型
type StringType struct{}

func (*StringType) String() string {
	return "str"
}

func (self *StringType) ID() string          { return self.String() }
func (self *StringType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *StringType) Equal(dst Type) bool { return self.ID() == dst.ID() }

func IsUnionType(t Type)bool{
	return stlbasic.Is[*UnionType](FlattenType(t))
}

func AsUnionType(t Type)*UnionType{
	return FlattenType(t).(*UnionType)
}

// UnionType 联合类型
type UnionType struct {
	Elems []Type
}

func (self *UnionType) String() string {
	elems := stlslices.Map(self.Elems, func(_ int, e Type) string {return e.String()})
	return fmt.Sprintf("<%s>", strings.Join(elems, ", "))
}

func (self *UnionType) ID() string          { return self.String() }
func (self *UnionType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *UnionType) Equal(dst Type) bool { return self.ID() == dst.ID() }

// Contain 是否包含类型
func (self *UnionType) Contain(dst Type) bool {
	if IsUnionType(dst){
		return stlslices.All(AsUnionType(dst).Elems, func(_ int, de Type) bool {
			return self.Contain(de)
		})
	}else{
		for _, e := range self.Elems {
			if IsUnionType(e){
				if AsUnionType(e).Contain(dst){
					return true
				}
			}else{
				if e.Equal(dst){
					return true
				}
			}
		}
		return false
	}
}

func IsPointer(t Type)bool{
	t = FlattenType(t)
	return IsRefType(t) || IsFuncType(t)
}

func IsRefType(t Type)bool{
	return stlbasic.Is[*RefType](FlattenType(t))
}

func AsRefType(t Type)*RefType{
	return FlattenType(t).(*RefType)
}

// RefType 引用类型
type RefType struct {
	Mut bool
	Elem Type
}

func (self *RefType) String() string {
	return stlbasic.Ternary(self.Mut, "&"+self.Elem.String(), "&mut "+self.Elem.String())
}

func (self *RefType) ID() string          { return self.String() }
func (self *RefType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *RefType) Equal(dst Type) bool { return self.ID() == dst.ID() }

// SelfType Self类型
type SelfType struct{
	Self util.Option[TypeDef]
}

func (self *SelfType) String() string {
	return self.Self.MustValue().String()
}

func (self *SelfType) ID() string          { return self.String() }
func (self *SelfType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *SelfType) Equal(dst Type) bool { return self.ID() == dst.ID() }

// AliasType 别名类型
type AliasType = TypeAliasDef

func (self *AliasType) String() string {
	return stlbasic.Ternary(self.Pkg.Equal(BuildInPackage), self.Name, fmt.Sprintf("%s::%s", self.Pkg, self.Name))
}

func (self *AliasType) ID() string          { return self.String() }
func (self *AliasType) Hash() uint64        { return stlbasic.Hash(self.ID()) }
func (self *AliasType) Equal(dst Type) bool { return self.ID() == dst.ID() }
