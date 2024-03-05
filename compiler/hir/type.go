package hir

import (
	"fmt"
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
	NoThing  = &NoThingType{}
	NoReturn = &NoReturnType{}
)

// Type 类型
// 运行时类型+编译时的工具类型
type Type interface {
	fmt.Stringer
	EqualTo(dst Type) bool
	HasDefault() bool
	Runtime() runtimeType.Type
}

// RuntimeType 可映射为运行时类型的类型
type RuntimeType interface {
	Type
	runtime()
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

// BuildInType 运行时类型去掉自定义类型
type BuildInType interface {
	RuntimeType
	buildin()
}

func ToBuildInType(t Type) BuildInType {
	switch tt := t.(type) {
	case BuildInType:
		return tt
	case *CustomType:
		return ToBuildInType(tt.Target)
	case *SelfType:
		return ToBuildInType(tt.Self.MustValue())
	case *AliasType:
		return ToBuildInType(tt.Target)
	default:
		panic("unreachable")
	}
}

type CallableType interface {
	BuildInType
	GetRet() Type
	GetParams() []Type
}

// TryType 尝试断言为指定类型
func TryType[T BuildInType](t Type) (T, bool) {
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
func IsType[T BuildInType](t Type) bool {
	_, ok := TryType[T](t)
	return ok
}

// AsType 断言为指定类型
func AsType[T BuildInType](t Type) T {
	at, ok := TryType[T](t)
	if !ok {
		panic("unreachable")
	}
	return at
}

func IsNumberType(t Type) bool { return IsIntType(t) || IsType[*FloatType](t) }
func IsIntType(t Type) bool    { return IsType[*SintType](t) || IsType[*UintType](t) }
func IsPointer(t Type) bool    { return IsType[*RefType](t) || IsType[*FuncType](t) }

// NoThingType 无返回值类型
type NoThingType struct{}

func (*NoThingType) String() string {
	return "void"
}

func (self *NoThingType) EqualTo(dst Type) bool {
	return stlbasic.Is[*NoThingType](ToRuntimeType(dst))
}

func (self *NoThingType) HasDefault() bool {
	return false
}

func (self *NoThingType) Runtime() runtimeType.Type {
	return runtimeType.TypeNoThing
}

func (self *NoThingType) buildin() {}
func (self *NoThingType) runtime() {}

// NoReturnType 空类型
type NoReturnType struct{}

func (*NoReturnType) String() string {
	return "X"
}

func (self *NoReturnType) EqualTo(dst Type) bool {
	return stlbasic.Is[*NoReturnType](ToRuntimeType(dst))
}

func (self *NoReturnType) HasDefault() bool {
	return false
}

func (self *NoReturnType) Runtime() runtimeType.Type {
	return runtimeType.TypeNoReturn
}

func (self *NoReturnType) buildin() {}
func (self *NoReturnType) runtime() {}

// SintType 有符号整型
type SintType struct {
	Bits uint8
}

func NewSintType(bits uint8) *SintType {
	return &SintType{Bits: bits}
}

func (self *SintType) String() string {
	return fmt.Sprintf("__buildin_i%d", self.Bits)
}

func (self *SintType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*SintType)
	if !ok {
		return false
	}
	return self.Bits == at.Bits
}

func (self *SintType) HasDefault() bool {
	return true
}

func (self *SintType) Runtime() runtimeType.Type {
	return runtimeType.NewSintType(self.Bits)
}

func (self *SintType) buildin() {}
func (self *SintType) runtime() {}

// UintType 无符号整型
type UintType struct {
	Bits uint8
}

func NewUintType(bits uint8) *UintType {
	return &UintType{Bits: bits}
}

func (self *UintType) String() string {
	return fmt.Sprintf("__buildin_u%d", self.Bits)
}

func (self *UintType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*UintType)
	if !ok {
		return false
	}
	return self.Bits == at.Bits
}

func (self *UintType) HasDefault() bool {
	return true
}

func (self *UintType) Runtime() runtimeType.Type {
	return runtimeType.NewUintType(self.Bits)
}

func (self *UintType) buildin() {}
func (self *UintType) runtime() {}

// FloatType 浮点型
type FloatType struct {
	Bits uint8
}

func NewFloatType(bits uint8) *FloatType {
	return &FloatType{Bits: bits}
}

func (self *FloatType) String() string {
	return fmt.Sprintf("__buildin_f%d", self.Bits)
}

func (self *FloatType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*FloatType)
	if !ok {
		return false
	}
	return self.Bits == at.Bits
}

func (self *FloatType) HasDefault() bool {
	return true
}

func (self *FloatType) Runtime() runtimeType.Type {
	return runtimeType.NewFloatType(self.Bits)
}

