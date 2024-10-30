package global

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/hir/types"
)

// CustomTypeDef 自定义类型定义
type CustomTypeDef struct {
	pkgGlobalAttr
	name   string
	target types.Type
}

func NewCustomTypeDef(name string, target types.Type) *CustomTypeDef {
	return &CustomTypeDef{
		name:   name,
		target: target,
	}
}

func (self *CustomTypeDef) String() string {
	if self.pkg.IsBuildIn() {
		return self.name
	}
	return fmt.Sprintf("%s::%s", self.pkg.String(), self.name)
}

func (self *CustomTypeDef) Equal(dst types.Type) bool {
	t, ok := dst.(types.CustomType)
	return ok && self.pkg.Equal(t.(Global).Package()) && self.name == t.Name()
}

func (self *CustomTypeDef) Name() string {
	return self.name
}

func (self *CustomTypeDef) Target() types.Type {
	return self.target
}

func (self *CustomTypeDef) SetTarget(t types.Type) {
	self.target = t
}

type CustomTypeWrap interface {
	types.CustomType
	TypeDef
	Define() *CustomTypeDef
}

func WrapCustomType(t types.CustomType) CustomTypeWrap {
	return t.(CustomTypeWrap)
}

func (self *CustomTypeDef) Type() CustomTypeWrap {
	switch self.target.(type) {
	case types.SintType:
		return NewCustomSintType(self)
	case types.UintType:
		return NewCustomUintType(self)
	case types.FloatType:
		return NewCustomFloatType(self)
	case types.BoolType:
		return NewCustomBoolType(self)
	case types.StrType:
		return NewCustomStrType(self)
	case types.RefType:
		return NewCustomRefType(self)
	case types.ArrayType:
		return NewCustomArrayType(self)
	case types.TupleType:
		return NewCustomTupleType(self)
	case types.FuncType:
		return NewCustomFuncType(self)
	case types.StructType:
		return NewCustomStructType(self)
	case types.EnumType:
		return NewCustomEnumType(self)
	case types.LambdaType:
		return NewCustomLambdaType(self)
	default:
		panic("unreachable")
	}
}

type customSintType struct {
	*CustomTypeDef
}

func NewCustomSintType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customSintType{CustomTypeDef: typedef}
}
func (self *customSintType) Custom()                 {}
func (self *customSintType) Define() *CustomTypeDef  { return self.CustomTypeDef }
func (self *customSintType) Number()                 {}
func (self *customSintType) Kind() types.IntTypeKind { return self.target.(types.SintType).Kind() }
func (self *customSintType) Signed()                 {}

type customUintType struct {
	*CustomTypeDef
}

func NewCustomUintType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customUintType{CustomTypeDef: typedef}
}
func (self *customUintType) Custom()                 {}
func (self *customUintType) Define() *CustomTypeDef  { return self.CustomTypeDef }
func (self *customUintType) Number()                 {}
func (self *customUintType) Kind() types.IntTypeKind { return self.target.(types.UintType).Kind() }
func (self *customUintType) Unsigned()               {}

type customFloatType struct {
	*CustomTypeDef
}

func NewCustomFloatType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customFloatType{CustomTypeDef: typedef}
}
func (self *customFloatType) Custom()                   {}
func (self *customFloatType) Define() *CustomTypeDef    { return self.CustomTypeDef }
func (self *customFloatType) Number()                   {}
func (self *customFloatType) Kind() types.FloatTypeKind { return self.target.(types.FloatType).Kind() }
func (self *customFloatType) Signed()                   {}

type customBoolType struct {
	*CustomTypeDef
}

func NewCustomBoolType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customBoolType{CustomTypeDef: typedef}
}
func (self *customBoolType) Custom()                {}
func (self *customBoolType) Define() *CustomTypeDef { return self.CustomTypeDef }
func (self *customBoolType) Bool()                  {}

type customStrType struct {
	*CustomTypeDef
}

func NewCustomStrType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customStrType{CustomTypeDef: typedef}
}
func (self *customStrType) Custom()                {}
func (self *customStrType) Define() *CustomTypeDef { return self.CustomTypeDef }
func (self *customStrType) Str()                   {}

type customRefType struct {
	*CustomTypeDef
}

func NewCustomRefType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customRefType{CustomTypeDef: typedef}
}
func (self *customRefType) Custom()                {}
func (self *customRefType) Define() *CustomTypeDef { return self.CustomTypeDef }
func (self *customRefType) Mutable() bool          { return self.target.(types.RefType).Mutable() }
func (self *customRefType) Pointer() types.Type    { return self.target.(types.RefType).Pointer() }

type customArrayType struct {
	*CustomTypeDef
}

func NewCustomArrayType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customArrayType{CustomTypeDef: typedef}
}
func (self *customArrayType) Custom()                {}
func (self *customArrayType) Define() *CustomTypeDef { return self.CustomTypeDef }
func (self *customArrayType) Elem() types.Type       { return self.target.(types.ArrayType).Elem() }
func (self *customArrayType) Size() uint             { return self.target.(types.ArrayType).Size() }

type customTupleType struct {
	*CustomTypeDef
}

func NewCustomTupleType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customTupleType{CustomTypeDef: typedef}
}
func (self *customTupleType) Custom()                {}
func (self *customTupleType) Define() *CustomTypeDef { return self.CustomTypeDef }
func (self *customTupleType) Fields() []types.Type   { return self.target.(types.TupleType).Fields() }

type customFuncType struct {
	*CustomTypeDef
}

func NewCustomFuncType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customFuncType{CustomTypeDef: typedef}
}
func (self *customFuncType) Custom()                {}
func (self *customFuncType) Define() *CustomTypeDef { return self.CustomTypeDef }
func (self *customFuncType) Ret() types.Type        { return self.target.(types.FuncType).Ret() }
func (self *customFuncType) Params() []types.Type   { return self.target.(types.FuncType).Params() }
func (self *customFuncType) Func()                  {}

type customStructType struct {
	*CustomTypeDef
}

func NewCustomStructType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customStructType{CustomTypeDef: typedef}
}
func (self *customStructType) Custom()                {}
func (self *customStructType) Define() *CustomTypeDef { return self.CustomTypeDef }
func (self *customStructType) Fields() []*types.Field { return self.target.(types.StructType).Fields() }

type customEnumType struct {
	*CustomTypeDef
}

func NewCustomEnumType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customEnumType{CustomTypeDef: typedef}
}
func (self *customEnumType) Custom()                {}
func (self *customEnumType) Define() *CustomTypeDef { return self.CustomTypeDef }
func (self *customEnumType) EnumFields() []*types.EnumField {
	return self.target.(types.EnumType).EnumFields()
}

type customLambdaType struct {
	*CustomTypeDef
}

func NewCustomLambdaType(typedef *CustomTypeDef) CustomTypeWrap {
	return &customLambdaType{CustomTypeDef: typedef}
}
func (self *customLambdaType) Custom()                {}
func (self *customLambdaType) Define() *CustomTypeDef { return self.CustomTypeDef }
func (self *customLambdaType) Ret() types.Type        { return self.target.(types.LambdaType).Ret() }
func (self *customLambdaType) Params() []types.Type   { return self.target.(types.LambdaType).Params() }
func (self *customLambdaType) Lambda()                {}
