package hir

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/slices"
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
	AssignableTo(dst Type) bool
	HasDefault()bool
}

func GetInnerType(t Type)Type{
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
	if self == dst{
		return true
	}
	_, ok := dst.(*EmptyType)
	return ok
}

func (self *EmptyType) AssignableTo(dst Type) bool {
	return false
}

func (*EmptyType) HasDefault()bool{
	return false
}

func IsNumberType(t Type)bool{
	return IsIntType(t) || IsFloatType(t)
}

func IsIntType(t Type)bool{
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

func (self *SintType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *SintType) HasSign() bool {
	return true
}

func (self *SintType) GetBits() uint {
	return self.Bits
}

func (*SintType) HasDefault()bool{
	return true
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

func (self *UintType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *UintType) HasSign() bool {
	return false
}

func (self *UintType) GetBits() uint {
	return self.Bits
}

func (*UintType) HasDefault()bool{
	return true
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

func (self *FloatType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *FloatType) GetBits() uint {
	return self.Bits
}

func (*FloatType) HasDefault()bool{
	return true
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

func (self *FuncType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (*FuncType) HasDefault()bool{
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
	if self == dst{
		return true
	}
	_, ok := dst.(*BoolType)
	return ok
}

func (self *BoolType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (*BoolType) HasDefault()bool{
	return true
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

func (self *ArrayType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *ArrayType) HasDefault()bool{
	return self.Elem.HasDefault()
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

func (self *TupleType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *TupleType) HasDefault()bool{
	for _, e := range self.Elems{
		if !e.HasDefault(){
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

func (self *StructType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *StructType) HasDefault()bool{
	// TODO: 结构体默认值
	for iter := self.Fields.Values().Iterator(); iter.Next(); {
		if !iter.Value().Second.HasDefault(){
			return false
		}
	}
	return true
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
	if self == dst{
		return true
	}
	_, ok := dst.(*StringType)
	return ok
}

func (self *StringType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *StringType) HasDefault()bool{
	return true
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

func (self *UnionType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
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

func (self *UnionType) HasDefault()bool{
	for _, e := range self.Elems {
		if !e.HasDefault(){
			return false
		}
	}
	return true
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

func (self *PtrType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *PtrType) HasDefault()bool{
	return true
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

func (self *RefType) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if self.ToPtrType().Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *RefType) ToPtrType() *PtrType {
	return &PtrType{Elem: self.Elem}
}

func (self *RefType) HasDefault()bool{
	return true
}

func IsGenericParam(t Type)bool{
	return stlbasic.Is[*GenericParam](GetInnerType(t))
}

func AsGenericParam(t Type)*GenericParam{
	return GetInnerType(t).(*GenericParam)
}

// GenericParam 泛型参数
type GenericParam struct {
	Name string
}

// ReplaceGenericParam 替换类型中包含的泛型参数
func ReplaceGenericParam(t Type, table hashmap.HashMap[*GenericParam, Type])Type{
	switch tt := t.(type) {
	case *EmptyType, NumberType, *BoolType, *StringType, *StructType:
		return tt
	case *FuncType:
		return &FuncType{
			Ret: ReplaceGenericParam(tt.Ret, table),
			Params: lo.Map(tt.Params, func(item Type, _ int) Type {
				return ReplaceGenericParam(item, table)
			}),
		}
	case *ArrayType:
		return &ArrayType{
			Size: tt.Size,
			Elem: ReplaceGenericParam(tt.Elem, table),
		}
	case *TupleType:
		return &TupleType{Elems: lo.Map(tt.Elems, func(item Type, _ int) Type {
			return ReplaceGenericParam(item, table)
		})}
	case *PtrType:
		return &PtrType{Elem: ReplaceGenericParam(tt.Elem, table)}
	case *RefType:
		return &RefType{Elem: ReplaceGenericParam(tt.Elem, table)}
	case *UnionType:
		return &UnionType{Elems: lo.Map(tt.Elems, func(item Type, _ int) Type {
			return ReplaceGenericParam(item, table)
		})}
	case *GenericParam:
		to := table.Get(tt)
		if to == nil{
			panic("unreachable")
		}
		return to
	default:
		panic("unreachable")
	}
}

func (self *GenericParam) String() string {
	return self.Name
}

func (self *GenericParam) Equal(dst Type) bool {
	return self == dst
}

func (self *GenericParam) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *GenericParam) HasDefault()bool{
	return false
}

// GenericStructInstance 泛型结构体实例
type GenericStructInstance struct {
	Define *GenericStructDef
	Params []Type
}

func (self *GenericStructInstance) String() string {
	params := stlslices.Map(self.Params, func(_ int, e Type) string {
		return e.String()
	})
	return fmt.Sprintf("%s::%s<%s>", self.Define.Pkg, self.Define.Name, strings.Join(params, ", "))
}

func (self *GenericStructInstance) Equal(dst Type) bool {
	if self == dst{
		return true
	}
	t, ok := dst.(*GenericStructInstance)
	if !ok || self.Define != t.Define || len(self.Params) != len(t.Params) {
		return false
	}
	for i, p := range self.Params{
		if !p.Equal(t.Params[i]){
			return false
		}
	}
	return true
}

func (self *GenericStructInstance) AssignableTo(dst Type) bool {
	if self.Equal(dst) {
		return true
	}
	if ut, ok := dst.(*UnionType); ok {
		if ut.Contain(self) {
			return true
		}
	}
	return false
}

func (self *GenericStructInstance) HasDefault()bool{
	// TODO: 泛型结构体实例默认值
	return true
}