func (self *FloatType) buildin() {}
func (self *FloatType) runtime() {}

// FuncType 函数类型
type FuncType struct {
	Ret    Type
	Params []Type
}

func NewFuncType(ret Type, params ...Type) *FuncType {
	return &FuncType{
		Ret:    ret,
		Params: params,
	}
}

func (self *FuncType) String() string {
	params := stlslices.Map(self.Params, func(_ int, e Type) string { return e.String() })
	ret := stlbasic.Ternary(self.Ret.EqualTo(NoThing), "", self.Ret.String())
	return fmt.Sprintf("func(%s)%s", strings.Join(params, ", "), ret)
}

func (self *FuncType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*FuncType)
	if !ok {
		return false
	}
	if !self.Ret.EqualTo(at.Ret) || len(self.Params) != len(at.Params) {
		return false
	}
	for i, p := range self.Params {
		if !p.EqualTo(at.Params[i]) {
			return false
		}
	}
	return true
}

func (self *FuncType) HasDefault() bool {
	if self.Ret.EqualTo(NoThing) {
		return true
	}
	return self.Ret.HasDefault()
}

func (self *FuncType) Runtime() runtimeType.Type {
	return runtimeType.NewFuncType(self.Ret.Runtime(), stlslices.Map(self.Params, func(_ int, e Type) runtimeType.Type {
		return e.Runtime()
	})...)
}

func (self *FuncType) buildin() {}
func (self *FuncType) runtime() {}

func (self *FuncType) GetRet() Type {
	return self.Ret
}

func (self *FuncType) GetParams() []Type {
	return self.Params
}

// ArrayType 数组型
type ArrayType struct {
	Size uint64
	Elem Type
}

func NewArrayType(size uint64, elem Type) *ArrayType {
	return &ArrayType{
		Size: size,
		Elem: elem,
	}
}

func (self *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", self.Size, self.Elem)
}

func (self *ArrayType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*ArrayType)
	if !ok {
		return false
	}
	return self.Size == at.Size && self.Elem.EqualTo(at.Elem)
}

func (self *ArrayType) HasDefault() bool {
	return self.Elem.HasDefault()
}

func (self *ArrayType) Runtime() runtimeType.Type {
	return runtimeType.NewArrayType(self.Size, self.Elem.Runtime())
}

func (self *ArrayType) buildin() {}
func (self *ArrayType) runtime() {}

// TupleType 元组型
type TupleType struct {
	Elems []Type
}

func NewTupleType(elem ...Type) *TupleType {
	return &TupleType{
		Elems: elem,
	}
}

func (self *TupleType) String() string {
	elems := stlslices.Map(self.Elems, func(_ int, e Type) string { return e.String() })
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *TupleType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*TupleType)
	if !ok {
		return false
	}
	if len(self.Elems) != len(at.Elems) {
		return false
	}
	for i, e := range self.Elems {
		if !e.EqualTo(at.Elems[i]) {
			return false
		}
	}
	return true
}

func (self *TupleType) HasDefault() bool {
	return stlslices.All(self.Elems, func(_ int, e Type) bool {
		return e.HasDefault()
	})
}

func (self *TupleType) Runtime() runtimeType.Type {
	return runtimeType.NewTupleType(stlslices.Map(self.Elems, func(_ int, e Type) runtimeType.Type {
		return e.Runtime()
	})...)
}

func (self *TupleType) buildin() {}
func (self *TupleType) runtime() {}

// RefType 引用类型
type RefType struct {
	Mut  bool
	Elem Type
}

func NewRefType(mut bool, elem Type) *RefType {
	return &RefType{
		Mut:  mut,
		Elem: elem,
	}
}

func (self *RefType) String() string {
	return stlbasic.Ternary(self.Mut, "&mut ", "&") + self.Elem.String()
}

func (self *RefType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*RefType)
	if !ok {
		return false
	}
	return self.Mut == at.Mut && self.Elem.EqualTo(at.Elem)
}

func (self *RefType) HasDefault() bool {
	return false
}

func (self *RefType) Runtime() runtimeType.Type {
	return runtimeType.NewRefType(self.Mut, self.Elem.Runtime())
}

func (self *RefType) buildin() {}
func (self *RefType) runtime() {}

// Field 字段
type Field struct {
	Public  bool
	Mutable bool
	Name    string
	Type    Type
}

// StructType 结构体类型
type StructType struct {
	Def    *CustomType
	Fields linkedhashmap.LinkedHashMap[string, Field]
}

func NewStructType(def *CustomType, fields linkedhashmap.LinkedHashMap[string, Field]) *StructType {
	return &StructType{
		Def:    def,
		Fields: fields,
	}
}

func (self *StructType) String() string {
	return self.Def.String()
}

