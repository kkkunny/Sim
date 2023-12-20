package hir

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/pair"
	"github.com/kkkunny/stl/container/stack"
	"github.com/samber/lo"
)

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

	Str = &StringType{}
)

// Type 类型
type Type interface {
	fmt.Stringer
	Equal(dst Type) bool
}

var typeReplaceStack = stack.NewStack[pair.Pair[Type, Type]]()

func GetInnerType(t Type)Type{
	switch tt := t.(type) {
	case *SelfType:
		for iter:=typeReplaceStack.Iterator(); iter.Next(); {
			if iter.Value().First == tt{
				return iter.Value().Second
			}
		}
	}
	return t
}

func IsEmptyType(t Type)bool{
	return stlbasic.Is[*EmptyType](GetInnerType(t))
}

func AsEmptyType(t Type)*EmptyType{
	return GetInnerType(t).(*EmptyType)
}

// EmptyType 空类型
type EmptyType struct{}

func (*EmptyType) String() string {
	return "void"
}

func (self *EmptyType) Equal(dst Type) bool {
	return IsEmptyType(dst)
}

func IsNumberType(t Type)bool{
	t = GetInnerType(t)
	return IsIntType(t) || IsFloatType(t)
}

func IsIntType(t Type)bool{
	t = GetInnerType(t)
	return IsSintType(t) || IsUintType(t)
}

func IsSintType(t Type)bool{
	return stlbasic.Is[*SintType](GetInnerType(t))
}

func AsSintType(t Type)*SintType{
	return GetInnerType(t).(*SintType)
}

// SintType 有符号整型
type SintType struct {
	Bits uint
}

func (self *SintType) String() string {
	if self.Bits == 0{
		return "isize"
	}
	return fmt.Sprintf("i%d", self.Bits)
}

func (self *SintType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	t, ok := dst.(*SintType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}

func IsUintType(t Type)bool{
	return stlbasic.Is[*UintType](GetInnerType(t))
}

func AsUintType(t Type)*UintType{
	return GetInnerType(t).(*UintType)
}

// UintType 无符号整型
type UintType struct {
	Bits uint
}

func (self *UintType) String() string {
	if self.Bits == 0{
		return "usize"
	}
	return fmt.Sprintf("u%d", self.Bits)
}

func (self *UintType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	t, ok := dst.(*UintType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}

func IsFloatType(t Type)bool{
	return stlbasic.Is[*FloatType](GetInnerType(t))
}

func AsFloatType(t Type)*FloatType{
	return GetInnerType(t).(*FloatType)
}

// FloatType 浮点型
type FloatType struct {
	Bits uint
}

func (self *FloatType) String() string {
	return fmt.Sprintf("f%d", self.Bits)
}

func (self *FloatType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	t, ok := dst.(*FloatType)
	if !ok {
		return false
	}
	return self.Bits == t.Bits
}

func IsFuncType(t Type)bool{
	return stlbasic.Is[*FuncType](GetInnerType(t))
}

func AsFuncType(t Type)*FuncType{
	return GetInnerType(t).(*FuncType)
}

// FuncType 函数类型
type FuncType struct {
	Ret    Type
	Params []Type
}

func (self *FuncType) String() string {
	params := lo.Map(self.Params, func(item Type, _ int) string {
		return item.String()
	})
	ret := stlbasic.Ternary(self.Ret.Equal(Empty), "", self.Ret.String())
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), ret)
}

