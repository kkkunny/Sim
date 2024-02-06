package hir

import (
	"fmt"
	"sort"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/slices"

	runtimeType "github.com/kkkunny/Sim/runtime/types"
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
// 除了运行时类型外，还有一些编译器需要用到的工具类型
type Type interface {
	fmt.Stringer
	EqualTo(dst Type) bool
}

// ToRuntimeType 转换成运行时类型
func ToRuntimeType(t Type)runtimeType.Type{
	return ToActualType(t).ToRuntimeType()
}

// ActualType 真实类型
// 该类型和运行时类型是1对1关系
type ActualType interface {
	Type
	ToRuntimeType()runtimeType.Type
}

// ToActualType 转换成真实类型
func ToActualType(t Type)ActualType{
	switch tt := t.(type) {
	case ActualType:
		return tt
	case *SelfType:
		return ToActualType(tt.Self.MustValue())
	case *AliasType:
		return ToActualType(tt.Target)
	default:
		panic("unreachable")
	}
}

// TryType 尝试断言为指定类型
func TryType[T ActualType](t Type)(T, bool){
	at, ok := ToActualType(t).(T)
	return at, ok
}

// IsType 是否可以断言为指定类型
func IsType[T ActualType](t Type)bool{
	_, ok := TryType[T](t)
	return ok
}

// AsType 断言为指定类型
func AsType[T ActualType](t Type)T{
	at, ok := TryType[T](t)
	if !ok{
		panic("unreachable")
	}
	return at
}

func IsNumberType(t Type)bool{return IsIntType(t) || IsType[*FloatType](t)}
func IsIntType(t Type)bool{return IsType[*SintType](t) || IsType[*UintType](t)}
func IsPointer(t Type)bool{return IsType[*RefType](t) || IsType[*FuncType](t)}

// EmptyType 空类型
type EmptyType struct{}

func (*EmptyType) String() string {
	return "empty"
}

func (self *EmptyType) EqualTo(dst Type) bool {
	return IsType[*EmptyType](dst)
}

func (self *EmptyType) ToRuntimeType()runtimeType.Type{
	return runtimeType.TypeEmpty
}

// SintType 有符号整型
type SintType struct {
	Bits uint
}

func (self *SintType) String() string {
	return stlbasic.Ternary(self.Bits==0, "isize", fmt.Sprintf("i%d", self.Bits))
}

func (self *SintType) EqualTo(dst Type) bool {
	at, ok := TryType[*SintType](dst)
	if !ok{
		return false
	}
	return self.Bits == at.Bits
}

func (self *SintType) ToRuntimeType()runtimeType.Type{
	switch self.Bits {
	case 0:
		return runtimeType.TypeIsize
	case 8:
		return runtimeType.TypeI8
	case 16:
		return runtimeType.TypeI16
	case 32:
		return runtimeType.TypeI32
	case 64:
		return runtimeType.TypeI64
	default:
		panic("unreachable")
	}
}

// UintType 无符号整型
type UintType struct {
	Bits uint
}

func (self *UintType) String() string {
	return stlbasic.Ternary(self.Bits==0, "usize", fmt.Sprintf("u%d", self.Bits))
}

func (self *UintType) EqualTo(dst Type) bool {
	at, ok := TryType[*UintType](dst)
	if !ok{
		return false
	}
	return self.Bits == at.Bits
}

func (self *UintType) ToRuntimeType()runtimeType.Type{
	switch self.Bits {
	case 0:
		return runtimeType.TypeUsize
	case 8:
		return runtimeType.TypeU8
	case 16:
		return runtimeType.TypeU16
	case 32:
		return runtimeType.TypeU32
	case 64:
		return runtimeType.TypeU64
	default:
		panic("unreachable")
	}
}

// FloatType 浮点型
type FloatType struct {
	Bits uint
}

func (self *FloatType) String() string {
	return fmt.Sprintf("f%d", self.Bits)
}

func (self *FloatType) EqualTo(dst Type) bool {
	at, ok := TryType[*FloatType](dst)
	if !ok{
		return false
	}
	return self.Bits == at.Bits
}

func (self *FloatType) ToRuntimeType()runtimeType.Type{
	switch self.Bits {
	case 32:
		return runtimeType.TypeF32
	case 64:
		return runtimeType.TypeF64
	default:
		panic("unreachable")
	}
}

// FuncType 函数类型
type FuncType struct {
	Ret    Type
	Params []Type
}

func NewFuncType(ret Type, params ...Type)*FuncType{
	return &FuncType{
		Ret: ret,
		Params: params,
	}
}

func (self *FuncType) String() string {
	params := stlslices.Map(self.Params, func(_ int, e Type) string { return e.String() })
	ret := stlbasic.Ternary(self.Ret.EqualTo(Empty), "", self.Ret.String())
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), ret)
}