func (self *StructType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*StructType)
	if !ok {
		return false
	}
	return self.Def.EqualTo(at.Def)
}

func (self *StructType) HasDefault() bool {
	return stliter.All(self.Fields, func(e pair.Pair[string, Field]) bool {
		return e.Second.Type.HasDefault()
	})
}

func (self *StructType) Runtime() runtimeType.Type {
	fields := stliter.Map[pair.Pair[string, Field], runtimeType.Field, dynarray.DynArray[runtimeType.Field]](self.Fields, func(e pair.Pair[string, Field]) runtimeType.Field {
		return runtimeType.NewField(e.Second.Type.Runtime(), e.First)
	}).ToSlice()
	return runtimeType.NewStructType(self.Def.Pkg.String(), fields...)
}

func (self *StructType) buildin() {}
func (self *StructType) runtime() {}

// 替代所有Self类型
func replaceAllSelfType(t Type, to *CustomType) Type {
	switch tt := t.(type) {
	case *NoThingType, *NoReturnType, *SintType, *UintType, *FloatType, *CustomType, *AliasType:
		return tt
	case *RefType:
		return NewRefType(tt.Mut, replaceAllSelfType(tt.Elem, to))
	case *FuncType:
		return NewFuncType(replaceAllSelfType(tt.Ret, to), stlslices.Map(tt.Params, func(_ int, e Type) Type {
			return replaceAllSelfType(e, to)
		})...)
	case *ArrayType:
		return NewArrayType(tt.Size, replaceAllSelfType(tt.Elem, to))
	case *TupleType:
		return NewTupleType(stlslices.Map(tt.Elems, func(_ int, e Type) Type {
			return replaceAllSelfType(e, to)
		})...)
	case *StructType:
		return NewStructType(tt.Def, stliter.Map[pair.Pair[string, Field], pair.Pair[string, Field], linkedhashmap.LinkedHashMap[string, Field]](tt.Fields, func(e pair.Pair[string, Field]) pair.Pair[string, Field] {
			return pair.NewPair(e.First, Field{
				Public:  e.Second.Public,
				Mutable: e.Second.Mutable,
				Name:    e.Second.Name,
				Type:    replaceAllSelfType(e.Second.Type, to),
			})
		}))
	case *SelfType:
		return stlbasic.Ternary[Type](tt.Self.IsNone(), NewSelfType(util.Some[*CustomType](to)), tt)
	case *LambdaType:
		return NewLambdaType(replaceAllSelfType(tt.Ret, to), stlslices.Map(tt.Params, func(_ int, e Type) Type {
			return replaceAllSelfType(e, to)
		})...)
	default:
		panic("unreachable")
	}
}

// SelfType Self类型
type SelfType struct {
	Self util.Option[*CustomType]
}

func NewSelfType(self util.Option[*CustomType]) *SelfType {
	return &SelfType{
		Self: self,
	}
}

func (self *SelfType) String() string {
	return self.Self.MustValue().String()
}

func (self *SelfType) EqualTo(dst Type) bool {
	return self.Self.MustValue().EqualTo(dst)
}

func (self *SelfType) HasDefault() bool {
	return self.Self.MustValue().HasDefault()
}

func (self *SelfType) Runtime() runtimeType.Type {
	return self.Self.MustValue().Runtime()
}

// CustomType 自定义类型
type CustomType = TypeDef

func TryCustomType(t Type) (*CustomType, bool) {
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

func IsCustomType(t Type) bool {
	_, ok := TryCustomType(t)
	return ok
}

func AsCustomType(t Type) *CustomType {
	ct, ok := TryCustomType(t)
	if !ok {
		panic("unreachable")
	}
	return ct
}

func (self *CustomType) String() string {
	return GetGlobalName(self.Pkg, self.Name)
}

func (self *CustomType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*CustomType)
	if !ok {
		return false
	}
	return self.Pkg == at.Pkg && self.Name == at.Name
}

func (self *CustomType) HasDefault() bool {
	return self.Target.HasDefault()
}

var customTypeRuntimeCache = make(map[*CustomType]runtimeType.Type)

func (self *CustomType) Runtime() runtimeType.Type {
	cache, ok := customTypeRuntimeCache[self]
	if ok {
		return cache
	}
	rt := runtimeType.NewCustomType(self.Pkg.String(), self.Name, nil, nil)
	customTypeRuntimeCache[self] = rt
	rt.Target = self.Target.Runtime()
	rt.Methods = stliter.Map[pair.Pair[string, *MethodDef], runtimeType.Method, dynarray.DynArray[runtimeType.Method]](self.Methods, func(e pair.Pair[string, *MethodDef]) runtimeType.Method {
		return runtimeType.NewMethod(e.Second.GetFuncType().Runtime().(*runtimeType.FuncType), e.First)
	}).ToSlice()
	return rt
}

