package global

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/hir/types"
)

// AliasTypeDef 类型别名定义
type AliasTypeDef struct {
	pkgGlobalAttr
	name   string
	target types.Type
}

func NewAliasTypeDef(name string, target types.Type) *AliasTypeDef {
	return &AliasTypeDef{
		name:   name,
		target: target,
	}
}

func (self *AliasTypeDef) String() string {
	if self.pkg.IsBuildIn() {
		return self.name
	}
	return fmt.Sprintf("%s::%s", self.pkg.String(), self.name)
}

func (self *AliasTypeDef) Equal(dst types.Type) bool {
	t, ok := dst.(*AliasTypeDef)
	if ok {
		return self.pkg.Equal(t.pkg) && self.name == t.name
	} else {
		return self.target.Equal(dst)
	}
}

func (self *AliasTypeDef) Name() string {
	return self.name
}

func (self *AliasTypeDef) Target() types.Type {
	return self.target
}

func (self *AliasTypeDef) SetTarget(t types.Type) {
	self.target = t
}

type AliasTypeWrap interface {
	types.AliasType
	TypeDef
	Define() *AliasTypeDef
}

func WrapAliasType(t types.AliasType) AliasTypeWrap {
	return t.(AliasTypeWrap)
}

func (self *AliasTypeDef) Type() AliasTypeWrap {
	switch self.target.(type) {
	case types.SintType:
		return NewAliasSintType(self)
	case types.UintType:
		return NewAliasUintType(self)
	case types.FloatType:
		return NewAliasFloatType(self)
	case types.BoolType:
		return NewAliasBoolType(self)
	case types.StrType:
		return NewAliasStrType(self)
	case types.RefType:
		return NewAliasRefType(self)
	case types.ArrayType:
		return NewAliasArrayType(self)
	case types.TupleType:
		return NewAliasTupleType(self)
	case types.FuncType:
		return NewAliasFuncType(self)
	case types.StructType:
		return NewAliasStructType(self)
	case types.EnumType:
		return NewAliasEnumType(self)
	case types.LambdaType:
		return NewAliasLambdaType(self)
	default:
		panic("unreachable")
	}
}

type aliasSintType struct {
	*AliasTypeDef
}

func NewAliasSintType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasSintType{AliasTypeDef: typedef}
}
func (self *aliasSintType) Alias()                  {}
func (self *aliasSintType) Define() *AliasTypeDef   { return self.AliasTypeDef }
func (self *aliasSintType) Number()                 {}
func (self *aliasSintType) Kind() types.IntTypeKind { return self.target.(types.SintType).Kind() }
func (self *aliasSintType) Signed()                 {}

type aliasUintType struct {
	*AliasTypeDef
}

func NewAliasUintType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasUintType{AliasTypeDef: typedef}
}
func (self *aliasUintType) Alias()                  {}
func (self *aliasUintType) Define() *AliasTypeDef   { return self.AliasTypeDef }
func (self *aliasUintType) Number()                 {}
func (self *aliasUintType) Kind() types.IntTypeKind { return self.target.(types.UintType).Kind() }
func (self *aliasUintType) Unsigned()               {}

type aliasFloatType struct {
	*AliasTypeDef
}

func NewAliasFloatType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasFloatType{AliasTypeDef: typedef}
}
func (self *aliasFloatType) Alias()                    {}
func (self *aliasFloatType) Define() *AliasTypeDef     { return self.AliasTypeDef }
func (self *aliasFloatType) Number()                   {}
func (self *aliasFloatType) Kind() types.FloatTypeKind { return self.target.(types.FloatType).Kind() }
func (self *aliasFloatType) Signed()                   {}

type aliasBoolType struct {
	*AliasTypeDef
}

func NewAliasBoolType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasBoolType{AliasTypeDef: typedef}
}
func (self *aliasBoolType) Alias()                {}
func (self *aliasBoolType) Define() *AliasTypeDef { return self.AliasTypeDef }
func (self *aliasBoolType) Bool()                 {}

type aliasStrType struct {
	*AliasTypeDef
}

func NewAliasStrType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasStrType{AliasTypeDef: typedef}
}
func (self *aliasStrType) Alias()                {}
func (self *aliasStrType) Define() *AliasTypeDef { return self.AliasTypeDef }
func (self *aliasStrType) Str()                  {}

type aliasRefType struct {
	*AliasTypeDef
}

func NewAliasRefType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasRefType{AliasTypeDef: typedef}
}
func (self *aliasRefType) Alias()                {}
func (self *aliasRefType) Define() *AliasTypeDef { return self.AliasTypeDef }
func (self *aliasRefType) Mutable() bool         { return self.target.(types.RefType).Mutable() }
func (self *aliasRefType) Pointer() types.Type   { return self.target.(types.RefType).Pointer() }

type aliasArrayType struct {
	*AliasTypeDef
}

func NewAliasArrayType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasArrayType{AliasTypeDef: typedef}
}
func (self *aliasArrayType) Alias()                {}
func (self *aliasArrayType) Define() *AliasTypeDef { return self.AliasTypeDef }
func (self *aliasArrayType) Elem() types.Type      { return self.target.(types.ArrayType).Elem() }
func (self *aliasArrayType) Size() uint            { return self.target.(types.ArrayType).Size() }

type aliasTupleType struct {
	*AliasTypeDef
}

func NewAliasTupleType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasTupleType{AliasTypeDef: typedef}
}
func (self *aliasTupleType) Alias()                {}
func (self *aliasTupleType) Define() *AliasTypeDef { return self.AliasTypeDef }
func (self *aliasTupleType) Fields() []types.Type  { return self.target.(types.TupleType).Fields() }

type aliasFuncType struct {
	*AliasTypeDef
}

func NewAliasFuncType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasFuncType{AliasTypeDef: typedef}
}
func (self *aliasFuncType) Alias()                {}
func (self *aliasFuncType) Define() *AliasTypeDef { return self.AliasTypeDef }
func (self *aliasFuncType) Ret() types.Type       { return self.target.(types.FuncType).Ret() }
func (self *aliasFuncType) Params() []types.Type  { return self.target.(types.FuncType).Params() }
func (self *aliasFuncType) Func()                 {}

type aliasStructType struct {
	*AliasTypeDef
}

func NewAliasStructType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasStructType{AliasTypeDef: typedef}
}
func (self *aliasStructType) Alias()                 {}
func (self *aliasStructType) Define() *AliasTypeDef  { return self.AliasTypeDef }
func (self *aliasStructType) Fields() []*types.Field { return self.target.(types.StructType).Fields() }

type aliasEnumType struct {
	*AliasTypeDef
}

func NewAliasEnumType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasEnumType{AliasTypeDef: typedef}
}
func (self *aliasEnumType) Alias()                {}
func (self *aliasEnumType) Define() *AliasTypeDef { return self.AliasTypeDef }
func (self *aliasEnumType) EnumFields() []*types.EnumField {
	return self.target.(types.EnumType).EnumFields()
}

type aliasLambdaType struct {
	*AliasTypeDef
}

func NewAliasLambdaType(typedef *AliasTypeDef) AliasTypeWrap {
	return &aliasLambdaType{AliasTypeDef: typedef}
}
func (self *aliasLambdaType) Alias()                {}
func (self *aliasLambdaType) Define() *AliasTypeDef { return self.AliasTypeDef }
func (self *aliasLambdaType) Ret() types.Type       { return self.target.(types.LambdaType).Ret() }
func (self *aliasLambdaType) Params() []types.Type  { return self.target.(types.LambdaType).Params() }
func (self *aliasLambdaType) Lambda()               {}
