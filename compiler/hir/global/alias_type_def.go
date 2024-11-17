package global

import (
	"fmt"
	"unsafe"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// AliasTypeDef 类型别名定义
type AliasTypeDef interface {
	types.AliasType
}

type _AliasTypeDef_ struct {
	pkgGlobalAttr
	name   string
	target hir.Type
}

func NewAliasTypeDef(name string, target hir.Type) AliasTypeDef {
	return &_AliasTypeDef_{
		name:   name,
		target: target,
	}
}

func (self *_AliasTypeDef_) String() string {
	if self.pkg.IsBuildIn() {
		return self.name
	}
	return fmt.Sprintf("%s::%s", self.pkg.String(), self.name)
}

func (self *_AliasTypeDef_) Equal(dst hir.Type) bool {
	if self.Hash() == dst.Hash() {
		return true
	}

	t, ok := dst.(AliasTypeDef)
	if ok {
		return self.target.Equal(t.Target())
	} else {
		return self.target.Equal(dst)
	}
}

func (self *_AliasTypeDef_) GetName() (string, bool) {
	return self.name, self.name != "_"
}

func (self *_AliasTypeDef_) Target() hir.Type {
	return self.target
}

func (self *_AliasTypeDef_) SetTarget(t hir.Type) {
	self.target = t
}

func (self *_AliasTypeDef_) Alias() {

}

func (self *_AliasTypeDef_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}

func (self *_AliasTypeDef_) Wrap(inner hir.Type) types.BuildInType {
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

func (self *aliasSintType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasSintType) String() string { return self.AliasTypeDef.String() }
func (self *aliasSintType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasUintType struct {
	AliasTypeDef
	types.UintType
}

func (self *aliasUintType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasUintType) String() string { return self.AliasTypeDef.String() }
func (self *aliasUintType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasFloatType struct {
	AliasTypeDef
	types.FloatType
}

func (self *aliasFloatType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasFloatType) String() string { return self.AliasTypeDef.String() }
func (self *aliasFloatType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasBoolType struct {
	AliasTypeDef
	types.BoolType
}

func (self *aliasBoolType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasBoolType) String() string { return self.AliasTypeDef.String() }
func (self *aliasBoolType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasStrType struct {
	AliasTypeDef
	types.StrType
}

func (self *aliasStrType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasStrType) String() string { return self.AliasTypeDef.String() }
func (self *aliasStrType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasRefType struct {
	AliasTypeDef
	types.RefType
}

func (self *aliasRefType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasRefType) String() string { return self.AliasTypeDef.String() }
func (self *aliasRefType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasArrayType struct {
	AliasTypeDef
	types.ArrayType
}

func (self *aliasArrayType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasArrayType) String() string { return self.AliasTypeDef.String() }
func (self *aliasArrayType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasTupleType struct {
	AliasTypeDef
	types.TupleType
}

func (self *aliasTupleType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasTupleType) String() string { return self.AliasTypeDef.String() }
func (self *aliasTupleType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasFuncType struct {
	AliasTypeDef
	types.FuncType
}

func (self *aliasFuncType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasFuncType) String() string { return self.AliasTypeDef.String() }
func (self *aliasFuncType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasLambdaType struct {
	AliasTypeDef
	types.LambdaType
}

func (self *aliasLambdaType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasLambdaType) String() string { return self.AliasTypeDef.String() }
func (self *aliasLambdaType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasStructType struct {
	AliasTypeDef
	types.StructType
}

func (self *aliasStructType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasStructType) String() string { return self.AliasTypeDef.String() }
func (self *aliasStructType) Hash() uint64   { return self.AliasTypeDef.Hash() }

type aliasEnumType struct {
	AliasTypeDef
	types.EnumType
}

func (self *aliasEnumType) Equal(dst hir.Type) bool {
	return self.AliasTypeDef.Equal(dst)
}
func (self *aliasEnumType) String() string { return self.AliasTypeDef.String() }
func (self *aliasEnumType) Hash() uint64   { return self.AliasTypeDef.Hash() }
