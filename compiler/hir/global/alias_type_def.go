package global

import (
	"fmt"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir/types"
)

// AliasTypeDef 类型别名定义
type AliasTypeDef interface {
	TypeDef
	types.AliasType
	Wrap(inner types.Type) types.BuildInType
}

type __AliasTypeDef__ struct {
	pkgGlobalAttr
	name   string
	target types.Type
}

func NewAliasTypeDef(name string, target types.Type) AliasTypeDef {
	return &__AliasTypeDef__{
		name:   name,
		target: target,
	}
}

func (self *__AliasTypeDef__) String() string {
	if self.pkg.IsBuildIn() {
		return self.name
	}
	return fmt.Sprintf("%s::%s", self.pkg.String(), self.name)
}

func (self *__AliasTypeDef__) Equal(dst types.Type, selfs ...types.Type) bool {
	if dst.Equal(types.Self) && len(selfs) > 0 {
		dst = stlslices.Last(selfs)
	}

	t, ok := dst.(*__AliasTypeDef__)
	if ok {
		return self.pkg.Equal(t.pkg) && self.name == t.name
	} else {
		return self.target.Equal(dst, selfs...)
	}
}

func (self *__AliasTypeDef__) Name() string {
	return self.name
}

func (self *__AliasTypeDef__) Target() types.Type {
	return self.target
}

func (self *__AliasTypeDef__) SetTarget(t types.Type) {
	self.target = t
}

func (self *__AliasTypeDef__) Alias() {

}

func (self *__AliasTypeDef__) Define() TypeDef {
	return self
}

func (self *__AliasTypeDef__) Wrap(inner types.Type) types.BuildInType {
	switch v := inner.(type) {
	case types.SintType:
		return &aliasSintType{AliasTypeDef: self, SintType: v}
	case types.UintType:
		return &aliasUintType{AliasTypeDef: self, UintType: v}
	case types.FloatType:
		return &aliasFloatType{AliasTypeDef: self, FloatType: v}
	case types.BoolType:
		return &aliasBoolType{AliasTypeDef: self, BoolType: v}
	case types.StrType:
		return &aliasStrType{AliasTypeDef: self, StrType: v}
	case types.RefType:
		return &aliasRefType{AliasTypeDef: self, RefType: v}
	case types.ArrayType:
		return &aliasArrayType{AliasTypeDef: self, ArrayType: v}
	case types.TupleType:
		return &aliasTupleType{AliasTypeDef: self, TupleType: v}
	case types.FuncType:
		return &aliasFuncType{AliasTypeDef: self, FuncType: v}
	case types.LambdaType:
		return &aliasLambdaType{AliasTypeDef: self, LambdaType: v}
	case types.StructType:
		return &aliasStructType{AliasTypeDef: self, StructType: v}
	case types.EnumType:
		return &aliasEnumType{AliasTypeDef: self, EnumType: v}
	default:
		panic("unreachable")
	}
}

type aliasSintType struct {
	AliasTypeDef
	types.SintType
}

func (self *aliasSintType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasSintType) String() string { return self.AliasTypeDef.String() }

type aliasUintType struct {
	AliasTypeDef
	types.UintType
}

func (self *aliasUintType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasUintType) String() string { return self.AliasTypeDef.String() }

type aliasFloatType struct {
	AliasTypeDef
	types.FloatType
}

func (self *aliasFloatType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasFloatType) String() string { return self.AliasTypeDef.String() }

type aliasBoolType struct {
	AliasTypeDef
	types.BoolType
}

func (self *aliasBoolType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasBoolType) String() string { return self.AliasTypeDef.String() }

type aliasStrType struct {
	AliasTypeDef
	types.StrType
}

func (self *aliasStrType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasStrType) String() string { return self.AliasTypeDef.String() }

type aliasRefType struct {
	AliasTypeDef
	types.RefType
}

func (self *aliasRefType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasRefType) String() string { return self.AliasTypeDef.String() }

type aliasArrayType struct {
	AliasTypeDef
	types.ArrayType
}

func (self *aliasArrayType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasArrayType) String() string { return self.AliasTypeDef.String() }

type aliasTupleType struct {
	AliasTypeDef
	types.TupleType
}

func (self *aliasTupleType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasTupleType) String() string { return self.AliasTypeDef.String() }

type aliasFuncType struct {
	AliasTypeDef
	types.FuncType
}

func (self *aliasFuncType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasFuncType) String() string { return self.AliasTypeDef.String() }

type aliasLambdaType struct {
	AliasTypeDef
	types.LambdaType
}

func (self *aliasLambdaType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasLambdaType) String() string { return self.AliasTypeDef.String() }

type aliasStructType struct {
	AliasTypeDef
	types.StructType
}

func (self *aliasStructType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasStructType) String() string { return self.AliasTypeDef.String() }

type aliasEnumType struct {
	AliasTypeDef
	types.EnumType
}

func (self *aliasEnumType) Equal(dst types.Type, selfs ...types.Type) bool {
	return self.AliasTypeDef.Equal(dst, selfs...)
}
func (self *aliasEnumType) String() string { return self.AliasTypeDef.String() }
