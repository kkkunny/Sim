package hir

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/pair"
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
	EqualTo(dst Type) bool
}

func FlattenType(t Type)Type{
	switch tt := t.(type) {
	case *SelfType:
		return FlattenType(tt.Self)
	case *AliasType:
		return FlattenType(tt.Target)
	case *GenericStructInst:
		return tt.StructType()
	default:
		return tt
	}
}

func IsEmptyType(t Type)bool{
	return stlbasic.Is[*EmptyType](FlattenType(t))
}

func AsEmptyType(t Type)*EmptyType{
	return FlattenType(t).(*EmptyType)
}

// EmptyType 空类型
type EmptyType struct{}

func (*EmptyType) String() string {
	return "empty"
}

func (self *EmptyType) EqualTo(dst Type) bool {
	return IsEmptyType(dst)
}

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
	if self.Bits == 0{
		return "isize"
	}
	return fmt.Sprintf("i%d", self.Bits)
}

func (self *SintType) EqualTo(dst Type) bool {
	if !IsSintType(dst){
		return false
	}
	dstType := AsSintType(dst)
	return self.Bits == dstType.Bits
}

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
	if self.Bits == 0{
		return "usize"
	}
	return fmt.Sprintf("u%d", self.Bits)
}

func (self *UintType) EqualTo(dst Type) bool {
	if !IsUintType(dst){
		return false
	}
	dstType := AsUintType(dst)
	return self.Bits == dstType.Bits
}

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

func (self *FloatType) EqualTo(dst Type) bool {
	if !IsFloatType(dst){
		return false
	}
	dstType := AsFloatType(dst)
	return self.Bits == dstType.Bits
}

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
	params := lo.Map(self.Params, func(item Type, _ int) string {
		return item.String()
	})
	ret := stlbasic.Ternary(self.Ret.EqualTo(Empty), "", self.Ret.String())
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), ret)
}

func (self *FuncType) EqualTo(dst Type) bool {
	if !IsFuncType(dst){
		return false
	}
	dstType := AsFuncType(dst)
	if !self.Ret.EqualTo(dstType.Ret) || len(self.Params) != len(dstType.Params){
		return false
	}
	for i, p := range self.Params{
		if !p.EqualTo(dstType.Params[i]){
			return false
		}
	}
	return true
}

func IsBoolType(t Type)bool{
	return stlbasic.Is[*BoolType](FlattenType(t))
}

func AsBoolType(t Type)*BoolType{
	return FlattenType(t).(*BoolType)
}

// BoolType 布尔型
type BoolType struct{}

func (*BoolType) String() string {
	return "bool"
}

func (self *BoolType) EqualTo(dst Type) bool {
	return IsBoolType(dst)
}

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

func (self *ArrayType) EqualTo(dst Type) bool {
	if !IsArrayType(dst){
		return false
	}
	dstType := AsArrayType(dst)
	return self.Elem.EqualTo(dstType.Elem)
}

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
	elems := lo.Map(self.Elems, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *TupleType) EqualTo(dst Type) bool {
	if !IsTupleType(dst){
		return false
	}
	dstType := AsTupleType(dst)
	if len(self.Elems) != len(dstType.Elems){
		return false
	}
	for i, e := range self.Elems{
		if !e.EqualTo(dstType.Elems[i]){
			return false
		}
	}
	return true
}

func IsStructType(t Type)bool{
	return stlbasic.Is[*StructType](FlattenType(t))
}

func AsStructType(t Type)*StructType{
	return FlattenType(t).(*StructType)
}

type StructType = StructDef

func (self *StructType) String() string {
	return fmt.Sprintf("%s::%s", self.Pkg, self.Name)
}

func (self *StructType) EqualTo(dst Type) bool {
	if !IsStructType(dst){
		return false
	}
	dstType := AsStructType(dst)
	return self.Pkg == dstType.Pkg && self.Name == dstType.Name
}

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

