package global

import (
	"fmt"
	"strings"
	"unsafe"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// CompileCallTypeDef 编译时调用类型定义
type CompileCallTypeDef interface {
	CustomTypeDef
	types.CompileCallType
}

type _CompileCallTypeDef_ struct {
	CustomTypeDef
	args []hir.Type
}

func NewCompileCallTypeDef(fn CustomTypeDef, args ...hir.Type) CompileCallTypeDef {
	return &_CompileCallTypeDef_{
		CustomTypeDef: fn,
		args:          args,
	}
}

func (self *_CompileCallTypeDef_) CompileCall() {}

func (self *_CompileCallTypeDef_) String() string {
	args := stlslices.Map(self.args, func(_ int, arg hir.Type) string {
		return arg.String()
	})
	return fmt.Sprintf("%s::<%s>", self.CustomTypeDef.String(), strings.Join(args, ","))
}

func (self *_CompileCallTypeDef_) Equal(dst hir.Type) bool {
	return self == dst
}

func (self *_CompileCallTypeDef_) GetName() (string, bool) {
	return self.String(), true
}

func (self *_CompileCallTypeDef_) Hash() uint64 {
	return uint64(uintptr(unsafe.Pointer(self)))
}

func (self *_CompileCallTypeDef_) getCompileParamMap() map[types.CompileParamType]hir.Type {
	var i int
	return stlslices.ToMap(self.CustomTypeDef.CompilerParams(), func(compileParam types.CompileParamType) (types.CompileParamType, hir.Type) {
		i++
		return compileParam, self.args[i-1]
	})
}

func (self *_CompileCallTypeDef_) Target() hir.Type {
	return types.ReplaceCompileParam(self.getCompileParamMap(), self.CustomTypeDef.Target())
}

func (self *_CompileCallTypeDef_) SetTarget(_ hir.Type) {
	panic("unreachable")
}

func (self *_CompileCallTypeDef_) AddMethod(m *MethodDef) bool {
	panic("unreachable")
}

func (self *_CompileCallTypeDef_) GetMethod(name string) (*MethodDef, bool) {
	return nil, false
}

func (self *_CompileCallTypeDef_) Args() []hir.Type {
	return self.args
}

func (self *_CompileCallTypeDef_) WithArgs(args []hir.Type) types.CompileCallType {
	return NewCompileCallTypeDef(self.CustomTypeDef, args...)
}

func (self *_CompileCallTypeDef_) Wrap(inner hir.Type) types.BuildInType {
	switch v := inner.(type) {
	case types.SintType:
		return &compileCallSintType{CompileCallTypeDef: self, SintType: v}
	case types.UintType:
		return &compileCallUintType{CompileCallTypeDef: self, UintType: v}
	case types.FloatType:
		return &compileCallFloatType{CompileCallTypeDef: self, FloatType: v}
	case types.BoolType:
		return &compileCallBoolType{CompileCallTypeDef: self, BoolType: v}
	case types.StrType:
		return &compileCallStrType{CompileCallTypeDef: self, StrType: v}
	case types.RefType:
		return &compileCallRefType{CompileCallTypeDef: self, RefType: v}
	case types.ArrayType:
		return &compileCallArrayType{CompileCallTypeDef: self, ArrayType: v}
	case types.TupleType:
		return &compileCallTupleType{CompileCallTypeDef: self, TupleType: v}
	case types.FuncType:
		return &compileCallFuncType{CompileCallTypeDef: self, FuncType: v}
	case types.LambdaType:
		return &compileCallLambdaType{CompileCallTypeDef: self, LambdaType: v}
	case types.StructType:
		return &compileCallStructType{CompileCallTypeDef: self, StructType: v}
	case types.EnumType:
		return &compileCallEnumType{CompileCallTypeDef: self, EnumType: v}
	default:
		panic("unreachable")
	}
}

type compileCallSintType struct {
	CompileCallTypeDef
	types.SintType
}

func (self *compileCallSintType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallSintType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallSintType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallUintType struct {
	CompileCallTypeDef
	types.UintType
}

func (self *compileCallUintType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallUintType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallUintType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallFloatType struct {
	CompileCallTypeDef
	types.FloatType
}

func (self *compileCallFloatType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallFloatType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallFloatType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallBoolType struct {
	CompileCallTypeDef
	types.BoolType
}

func (self *compileCallBoolType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallBoolType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallBoolType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallStrType struct {
	CompileCallTypeDef
	types.StrType
}

func (self *compileCallStrType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallStrType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallStrType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallRefType struct {
	CompileCallTypeDef
	types.RefType
}

func (self *compileCallRefType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallRefType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallRefType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallArrayType struct {
	CompileCallTypeDef
	types.ArrayType
}

func (self *compileCallArrayType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallArrayType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallArrayType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallTupleType struct {
	CompileCallTypeDef
	types.TupleType
}

func (self *compileCallTupleType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallTupleType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallTupleType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallFuncType struct {
	CompileCallTypeDef
	types.FuncType
}

func (self *compileCallFuncType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallFuncType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallFuncType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallLambdaType struct {
	CompileCallTypeDef
	types.LambdaType
}

func (self *compileCallLambdaType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallLambdaType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallLambdaType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallStructType struct {
	CompileCallTypeDef
	types.StructType
}

func (self *compileCallStructType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallStructType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallStructType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }

type compileCallEnumType struct {
	CompileCallTypeDef
	types.EnumType
}

func (self *compileCallEnumType) Equal(dst hir.Type) bool {
	return self.CompileCallTypeDef.Equal(dst)
}
func (self *compileCallEnumType) String() string { return self.CompileCallTypeDef.String() }
func (self *compileCallEnumType) Hash() uint64   { return self.CompileCallTypeDef.Hash() }
