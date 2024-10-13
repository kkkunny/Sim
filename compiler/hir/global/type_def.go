package global

import (
	"fmt"

	stlbasic "github.com/kkkunny/stl/basic"

	"github.com/kkkunny/Sim/compiler/hir/types"
)

// TypeDef 类型定义
type TypeDef struct {
	pkgGlobalAttr
	name   string
	target types.Type
}

func (self *TypeDef) String() string {
	return stlbasic.Ternary(self.pkg.Equal(self.pkg.module.BuildinPkg), self.name, fmt.Sprintf("%s::%s", self.pkg.String(), self.name))
}

func (self *TypeDef) Equal(dst types.Type) bool {
	t, ok := dst.(types.CustomType)
	return ok && self.pkg.Equal(t.(Global).Package()) && self.name == t.Name()
}

func (self *TypeDef) Name() string {
	return self.name
}

func (self *TypeDef) Target() types.Type {
	return self.target
}

type CustomTypeWrap interface {
	types.CustomType
	Global
	Define() *TypeDef
}

func (self *TypeDef) Type() CustomTypeWrap {
	switch t := self.target.(type) {
	case types.SintType:
		return &customSintType{TypeDef: self, SintType: t}
	case types.UintType:
		return &customUintType{TypeDef: self, UintType: t}
	case types.FloatType:
		return &customFloatType{TypeDef: self, FloatType: t}
	case types.BoolType:
		return &customBoolType{TypeDef: self, BoolType: t}
	case types.StrType:
		return &customStrType{TypeDef: self, StrType: t}
	case types.RefType:
		return &customRefType{TypeDef: self, RefType: t}
	case types.ArrayType:
		return &customArrayType{TypeDef: self, ArrayType: t}
	case types.TupleType:
		return &customTupleType{TypeDef: self, TupleType: t}
	case types.FuncType:
		return &customFuncType{TypeDef: self, FuncType: t}
	case types.StructType:
		return &customStructType{TypeDef: self, StructType: t}
	case types.EnumType:
		return &customEnumType{TypeDef: self, EnumType: t}
	case types.LambdaType:
		return &customLambdaType{TypeDef: self, LambdaType: t}
	default:
		panic("unreachable")
	}
}

type customSintType struct {
	*TypeDef
	types.SintType
}

func (self *customSintType) Custom()          {}
func (self *customSintType) Define() *TypeDef { return self.TypeDef }

type customUintType struct {
	*TypeDef
	types.UintType
}

func (self *customUintType) Custom()          {}
func (self *customUintType) Define() *TypeDef { return self.TypeDef }

type customFloatType struct {
	*TypeDef
	types.FloatType
}

func (self *customFloatType) Custom()          {}
func (self *customFloatType) Define() *TypeDef { return self.TypeDef }

type customBoolType struct {
	*TypeDef
	types.BoolType
}

func (self *customBoolType) Custom()          {}
func (self *customBoolType) Define() *TypeDef { return self.TypeDef }

type customStrType struct {
	*TypeDef
	types.StrType
}

func (self *customStrType) Custom()          {}
func (self *customStrType) Define() *TypeDef { return self.TypeDef }

type customRefType struct {
	*TypeDef
	types.RefType
}

func (self *customRefType) Custom()          {}
func (self *customRefType) Define() *TypeDef { return self.TypeDef }

type customArrayType struct {
	*TypeDef
	types.ArrayType
}

func (self *customArrayType) Custom()          {}
func (self *customArrayType) Define() *TypeDef { return self.TypeDef }

type customTupleType struct {
	*TypeDef
	types.TupleType
}

func (self *customTupleType) Custom()          {}
func (self *customTupleType) Define() *TypeDef { return self.TypeDef }

type customFuncType struct {
	*TypeDef
	types.FuncType
}

func (self *customFuncType) Custom()          {}
func (self *customFuncType) Define() *TypeDef { return self.TypeDef }

type customStructType struct {
	*TypeDef
	types.StructType
}

func (self *customStructType) Custom()          {}
func (self *customStructType) Define() *TypeDef { return self.TypeDef }

type customEnumType struct {
	*TypeDef
	types.EnumType
}

func (self *customEnumType) Custom()          {}
func (self *customEnumType) Define() *TypeDef { return self.TypeDef }

type customLambdaType struct {
	*TypeDef
	types.LambdaType
}

func (self *customLambdaType) Custom()          {}
func (self *customLambdaType) Define() *TypeDef { return self.TypeDef }