func (self *FuncType) EqualTo(dst Type) bool {
	at, ok := TryType[*FuncType](dst)
	if !ok{
		return false
	}
	if !self.Ret.EqualTo(at.Ret) || len(self.Params) != len(at.Params){
		return false
	}
	for i, p := range self.Params{
		if !p.EqualTo(at.Params[i]){
			return false
		}
	}
	return true
}

func (self *FuncType) ToRuntimeType()runtimeType.Type{
	return runtimeType.NewFuncType(ToRuntimeType(self.Ret), stlslices.Map(self.Params, func(_ int, e Type) runtimeType.Type {
		return ToRuntimeType(e)
	})...)
}

// BoolType 布尔型
type BoolType struct{}

func (*BoolType) String() string {
	return "bool"
}

func (self *BoolType) EqualTo(dst Type) bool {
	return IsType[*BoolType](dst)
}

func (self *BoolType) ToRuntimeType()runtimeType.Type{
	return runtimeType.TypeBool
}

// ArrayType 数组型
type ArrayType struct {
	Size uint64
	Elem Type
}

func NewArrayType(size uint64, elem Type)*ArrayType{
	return &ArrayType{
		Size: size,
		Elem: elem,
	}
}

func (self *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", self.Size, self.Elem)
}

func (self *ArrayType) EqualTo(dst Type) bool {
	at, ok := TryType[*ArrayType](dst)
	if !ok{
		return false
	}
	return self.Size == at.Size && self.Elem.EqualTo(at.Elem)
}

func (self *ArrayType) ToRuntimeType()runtimeType.Type{
	return runtimeType.NewArrayType(self.Size, ToRuntimeType(self.Elem))
}

// TupleType 元组型
type TupleType struct {
	Elems []Type
}

func NewTupleType(elem ...Type)*TupleType{
	return &TupleType{
		Elems: elem,
	}
}

func (self *TupleType) String() string {
	elems := stlslices.Map(self.Elems, func(_ int, e Type) string {return e.String()})
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *TupleType) EqualTo(dst Type) bool {
	at, ok := TryType[*TupleType](dst)
	if !ok{
		return false
	}
	if len(self.Elems) != len(at.Elems){
		return false
	}
	for i, e := range self.Elems{
		if !e.EqualTo(at.Elems[i]){
			return false
		}
	}
	return true
}

func (self *TupleType) ToRuntimeType()runtimeType.Type{
	return runtimeType.NewTupleType(stlslices.Map(self.Elems, func(_ int, e Type) runtimeType.Type {
		return ToRuntimeType(e)
	})...)
}

type StructType = StructDef

func (self *StructType) String() string {
	return stlbasic.Ternary(self.Pkg.Equal(BuildInPackage), self.Name, fmt.Sprintf("%s::%s", self.Pkg, self.Name))
}

func (self *StructType) EqualTo(dst Type) bool {
	at, ok := TryType[*StructType](dst)
	if !ok{
		return false
	}
	return self.Pkg == at.Pkg && self.Name == at.Name
}

func (self *StructType) ToRuntimeType()runtimeType.Type{
	fields := stliter.Map[pair.Pair[string, Field], runtimeType.Field, dynarray.DynArray[runtimeType.Field]](self.Fields, func(e pair.Pair[string, Field]) runtimeType.Field {
		return runtimeType.NewField(ToRuntimeType(e.Second.Type), e.First)
	}).ToSlice()
	methods := stliter.Map[pair.Pair[string, GlobalMethod], runtimeType.Method, dynarray.DynArray[runtimeType.Method]](self.Methods, func(e pair.Pair[string, GlobalMethod]) runtimeType.Method {
		return runtimeType.NewMethod(ToRuntimeType(e.Second.GetFuncType()).(*runtimeType.FuncType), e.First)
	}).ToSlice()
	return runtimeType.NewStructType(self.Pkg.String(), self.Name, fields, methods)
}

