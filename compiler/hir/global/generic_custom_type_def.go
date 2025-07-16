package global

import (
	"fmt"
	"strings"

	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/utils"
)

// GenericCustomTypeDef 泛型类型定义
type GenericCustomTypeDef interface {
	CustomTypeDef
	types.GenericCustomType
}

type _GenericCustomTypeDef_ struct {
	CustomTypeDef
	args []hir.Type
}

func NewGenericCustomTypeDef(origin CustomTypeDef, args ...hir.Type) GenericCustomTypeDef {
	return &_GenericCustomTypeDef_{
		CustomTypeDef: origin,
		args:          args,
	}
}

func (self *_GenericCustomTypeDef_) String() string {
	args := stlslices.Map(self.args, func(_ int, arg hir.Type) string {
		return arg.String()
	})
	return fmt.Sprintf("%s::<%s>", self.CustomTypeDef.String(), strings.Join(args, ","))
}

func (self *_GenericCustomTypeDef_) Equal(dst hir.Type) bool {
	gctd, ok := types.As[GenericCustomTypeDef](dst, true)
	if !ok {
		return false
	} else if !self.CustomTypeDef.Equal(gctd.(CustomTypeDef)) {
		return false
	}
	return stlslices.All(self.args, func(i int, e hir.Type) bool {
		return e.Equal(gctd.GenericArgs()[i])
	})
}

func (self *_GenericCustomTypeDef_) GetName() (utils.Name, bool) {
	name, ok := self.CustomTypeDef.GetName()
	if !ok {
		return utils.Name{}, false
	}
	return utils.Name{
		Value:    self.String(),
		Position: name.Position,
	}, true
}

func (self *_GenericCustomTypeDef_) TotalName(genericParamMap hashmap.HashMap[types.VirtualType, hir.Type]) string {
	args := stlslices.Map(self.args, func(_ int, arg hir.Type) string {
		return types.ReplaceVirtualType(genericParamMap, arg).String()
	})
	return fmt.Sprintf("%s::<%s>", self.CustomTypeDef.String(), strings.Join(args, ","))
}

func (self *_GenericCustomTypeDef_) Hash() uint64 {
	return self.CustomTypeDef.Hash()
}

func (self *_GenericCustomTypeDef_) GenericParamMap() hashmap.HashMap[types.VirtualType, hir.Type] {
	var i int
	return hashmap.AnyWith[types.VirtualType, hir.Type](stlslices.FlatMap(self.CustomTypeDef.GenericParams(), func(_ int, compileParam types.GenericParamType) []any {
		i++
		return []any{types.VirtualType(compileParam), self.args[i-1]}
	})...)
}

func (self *_GenericCustomTypeDef_) Target() hir.Type {
	return types.ReplaceVirtualType(self.GenericParamMap(), self.CustomTypeDef.Target())
}

func (self *_GenericCustomTypeDef_) SetTarget(_ hir.Type) {
	panic("unreachable")
}

func (self *_GenericCustomTypeDef_) AddMethod(m MethodDef) bool {
	panic("unreachable")
}

func (self *_GenericCustomTypeDef_) GetMethod(name string) (MethodDef, bool) {
	method, ok := self.CustomTypeDef.GetMethod(name)
	if !ok {
		return nil, false
	}
	return newGenericCustomTypeMethodDef(method.(*OriginMethodDef), self), true
}

func (self *_GenericCustomTypeDef_) GenericArgs() []hir.Type {
	return self.args
}

func (self *_GenericCustomTypeDef_) WithGenericArgs(args []hir.Type) types.GenericCustomType {
	return NewGenericCustomTypeDef(self.CustomTypeDef, args...)
}

func (self *_GenericCustomTypeDef_) Wrap(inner hir.Type) hir.BuildInType {
	switch v := inner.(type) {
	case types.SintType:
		return &genericCustomSintType{GenericCustomTypeDef: self, SintType: v}
	case types.UintType:
		return &genericCustomUintType{GenericCustomTypeDef: self, UintType: v}
	case types.FloatType:
		return &genericCustomFloatType{GenericCustomTypeDef: self, FloatType: v}
	case types.BoolType:
		return &genericCustomBoolType{GenericCustomTypeDef: self, BoolType: v}
	case types.StrType:
		return &genericCustomStrType{GenericCustomTypeDef: self, StrType: v}
	case types.RefType:
		return &genericCustomRefType{GenericCustomTypeDef: self, RefType: v}
	case types.ArrayType:
		return &genericCustomArrayType{GenericCustomTypeDef: self, ArrayType: v}
	case types.TupleType:
		return &genericCustomTupleType{GenericCustomTypeDef: self, TupleType: v}
	case types.FuncType:
		return &genericCustomFuncType{GenericCustomTypeDef: self, FuncType: v}
	case types.LambdaType:
		return &genericCustomLambdaType{GenericCustomTypeDef: self, LambdaType: v}
	case types.StructType:
		return &genericCustomStructType{GenericCustomTypeDef: self, StructType: v}
	case types.EnumType:
		return &genericCustomEnumType{GenericCustomTypeDef: self, EnumType: v}
	default:
		panic("unreachable")
	}
}

type genericCustomSintType struct {
	GenericCustomTypeDef
	types.SintType
}

func (self *genericCustomSintType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomSintType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomSintType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomUintType struct {
	GenericCustomTypeDef
	types.UintType
}

func (self *genericCustomUintType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomUintType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomUintType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomFloatType struct {
	GenericCustomTypeDef
	types.FloatType
}

func (self *genericCustomFloatType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomFloatType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomFloatType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomBoolType struct {
	GenericCustomTypeDef
	types.BoolType
}

func (self *genericCustomBoolType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomBoolType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomBoolType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomStrType struct {
	GenericCustomTypeDef
	types.StrType
}

func (self *genericCustomStrType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomStrType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomStrType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomRefType struct {
	GenericCustomTypeDef
	types.RefType
}

func (self *genericCustomRefType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomRefType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomRefType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomArrayType struct {
	GenericCustomTypeDef
	types.ArrayType
}

func (self *genericCustomArrayType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomArrayType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomArrayType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomTupleType struct {
	GenericCustomTypeDef
	types.TupleType
}

func (self *genericCustomTupleType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomTupleType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomTupleType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomFuncType struct {
	GenericCustomTypeDef
	types.FuncType
}

func (self *genericCustomFuncType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomFuncType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomFuncType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomLambdaType struct {
	GenericCustomTypeDef
	types.LambdaType
}

func (self *genericCustomLambdaType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomLambdaType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomLambdaType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomStructType struct {
	GenericCustomTypeDef
	types.StructType
}

func (self *genericCustomStructType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomStructType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomStructType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }

type genericCustomEnumType struct {
	GenericCustomTypeDef
	types.EnumType
}

func (self *genericCustomEnumType) Equal(dst hir.Type) bool {
	return self.GenericCustomTypeDef.Equal(dst)
}
func (self *genericCustomEnumType) String() string { return self.GenericCustomTypeDef.String() }
func (self *genericCustomEnumType) Hash() uint64   { return self.GenericCustomTypeDef.Hash() }
