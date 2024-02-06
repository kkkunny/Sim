package hir

import (
	"fmt"
	"sort"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/dynarray"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedhashmap"
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
// 运行时类型+编译时的工具类型
type Type interface {
	fmt.Stringer
	EqualTo(dst Type) bool
	HasDefault()bool
	Runtime()runtimeType.Type
}

// RuntimeType 可映射为运行时类型的类型
type RuntimeType interface {
	Type
	runtime()
}

// BuildInType 运行时类型去掉自定义类型
type BuildInType interface {
	RuntimeType
	buildin()
}

func ToRuntimeType(t Type) RuntimeType {
	switch tt := t.(type) {
	case RuntimeType:
		return tt
	case *SelfType:
		return ToRuntimeType(tt.Self.MustValue())
	case *AliasType:
		return ToRuntimeType(tt.Target)
	default:
		panic("unreachable")
	}
}

// TryType 尝试断言为指定类型
func TryType[T BuildInType](t Type)(T, bool){
	switch tt := t.(type) {
	case T:
		return tt, true
	case BuildInType:
		return stlbasic.Default[T](), false
	case *SelfType:
		return TryType[T](tt.Self.MustValue())
	case *AliasType:
		return TryType[T](tt.Target)
	case *CustomType:
		return TryType[T](tt.Target)
	default:
		panic("unreachable")
	}
}

// IsType 是否可以断言为指定类型
func IsType[T BuildInType](t Type)bool{
	_, ok := TryType[T](t)
	return ok
}

// AsType 断言为指定类型
func AsType[T BuildInType](t Type)T{
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
	return ""
}

func (self *EmptyType) EqualTo(dst Type) bool {
	return IsType[*EmptyType](dst)
}

func (self *EmptyType) HasDefault()bool{
	return false
}

func (self *EmptyType) Runtime()runtimeType.Type{
	return runtimeType.TypeEmpty
}

func (self *EmptyType) buildin(){}
func (self *EmptyType) runtime(){}

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

func (self *SintType) HasDefault()bool{
	return true
}

func (self *SintType) Runtime()runtimeType.Type{
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

func (self *SintType) buildin(){}
func (self *SintType) runtime(){}

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

func (self *UintType) HasDefault()bool{
	return true
}

func (self *UintType) Runtime()runtimeType.Type{
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

func (self *UintType) buildin(){}
func (self *UintType) runtime(){}

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

func (self *FloatType) HasDefault()bool{
	return true
}

func (self *FloatType) Runtime()runtimeType.Type{
	switch self.Bits {
	case 32:
		return runtimeType.TypeF32
	case 64:
		return runtimeType.TypeF64
	default:
		panic("unreachable")
	}
}

func (self *FloatType) buildin(){}
func (self *FloatType) runtime(){}

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
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), self.Ret)
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

func (self *FuncType) HasDefault()bool{
	if self.Ret.EqualTo(Empty){
		return true
	}
	return self.Ret.HasDefault()
}

func (self *FuncType) Runtime()runtimeType.Type{
	return runtimeType.NewFuncType(self.Ret.Runtime(), stlslices.Map(self.Params, func(_ int, e Type) runtimeType.Type {
		return e.Runtime()
	})...)
}

func (self *FuncType) buildin(){}
func (self *FuncType) runtime(){}

// BoolType 布尔型
type BoolType struct{}

func (*BoolType) String() string {
	return "bool"
}

func (self *BoolType) EqualTo(dst Type) bool {
	return IsType[*BoolType](dst)
}

func (self *BoolType) HasDefault()bool{
	return true
}

func (self *BoolType) Runtime()runtimeType.Type{
	return runtimeType.TypeBool
}

func (self *BoolType) buildin(){}
func (self *BoolType) runtime(){}

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

func (self *ArrayType) HasDefault()bool{
	return self.Elem.HasDefault()
}

func (self *ArrayType) Runtime()runtimeType.Type{
	return runtimeType.NewArrayType(self.Size, self.Elem.Runtime())
}

func (self *ArrayType) buildin(){}
func (self *ArrayType) runtime(){}

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

func (self *TupleType) HasDefault()bool{
	return stlslices.All(self.Elems, func(_ int, e Type) bool {
		return e.HasDefault()
	})
}

func (self *TupleType) Runtime()runtimeType.Type{
	return runtimeType.NewTupleType(stlslices.Map(self.Elems, func(_ int, e Type) runtimeType.Type {
		return e.Runtime()
	})...)
}

func (self *TupleType) buildin(){}
func (self *TupleType) runtime(){}

// StringType 字符串型
type StringType struct{}

func (*StringType) String() string {
	return "str"
}

func (self *StringType) EqualTo(dst Type) bool {
	return IsType[*StringType](dst)
}

func (self *StringType) HasDefault()bool{
	return true
}

func (self *StringType) Runtime()runtimeType.Type{
	return runtimeType.TypeStr
}

func (self *StringType) buildin(){}
func (self *StringType) runtime(){}

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

func (self *UnionType) HasDefault()bool{
	return self.Elems[0].HasDefault()
}

func (self *UnionType) Runtime()runtimeType.Type{
	return runtimeType.NewUnionType(stlslices.Map(self.Elems, func(_ int, e Type) runtimeType.Type {
		return e.Runtime()
	})...)
}

func (self *UnionType) buildin(){}
func (self *UnionType) runtime(){}

// IndexElem 取元素下标
func (self *UnionType) IndexElem(t Type)int{
	for i, elem := range self.Elems{
		if elem.EqualTo(t){
			return i
		}
	}
	return -1
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
	return stlbasic.Ternary(self.Mut, "&mut ", "&") + self.Elem.String()
}

func (self *RefType) EqualTo(dst Type) bool {
	at, ok := TryType[*RefType](dst)
	if !ok{
		return false
	}
	return self.Mut == at.Mut && self.Elem.EqualTo(at.Elem)
}

func (self *RefType) HasDefault()bool{
	return false
}

func (self *RefType) Runtime()runtimeType.Type{
	return runtimeType.NewRefType(self.Mut, self.Elem.Runtime())
}

func (self *RefType) buildin(){}
func (self *RefType) runtime(){}

// Field 字段
type Field struct {
	Public  bool
	Mutable bool
	Name string
	Type    Type
}

// StructType 结构体类型
type StructType struct {
	Pkg Package
	Fields  linkedhashmap.LinkedHashMap[string, Field]
}

func NewStructType(pkg Package, fields linkedhashmap.LinkedHashMap[string, Field])*StructType{
	return &StructType{
		Pkg: pkg,
		Fields: fields,
	}
}

func (self *StructType) String() string {
	return fmt.Sprintf("{%s}", strings.Join(stliter.Map[pair.Pair[string, Field], string, dynarray.DynArray[string]](self.Fields, func(e pair.Pair[string, Field]) string {
		return fmt.Sprintf("%s: %s", e.Second.Name, e.Second.Type.String())
	}).ToSlice(), ", "))
}

func (self *StructType) EqualTo(dst Type) bool {
	at, ok := TryType[*StructType](dst)
	if !ok || !self.Pkg.Equal(at.Pkg){
		return false
	}
	for iter1, iter2:=self.Fields.Iterator(), at.Fields.Iterator(); iter1.Next()&&iter2.Next(); {
		f1, f2 := iter1.Value().Second, iter2.Value().Second
		if f1.Public != f2.Public || f1.Mutable != f2.Mutable || f1.Name != f2.Name || !f1.Type.EqualTo(f2.Type){
			return false
		}
	}
	return true
}

func (self *StructType) HasDefault()bool{
	return stliter.All(self.Fields, func(e pair.Pair[string, Field]) bool {
		return e.Second.Type.HasDefault()
	})
}

func (self *StructType) Runtime()runtimeType.Type{
	fields := stliter.Map[pair.Pair[string, Field], runtimeType.Field, dynarray.DynArray[runtimeType.Field]](self.Fields, func(e pair.Pair[string, Field]) runtimeType.Field {
		return runtimeType.NewField(e.Second.Type.Runtime(), e.First)
	}).ToSlice()
	return runtimeType.NewStructType(self.Pkg.String(), fields...)
}

func (self *StructType) buildin(){}
func (self *StructType) runtime(){}

// SelfType Self类型
type SelfType struct{
	Self util.Option[GlobalType]
}

func NewSelfType(self GlobalType)*SelfType{
	return &SelfType{
		Self: util.Some(self),
	}
}

func (self *SelfType) String() string {
	return self.Self.MustValue().String()
}

func (self *SelfType) EqualTo(dst Type) bool {
	return ToRuntimeType(self).EqualTo(dst)
}

func (self *SelfType) HasDefault()bool{
	return self.Self.MustValue().HasDefault()
}

func (self *SelfType) Runtime()runtimeType.Type{
	return self.Self.MustValue().Runtime()
}

// CustomType 自定义类型
type CustomType = TypeDef

func TryCustomType(t Type)(*CustomType, bool){
	switch tt := t.(type) {
	case *CustomType:
		return tt, true
	case BuildInType:
		return nil, false
	case *SelfType:
		return TryCustomType(tt.Self.MustValue())
	case *AliasType:
		return TryCustomType(tt.Target)
	default:
		panic("unreachable")
	}
}

func IsCustomType(t Type)bool{
	_, ok := TryCustomType(t)
	return ok
}

func AsCustomType(t Type)*CustomType{
	ct, ok := TryCustomType(t)
	if !ok{
		panic("unreachable")
	}
	return ct
}

func (self *CustomType) String() string {
	return stlbasic.Ternary(self.Pkg.Equal(BuildInPackage), self.Name, fmt.Sprintf("%s::%s", self.Pkg, self.Name))
}

func (self *CustomType) EqualTo(dst Type) bool {
	at, ok := TryCustomType(dst)
	if !ok{
		return false
	}
	return self.Pkg == at.Pkg && self.Name == at.Name
}

func (self *CustomType) HasDefault()bool{
	return self.Target.HasDefault()
}

func (self *CustomType) Runtime()runtimeType.Type{
	methods := stliter.Map[pair.Pair[string, GlobalMethod], runtimeType.Method, dynarray.DynArray[runtimeType.Method]](self.Methods, func(e pair.Pair[string, GlobalMethod]) runtimeType.Method {
		return runtimeType.NewMethod(e.Second.GetFuncType().Runtime().(*runtimeType.FuncType), e.First)
	}).ToSlice()
	return runtimeType.NewCustomType(self.Pkg.String(), self.Name, self.Target.Runtime(), methods)
}

func (self *CustomType) runtime(){}

// AliasType 别名类型
type AliasType = TypeAliasDef

func (self *AliasType) String() string {
	return self.Target.String()
}

func (self *AliasType) EqualTo(dst Type) bool {
	return self.Target.EqualTo(dst)
}

func (self *AliasType) HasDefault()bool{
	return self.Target.HasDefault()
}

func (self *AliasType) Runtime()runtimeType.Type{
	return self.Target.Runtime()
}
