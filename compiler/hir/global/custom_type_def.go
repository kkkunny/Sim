package global

import (
	"fmt"
	"unsafe"

	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// CustomTypeDef 自定义类型定义
type CustomTypeDef interface {
	types.CustomType
	AddMethod(m MethodDef) bool
	GetMethod(name string) (MethodDef, bool)
	HasImpl(trait *Trait) bool
	CompilerParams() []types.GenericParamType
}

type _CustomTypeDef_ struct {
	pkgGlobalAttr
	name          string
	compileParams []types.GenericParamType
	target        hir.Type
	methods       hashmap.HashMap[string, MethodDef]
}

func NewCustomTypeDef(name string, compileParams []types.GenericParamType, target hir.Type) CustomTypeDef {
	return &_CustomTypeDef_{
		name:          name,
		compileParams: compileParams,
		target:        target,
		methods:       hashmap.StdWith[string, MethodDef](),
	}
}

func (self *_CustomTypeDef_) String() string {
	if self.pkg.IsBuildIn() {
		return self.name
	}
	return fmt.Sprintf("%s::%s", self.pkg.String(), self.name)
}

func (self *_CustomTypeDef_) Equal(dst hir.Type) bool {
	return self == dst
}

func (self *_CustomTypeDef_) GetName() (string, bool) {
	return self.name, self.name != "_"
}

func (self *_CustomTypeDef_) Target() hir.Type {
	return self.target
}

func (self *_CustomTypeDef_) SetTarget(t hir.Type) {
	self.target = t
}

func (self *_CustomTypeDef_) Custom() {

}

func (self *_CustomTypeDef_) AddMethod(m MethodDef) bool {
	mn, ok := m.GetName()
	if !ok {
		return true
	}
	if types.Is[types.StructType](self.target) && stlval.IgnoreWith(types.As[types.StructType](self.target)).Fields().Contain(mn) {
		return false
	}
	if self.methods.Contain(mn) {
		return false
	}
	self.methods.Set(mn, m)
	return true
}

func (self *_CustomTypeDef_) GetMethod(name string) (MethodDef, bool) {
	method := self.methods.Get(name)
	return method, method != nil
}

func (self *_CustomTypeDef_) HasImpl(trait *Trait) bool {
	return stlslices.All(trait.Methods(), func(_ int, dstF *FuncDecl) bool {
		method, ok := self.GetMethod(stlval.IgnoreWith(dstF.GetName()))
		if !ok {
			return false
		}
		return method.Type().Equal(types.ReplaceVirtualType(hashmap.AnyWith[types.VirtualType, hir.Type](types.Self, self), dstF.Type()))
	})
}

func (self *_CustomTypeDef_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}

func (self *_CustomTypeDef_) CompilerParams() []types.GenericParamType {
	return self.compileParams
}

func (self *_CustomTypeDef_) Wrap(inner hir.Type) hir.BuildInType {
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

func (self *customSintType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customSintType) String() string { return self.CustomTypeDef.String() }
func (self *customSintType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customUintType struct {
	CustomTypeDef
	types.UintType
}

func (self *customUintType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customUintType) String() string { return self.CustomTypeDef.String() }
func (self *customUintType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customFloatType struct {
	CustomTypeDef
	types.FloatType
}

func (self *customFloatType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customFloatType) String() string { return self.CustomTypeDef.String() }
func (self *customFloatType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customBoolType struct {
	CustomTypeDef
	types.BoolType
}

func (self *customBoolType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customBoolType) String() string { return self.CustomTypeDef.String() }
func (self *customBoolType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customStrType struct {
	CustomTypeDef
	types.StrType
}

func (self *customStrType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customStrType) String() string { return self.CustomTypeDef.String() }
func (self *customStrType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customRefType struct {
	CustomTypeDef
	types.RefType
}

func (self *customRefType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customRefType) String() string { return self.CustomTypeDef.String() }
func (self *customRefType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customArrayType struct {
	CustomTypeDef
	types.ArrayType
}

func (self *customArrayType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customArrayType) String() string { return self.CustomTypeDef.String() }
func (self *customArrayType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customTupleType struct {
	CustomTypeDef
	types.TupleType
}

func (self *customTupleType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customTupleType) String() string { return self.CustomTypeDef.String() }
func (self *customTupleType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customFuncType struct {
	CustomTypeDef
	types.FuncType
}

func (self *customFuncType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customFuncType) String() string { return self.CustomTypeDef.String() }
func (self *customFuncType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customLambdaType struct {
	CustomTypeDef
	types.LambdaType
}

func (self *customLambdaType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customLambdaType) String() string { return self.CustomTypeDef.String() }
func (self *customLambdaType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customStructType struct {
	CustomTypeDef
	types.StructType
}

func (self *customStructType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customStructType) String() string { return self.CustomTypeDef.String() }
func (self *customStructType) Hash() uint64   { return self.CustomTypeDef.Hash() }

type customEnumType struct {
	CustomTypeDef
	types.EnumType
}

func (self *customEnumType) Equal(dst hir.Type) bool {
	return self.CustomTypeDef.Equal(dst)
}
func (self *customEnumType) String() string { return self.CustomTypeDef.String() }
func (self *customEnumType) Hash() uint64   { return self.CustomTypeDef.Hash() }