func (self *CustomType) runtime() {}

// AliasType 别名类型
type AliasType = TypeAliasDef

func (self *AliasType) String() string {
	return self.Target.String()
}

func (self *AliasType) EqualTo(dst Type) bool {
	return self.Target.EqualTo(dst)
}

func (self *AliasType) HasDefault() bool {
	return self.Target.HasDefault()
}

func (self *AliasType) Runtime() runtimeType.Type {
	return self.Target.Runtime()
}

// LambdaType 匿名函数类型
type LambdaType struct {
	Ret    Type
	Params []Type
}

func NewLambdaType(ret Type, params ...Type) *LambdaType {
	return &LambdaType{
		Ret:    ret,
		Params: params,
	}
}

func (self *LambdaType) String() string {
	params := stlslices.Map(self.Params, func(_ int, e Type) string { return e.String() })
	return fmt.Sprintf("(%s)->%s", strings.Join(params, ", "), self.Ret)
}

func (self *LambdaType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*LambdaType)
	if !ok {
		return false
	}
	if !self.Ret.EqualTo(at.Ret) || len(self.Params) != len(at.Params) {
		return false
	}
	for i, p := range self.Params {
		if !p.EqualTo(at.Params[i]) {
			return false
		}
	}
	return true
}

func (self *LambdaType) HasDefault() bool {
	if self.Ret.EqualTo(NoThing) {
		return true
	}
	return self.Ret.HasDefault()
}

func (self *LambdaType) Runtime() runtimeType.Type {
	return runtimeType.NewLambdaType(self.Ret.Runtime(), stlslices.Map(self.Params, func(_ int, e Type) runtimeType.Type {
		return e.Runtime()
	})...)
}

func (self *LambdaType) buildin() {}
func (self *LambdaType) runtime() {}

func (self *LambdaType) ToFuncType() *FuncType {
	return NewFuncType(self.Ret, self.Params...)
}

func (self *LambdaType) GetRet() Type {
	return self.Ret
}

func (self *LambdaType) GetParams() []Type {
	return self.Params
}

// EnumField 枚举字段
type EnumField struct {
	Name  string
	Elems []Type
}

// EnumType 枚举类型
type EnumType struct {
	Def    *CustomType
	Fields linkedhashmap.LinkedHashMap[string, EnumField]
}

func NewEnumType(def *CustomType, fields linkedhashmap.LinkedHashMap[string, EnumField]) *EnumType {
	return &EnumType{
		Def:    def,
		Fields: fields,
	}
}

func (self *EnumType) String() string {
	return self.Def.String()
}

func (self *EnumType) EqualTo(dst Type) bool {
	at, ok := ToRuntimeType(dst).(*EnumType)
	if !ok {
		return false
	}
	return self.Def.EqualTo(at.Def)
}

func (self *EnumType) HasDefault() bool {
	return false
}

func (self *EnumType) Runtime() runtimeType.Type {
	fields := stliter.Map[pair.Pair[string, EnumField], runtimeType.EnumField, dynarray.DynArray[runtimeType.EnumField]](self.Fields, func(e pair.Pair[string, EnumField]) runtimeType.EnumField {
		return runtimeType.NewEnumField(e.Second.Name, stlslices.Map(e.Second.Elems, func(_ int, e Type) runtimeType.Type {
			return e.Runtime()
		})...)
	}).ToSlice()
	return runtimeType.NewEnumType(self.Def.Pkg.String(), fields...)
}

func (self *EnumType) buildin() {}
func (self *EnumType) runtime() {}

// IsSimple 是否是简单枚举
func (self *EnumType) IsSimple() bool {
	return stliter.All(self.Fields, func(e pair.Pair[string, EnumField]) bool {
		return len(e.Second.Elems) == 0
	})
}

// GenericParam 泛型参数
type GenericParam struct {
	Belong *GenericFuncDef
	Name   string
}

func NewGenericParam(belong *GenericFuncDef, name string) *GenericParam {
	return &GenericParam{
		Belong: belong,
		Name:   name,
	}
}

func (self *GenericParam) String() string {
	return fmt.Sprintf("%s::%s", GetGlobalName(self.Belong.Pkg, self.Belong.Name), self.Name)
}

func (self *GenericParam) EqualTo(dst Type) bool {
	gp, ok := dst.(*GenericParam)
	if !ok {
		return false
	}
	return self.Belong == gp.Belong && self.Name == gp.Name
}

func (self *GenericParam) HasDefault() bool {
	return false
}

func (self *GenericParam) Runtime() runtimeType.Type {
	panic("unreachable")
}