func (self *StringType) EqualTo(dst Type) bool {
	return IsStringType(dst)
}

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
	elemStrs := lo.Map(self.Elems, func(item Type, _ int) string {
		return item.String()
	})
	return fmt.Sprintf("<%s>", strings.Join(elemStrs, ", "))
}

func (self *UnionType) EqualTo(dst Type) bool {
	if !IsUnionType(dst){
		return false
	}
	dstType := AsUnionType(dst)
	if len(self.Elems) != len(dstType.Elems){
		return false
	}
	for i, e := range dstType.Elems{
		if !e.EqualTo(dstType.Elems[i]){
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
		if e.EqualTo(elem){
			return i
		}
	}
	return -1
}

func IsPtrType(t Type)bool{
	return stlbasic.Is[*PtrType](FlattenType(t))
}

func AsPtrType(t Type)*PtrType{
	return FlattenType(t).(*PtrType)
}

// PtrType 指针类型
type PtrType struct {
	Elem Type
}

func (self *PtrType) String() string {
	return "*?" + self.Elem.String()
}

func (self *PtrType) EqualTo(dst Type) bool {
	if !IsPtrType(dst){
		return false
	}
	return self.Elem.EqualTo(AsPtrType(dst).Elem)
}

func (self *PtrType) ToRefType() *RefType {
	return &RefType{Elem: self.Elem}
}

func IsRefType(t Type)bool{
	return stlbasic.Is[*RefType](FlattenType(t))
}

func AsRefType(t Type)*RefType{
	return FlattenType(t).(*RefType)
}

// RefType 引用类型
type RefType struct {
	Elem Type
}

func (self *RefType) String() string {
	return "*" + self.Elem.String()
}

func (self *RefType) EqualTo(dst Type) bool {
	if !IsRefType(dst){
		return false
	}
	return self.Elem.EqualTo(AsRefType(dst).Elem)
}

func (self *RefType) ToPtrType() *PtrType {
	return &PtrType{Elem: self.Elem}
}

// SelfType Self类型
type SelfType struct{
	Self TypeDef
}

func (self *SelfType) String() string {
	return self.Self.String()
}

func (self *SelfType) EqualTo(dst Type) bool {
	return self.Self.EqualTo(dst)
}

// AliasType 别名类型
type AliasType = TypeAliasDef

func (self *AliasType) String() string {
	return self.Target.String()
}

func (self *AliasType) EqualTo(dst Type) bool {
	return self.Target.EqualTo(dst)
}

// ReplaceAllGenericIdent 替换所有泛型标识符类型
func ReplaceAllGenericIdent(maps hashmap.HashMap[*GenericIdentType, Type], t Type)Type{
	switch tt := t.(type) {
	case *EmptyType, *SintType, *UintType, *FloatType, *StringType, *AliasType, *SelfType, *BoolType:
		return tt
	case *PtrType:
		return &PtrType{Elem: ReplaceAllGenericIdent(maps, tt.Elem)}
	case *RefType:
		return &RefType{Elem: ReplaceAllGenericIdent(maps, tt.Elem)}
	case *FuncType:
		return &FuncType{
			Ret: ReplaceAllGenericIdent(maps, tt.Ret),
			Params: stlslices.Map(tt.Params, func(_ int, e Type) Type {
				return ReplaceAllGenericIdent(maps, e)
			}),
		}
	case *ArrayType:
		return &ArrayType{
			Size: tt.Size,
			Elem: ReplaceAllGenericIdent(maps, tt.Elem),
		}
	case *TupleType:
		return &TupleType{Elems: stlslices.Map(tt.Elems, func(_ int, e Type) Type {
			return ReplaceAllGenericIdent(maps, e)
		})}
	case *StructType:
		fields := stliter.Map[pair.Pair[string, pair.Pair[bool, Type]], pair.Pair[string, pair.Pair[bool, Type]], linkedhashmap.LinkedHashMap[string, pair.Pair[bool, Type]]](tt.Fields, func(e pair.Pair[string, pair.Pair[bool, Type]]) pair.Pair[string, pair.Pair[bool, Type]] {
			return pair.NewPair[string, pair.Pair[bool, Type]](e.First, pair.NewPair[bool, Type](e.Second.First, ReplaceAllGenericIdent(maps, e.Second.Second)))
		})
		return &StructType{
			Pkg: tt.Pkg,
			Public: tt.Public,
			Name: tt.Name,
			Fields: fields,
			Methods: tt.Methods,
		}
	case *UnionType:
		return &UnionType{Elems: stlslices.Map(tt.Elems, func(_ int, e Type) Type {
			return ReplaceAllGenericIdent(maps, e)
		})}
	case *GenericIdentType:
		return maps.Get(tt)
	case *GenericStructInst:
		return &GenericStructInst{
			Define: tt.Define,
			Params: stlslices.Map(tt.Params, func(i int, e Type) Type {
				return ReplaceAllGenericIdent(maps, e)
			}),
		}
	default:
		panic("unreachable")
	}
}

// GenericIdentType 泛型标识符类型
type GenericIdentType struct {
	Belong GlobalGenericDef
	Name string
}

func (self *GenericIdentType) String() string {
	return self.Name
}

func (self *GenericIdentType) EqualTo(dst Type) bool {
	dstType, ok := FlattenType(dst).(*GenericIdentType)
	if !ok{
		return false
	}
	return self.Belong == dstType.Belong && self.Name == dstType.Name
}

// GenericStructInst 泛型结构体实例
type GenericStructInst struct {
	Define *GenericStructDef
	Params []Type
}

func (self *GenericStructInst) GetPackage()Package{
	return self.Define.GetPackage()
}
func (self *GenericStructInst) GetPublic() bool{
	return self.Define.GetPublic()
}

func (self *GenericStructInst) String() string {
	return fmt.Sprintf("%s::%s", self.Define.Pkg, self.GetName())
}

func (self *GenericStructInst) EqualTo(dst Type) bool {
	return self.StructType().EqualTo(dst)
}

func (self *GenericStructInst) GetName()string{
	return fmt.Sprintf("%s::<%s>", self.Define.Name, strings.Join(stlslices.Map(self.Params, func(i int, e Type) string {
		return e.String()
	}), ", "))
}

func (self *GenericStructInst) StructType()*StructType{
	maps := hashmap.NewHashMapWithCapacity[*GenericIdentType, Type](uint(len(self.Params)))
	var i int
	for iter:=self.Define.GenericParams.Iterator(); iter.Next(); {
		maps.Set(iter.Value().Second, self.Params[i])
		i++
	}
	fields := stliter.Map[pair.Pair[string, pair.Pair[bool, Type]], pair.Pair[string, pair.Pair[bool, Type]], linkedhashmap.LinkedHashMap[string, pair.Pair[bool, Type]]](self.Define.Fields, func(e pair.Pair[string, pair.Pair[bool, Type]]) pair.Pair[string, pair.Pair[bool, Type]] {
		return pair.NewPair[string, pair.Pair[bool, Type]](e.First, pair.NewPair[bool, Type](e.Second.First, ReplaceAllGenericIdent(maps, e.Second.Second)))
	})
	methods := stliter.Map[pair.Pair[string, *GenericStructMethodDef], pair.Pair[string, GlobalMethod], hashmap.HashMap[string, GlobalMethod]](self.Define.Methods, func(e pair.Pair[string, *GenericStructMethodDef]) pair.Pair[string, GlobalMethod] {
		return pair.NewPair[string, GlobalMethod](e.First, e.Second)
	})
	st := &StructType{
		Pkg: self.Define.Pkg,
		Public: self.Define.Public,
		Name: self.GetName(),
		Fields: fields,
		Methods: methods,
		genericParams: self.Params,
	}
	return st
}