func (self *FuncType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
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

func IsBoolType(t Type)bool{
	return stlbasic.Is[*BoolType](GetInnerType(t))
}

func AsBoolType(t Type)*BoolType{
	return GetInnerType(t).(*BoolType)
}

// BoolType 布尔型
type BoolType struct{}

func (*BoolType) String() string {
	return "bool"
}

func (self *BoolType) Equal(dst Type) bool {
	return IsBoolType(dst)
}

func IsArrayType(t Type)bool{
	return stlbasic.Is[*ArrayType](GetInnerType(t))
}

func AsArrayType(t Type)*ArrayType{
	return GetInnerType(t).(*ArrayType)
}

// ArrayType 数组型
type ArrayType struct {
	Size uint
	Elem Type
}

func (self *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", self.Size, self.Elem)
}

func (self *ArrayType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	t, ok := dst.(*ArrayType)
	if !ok {
		return false
	}
	return self.Size == t.Size && self.Elem.Equal(t.Elem)
}

func IsTupleType(t Type)bool{
	return stlbasic.Is[*TupleType](GetInnerType(t))
}

func AsTupleType(t Type)*TupleType{
	return GetInnerType(t).(*TupleType)
}

// TupleType 元组型
type TupleType struct {
	Elems []Type
}

func (self *TupleType) String() string {
	elems := lo.Map(self.Elems, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *TupleType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
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

func IsStructType(t Type)bool{
	return stlbasic.Is[*StructType](GetInnerType(t))
}

func AsStructType(t Type)*StructType{
	return GetInnerType(t).(*StructType)
}

type StructType = StructDef

func (self *StructType) String() string {
	return fmt.Sprintf("%s::%s", self.Pkg, self.Name)
}

func (self *StructType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	t, ok := dst.(*StructDef)
	if !ok {
		return false
	}
	return self.Pkg==t.Pkg && self.Name==t.Name
}

func IsStringType(t Type)bool{
	return stlbasic.Is[*StringType](GetInnerType(t))
}

func AsStringType(t Type)*StringType{
	return GetInnerType(t).(*StringType)
}

// StringType 字符串型
type StringType struct{}

func (*StringType) String() string {
	return "str"
}

func (self *StringType) Equal(dst Type) bool {
	return IsStringType(dst)
}

func IsUnionType(t Type)bool{
	return stlbasic.Is[*UnionType](GetInnerType(t))
}

func AsUnionType(t Type)*UnionType{
	return GetInnerType(t).(*UnionType)
}

// UnionType 联合类型
type UnionType struct {
	Elems []Type
}

func (self *UnionType) String() string {
	elemStrs := lo.Map(self.Elems, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("<%s>", strings.Join(elemStrs, ", "))
}

func (self *UnionType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	ut, ok := dst.(*UnionType)
	if !ok {
		return false
	}
	if len(self.Elems) != len(ut.Elems){
		return false
	}
	for i, e := range self.Elems {
		if !e.Equal(ut.Elems[i]) {
			return false
		}
	}
	return true
}

// Contain 是否包含类型
func (self *UnionType) Contain(elem Type) bool {
	return self.GetElemIndex(elem) >= 0
}

// GetElemIndex 获取子类型下标
func (self *UnionType) GetElemIndex(elem Type) int {
	for i, e := range self.Elems {
		if e.Equal(elem){
			return i
		}
	}
	return -1
}

func IsPtrType(t Type)bool{
	return stlbasic.Is[*PtrType](GetInnerType(t))
}

func AsPtrType(t Type)*PtrType{
	return GetInnerType(t).(*PtrType)
}

// PtrType 指针类型
type PtrType struct {
	Elem Type
}

func (self *PtrType) String() string {
	return "*" + self.Elem.String() + "?"
}

func (self *PtrType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	pt, ok := dst.(*PtrType)
	if !ok {
		return false
	}
	return self.Elem.Equal(pt.Elem)
}

func IsRefType(t Type)bool{
	return stlbasic.Is[*RefType](GetInnerType(t))
}

func AsRefType(t Type)*RefType{
	return GetInnerType(t).(*RefType)
}

// RefType 引用类型
type RefType struct {
	Elem Type
}

func (self *RefType) String() string {
	return "*" + self.Elem.String()
}

func (self *RefType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	pt, ok := dst.(*RefType)
	if !ok {
		return false
	}
	return self.Elem.Equal(pt.Elem)
}

func (self *RefType) ToPtrType() *PtrType {
	return &PtrType{Elem: self.Elem}
}

func IsSelfType(t Type)bool{
	return stlbasic.Is[*SelfType](GetInnerType(t))
}

func AsSelfType(t Type)*SelfType{
	return GetInnerType(t).(*SelfType)
}

// SelfType Self类型
type SelfType struct{
	Self TypeDef
}

func (self *SelfType) String() string {
	return self.Self.String()
}

func (self *SelfType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	_, ok := dst.(*SelfType)
	return ok
}

// AliasType 别名类型
type AliasType = TypeAliasDef

func (self *AliasType) String() string {
	return self.Target.String()
}

func (self *AliasType) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	_, ok := dst.(*AliasType)
	return ok
}
