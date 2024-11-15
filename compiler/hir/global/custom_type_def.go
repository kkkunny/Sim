package global

import (
	"fmt"
	"unsafe"

	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir/types"
)

// CustomTypeDef 自定义类型定义
type CustomTypeDef interface {
	TypeDef
	types.CustomType
	AddMethod(m *MethodDef) bool
	GetMethod(name string) (*MethodDef, bool)
	HasImpl(trait *Trait) bool
	Wrap(inner types.Type) types.BuildInType
}

type __CustomTypeDef__ struct {
	pkgGlobalAttr
	name    string
	target  types.Type
	methods hashmap.HashMap[string, *MethodDef]
}

func NewCustomTypeDef(name string, target types.Type) CustomTypeDef {
	return &__CustomTypeDef__{
		name:    name,
		target:  target,
		methods: hashmap.StdWith[string, *MethodDef](),
	}
}

func (self *__CustomTypeDef__) String() string {
	if self.pkg.IsBuildIn() {
		return self.name
	}
	return fmt.Sprintf("%s::%s", self.pkg.String(), self.name)
}

func (self *__CustomTypeDef__) Equal(dst types.Type) bool {
	t, ok := dst.(CustomTypeDef)
	return ok && self.pkg.Equal(t.Package()) && self.name == t.Name()
}

func (self *__CustomTypeDef__) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	if dst.Equal(types.Self) && len(selfs) > 0 {
		dst = stlslices.Last(selfs)
	}

	t, ok := dst.(CustomTypeDef)
	return ok && self.pkg.Equal(t.Package()) && self.name == t.Name()
}

func (self *__CustomTypeDef__) Name() string {
	return self.name
}

func (self *__CustomTypeDef__) Target() types.Type {
	return self.target
}

func (self *__CustomTypeDef__) SetTarget(t types.Type) {
	self.target = t
}

func (self *__CustomTypeDef__) Custom() {

}

func (self *__CustomTypeDef__) AddMethod(m *MethodDef) bool {
	if types.Is[types.StructType](self.target) && stlval.IgnoreWith(types.As[types.StructType](self.target)).Fields().Contain(m.Name()) {
		return false
	}
	if self.methods.Contain(m.Name()) {
		return false
	}
	self.methods.Set(m.Name(), m)
	return true
}

func (self *__CustomTypeDef__) GetMethod(name string) (*MethodDef, bool) {
	method := self.methods.Get(name)
	return method, method != nil
}

func (self *__CustomTypeDef__) HasImpl(trait *Trait) bool {
	return stlslices.All(trait.Methods.Values(), func(_ int, dstF *FuncDecl) bool {
		method, ok := self.GetMethod(dstF.Name())
		if !ok {
			return false
		}
		return method.Type().EqualWithSelf(dstF.Type(), self)
	})
}

func (self *__CustomTypeDef__) Define() TypeDef {
	return self
}

func (self *__CustomTypeDef__) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}

func (self *__CustomTypeDef__) Wrap(inner types.Type) types.BuildInType {
	switch v := inner.(type) {
	case types.SintType:
		return &customSintType{CustomTypeDef: self, SintType: v}
	case types.UintType:
		return &customUintType{CustomTypeDef: self, UintType: v}
	case types.FloatType:
		return &customFloatType{CustomTypeDef: self, FloatType: v}
	case types.BoolType:
		return &customBoolType{CustomTypeDef: self, BoolType: v}
	case types.StrType:
		return &customStrType{CustomTypeDef: self, StrType: v}
	case types.RefType:
		return &customRefType{CustomTypeDef: self, RefType: v}
	case types.ArrayType:
		return &customArrayType{CustomTypeDef: self, ArrayType: v}
	case types.TupleType:
		return &customTupleType{CustomTypeDef: self, TupleType: v}
	case types.FuncType:
		return &customFuncType{CustomTypeDef: self, FuncType: v}
	case types.LambdaType:
		return &customLambdaType{CustomTypeDef: self, LambdaType: v}
	case types.StructType:
		return &customStructType{CustomTypeDef: self, StructType: v}
	case types.EnumType:
		return &customEnumType{CustomTypeDef: self, EnumType: v}
	default:
		panic("unreachable")
	}
}

type customSintType struct {
	CustomTypeDef
	types.SintType
}

func (self *customSintType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customSintType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customSintType) String() string { return self.CustomTypeDef.String() }
func (self *customSintType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customUintType struct {
	CustomTypeDef
	types.UintType
}

func (self *customUintType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customUintType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customUintType) String() string { return self.CustomTypeDef.String() }
func (self *customUintType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customFloatType struct {
	CustomTypeDef
	types.FloatType
}

func (self *customFloatType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customFloatType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customFloatType) String() string { return self.CustomTypeDef.String() }
func (self *customFloatType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customBoolType struct {
	CustomTypeDef
	types.BoolType
}

func (self *customBoolType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customBoolType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customBoolType) String() string { return self.CustomTypeDef.String() }
func (self *customBoolType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customStrType struct {
	CustomTypeDef
	types.StrType
}

func (self *customStrType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customStrType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customStrType) String() string { return self.CustomTypeDef.String() }
func (self *customStrType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customRefType struct {
	CustomTypeDef
	types.RefType
}

func (self *customRefType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customRefType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customRefType) String() string { return self.CustomTypeDef.String() }
func (self *customRefType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customArrayType struct {
	CustomTypeDef
	types.ArrayType
}

func (self *customArrayType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customArrayType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customArrayType) String() string { return self.CustomTypeDef.String() }
func (self *customArrayType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customTupleType struct {
	CustomTypeDef
	types.TupleType
}

func (self *customTupleType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customTupleType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customTupleType) String() string { return self.CustomTypeDef.String() }
func (self *customTupleType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customFuncType struct {
	CustomTypeDef
	types.FuncType
}

func (self *customFuncType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customFuncType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customFuncType) String() string { return self.CustomTypeDef.String() }
func (self *customFuncType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customLambdaType struct {
	CustomTypeDef
	types.LambdaType
}

func (self *customLambdaType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customLambdaType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customLambdaType) String() string { return self.CustomTypeDef.String() }
func (self *customLambdaType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customStructType struct {
	CustomTypeDef
	types.StructType
}

func (self *customStructType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customStructType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customStructType) String() string { return self.CustomTypeDef.String() }
func (self *customStructType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customEnumType struct {
	CustomTypeDef
	types.EnumType
}

func (self *customEnumType) Equal(dst types.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customEnumType) EqualWithSelf(dst types.Type, selfs ...types.Type) bool {
	return self.CustomTypeDef.EqualWithSelf(dst, selfs...)
}
func (self *customEnumType) String() string { return self.CustomTypeDef.String() }
func (self *customEnumType) Hash() uint64   { return self.CustomTypeDef.Hash() }