// StringType 字符串型
type StringType struct{}

func (*StringType) String() string {
	return "str"
}

func (self *StringType) EqualTo(dst Type) bool {
	return IsType[*StringType](dst)
}

func (self *StringType) ToRuntimeType()runtimeType.Type{
	return runtimeType.TypeStr
}

// UnionType 联合类型
type UnionType struct {
	Elems []Type
}

func NewUnionType(elems ...Type)*UnionType{
	elems = stlslices.FlatMap(elems, func(_ int, e Type) []Type {
		if at, ok := TryType[*UnionType](e); ok{
			return at.Elems
		}else{
			return []Type{e}
		}
	})
	sort.Slice(elems, func(i, j int) bool {
		return elems[i].String() < elems[j].String()
	})

	flatElems := make([]Type, 0, len(elems))
loop:
	for _, e := range elems{
		for _, fe := range flatElems{
			if e.EqualTo(fe){
				continue loop
			}
		}
		flatElems = append(flatElems, e)
	}

	return &UnionType{Elems: flatElems}
}

func (self *UnionType) String() string {
	elems := stlslices.Map(self.Elems, func(_ int, e Type) string {return e.String()})
	return fmt.Sprintf("<%s>", strings.Join(elems, ", "))
}

func (self *UnionType) EqualTo(dst Type) bool {
	at, ok := TryType[*UnionType](dst)
	if !ok{
		return false
	}
	if len(self.Elems) != len(at.Elems){
		return false
	}
	for i, e := range self.Elems{
		if !e.EqualTo(at.Elems[i]){
			return false
		}
	}
	return true
}

func (self *UnionType) ToRuntimeType()runtimeType.Type{
	return runtimeType.NewUnionType(stlslices.Map(self.Elems, func(_ int, e Type) runtimeType.Type {
		return ToRuntimeType(e)
	})...)
}

// Contain 是否包含类型
func (self *UnionType) Contain(dst Type) bool {
	if ut, ok := TryType[*UnionType](dst); ok{
		return stlslices.All(ut.Elems, func(_ int, e Type) bool {return self.Contain(e)})
	}else{
		for _, e := range self.Elems {
			if e.EqualTo(dst){
				return true
			}
		}
		return false
	}
}

// RefType 引用类型
type RefType struct {
	Mut bool
	Elem Type
}

func NewRefType(mut bool, elem Type)*RefType{
	return &RefType{
		Mut: mut,
		Elem: elem,
	}
}

func (self *RefType) String() string {
	return stlbasic.Ternary(self.Mut, "&"+self.Elem.String(), "&mut "+self.Elem.String())
}

func (self *RefType) EqualTo(dst Type) bool {
	at, ok := TryType[*RefType](dst)
	if !ok{
		return false
	}
	return self.Mut == at.Mut && self.Elem.EqualTo(at.Elem)
}

func (self *RefType) ToRuntimeType()runtimeType.Type{
	return runtimeType.NewRefType(self.Mut, ToRuntimeType(self.Elem))
}

// SelfType Self类型
type SelfType struct{
	Self util.Option[TypeDef]
}

func NewSelfType(self TypeDef)*SelfType{
	return &SelfType{
		Self: util.Some(self),
	}
}

func (self *SelfType) String() string {
	return self.Self.MustValue().String()
}

func (self *SelfType) EqualTo(dst Type) bool {
	return ToActualType(self).EqualTo(dst)
}

// AliasType 别名类型
type AliasType = TypeAliasDef

func (self *AliasType) String() string {
	return stlbasic.Ternary(self.Pkg.Equal(BuildInPackage), self.Name, fmt.Sprintf("%s::%s", self.Pkg, self.Name))
}

func (self *AliasType) EqualTo(dst Type) bool {
	return ToActualType(self).EqualTo(dst)
}
