package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (self *CodeGenerator) codegenType(t types.Type) llvm.Type {
	switch t := t.(type) {
	case types.NoThingType, types.NoReturnType:
		return self.builder.VoidType()
	case types.CustomType:
		tObj := self.types.Get(t)
		if tObj != nil {
			return self.types.Get(t.(global.TypeDef).Define().(types.CustomType))
		}
		return self.codegenType(t.Target())
	case types.AliasType:
		return self.codegenType(t.Target())
	case types.IntType:
		return self.codegenIntType(t)
	case types.FloatType:
		return self.codegenFloatType(t)
	case types.BoolType:
		return self.builder.BooleanType()
	case types.StrType:
		return self.builder.Str()
	case types.RefType, types.FuncType:
		return self.builder.OpaquePointerType()
	case types.ArrayType:
		return self.codegenArrayType(t)
	case types.TupleType:
		return self.codegenTupleType(t)
	case types.LambdaType:
		return self.codegenLambdaType()
	case types.StructType:
		return self.codegenStructType(t)
	case types.EnumType:
		return self.codegenEnumType(t)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenIntType(ir types.IntType) llvm.IntegerType {
	switch ir.Kind() {
	case types.IntTypeKindSize:
		return self.builder.Isize()
	case types.IntTypeKindByte:
		return self.builder.I8()
	case types.IntTypeKindShort:
		return self.builder.I16()
	case types.IntTypeKindInt:
		return self.builder.I32()
	case types.IntTypeKindLong:
		return self.builder.I64()
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFloatType(ir types.FloatType) llvm.FloatType {
	switch ir.Kind() {
	case types.FloatTypeKindHalf:
		return self.builder.F16()
	case types.FloatTypeKindFloat:
		return self.builder.F32()
	case types.FloatTypeKindDouble:
		return self.builder.F64()
	case types.FloatTypeKindFP128:
		return self.builder.F128()
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFuncType(ir types.FuncType) llvm.FunctionType {
	ret := self.codegenType(ir.Ret())
	params := stlslices.Map(ir.Params(), func(_ int, p types.Type) llvm.Type {
		return self.codegenType(p)
	})
	return self.builder.FunctionType(false, ret, params...)
}

func (self *CodeGenerator) codegenArrayType(ir types.ArrayType) llvm.ArrayType {
	elem := self.codegenType(ir.Elem())
	return self.builder.ArrayType(elem, uint32(ir.Size()))
}

func (self *CodeGenerator) codegenTupleType(ir types.TupleType) llvm.StructType {
	elems := stlslices.Map(ir.Elems(), func(_ int, e types.Type) llvm.Type {
		return self.codegenType(e)
	})
	return self.builder.StructType(false, elems...)
}

func (self *CodeGenerator) codegenStructType(ir types.StructType) llvm.StructType {
	elems := stlslices.Map(ir.Fields().Values(), func(_ int, f *types.Field) llvm.Type {
		return self.codegenType(f.Type())
	})
	return self.builder.StructType(false, elems...)
}

func (self *CodeGenerator) codegenLambdaType() llvm.StructType {
	return self.builder.StructType(false, self.builder.OpaquePointerType(), self.builder.OpaquePointerType(), self.builder.OpaquePointerType())
}

func (self *CodeGenerator) codegenEnumType(ir types.EnumType) llvm.Type {
	if ir.Simple() {
		return self.builder.I8()
	}

	var maxSizeType llvm.Type
	var maxSize uint
	for iter := ir.EnumFields().Iterator(); iter.Next(); {
		elemIr, ok := iter.Value().E2().Elem()
		if !ok {
			continue
		}
		elem := self.codegenType(elemIr)
		if esize := self.builder.GetStoreSizeOfType(elem); esize > maxSize {
			maxSizeType, maxSize = elem, esize
		}
	}
	return self.builder.StructType(false, maxSizeType, self.builder.I8())
}
