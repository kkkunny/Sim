package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

func (self *CodeGenerator) codegenType(t hir.Type) llvm.Type {
	switch t := hir.ToRuntimeType(t).(type) {
	case *hir.NoThingType, *hir.NoReturnType:
		return self.codegenEmptyType()
	case *hir.SintType:
		return self.codegenSintType(t)
	case *hir.UintType:
		return self.codegenUintType(t)
	case *hir.FloatType:
		return self.codegenFloatType(t)
	case *hir.ArrayType:
		return self.codegenArrayType(t)
	case *hir.TupleType:
		return self.codegenTupleType(t)
	case *hir.CustomType:
		return self.codegenCustomType(t)
	case *hir.FuncType, *hir.RefType:
		return self.codegenRefType()
	case *hir.StructType:
		return self.codegenStructType(t)
	case *hir.LambdaType:
		return self.codegenLambdaType(t)
	case *hir.EnumType:
		return self.codegenEnumType(t)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenEmptyType() llvm.VoidType {
	return self.builder.VoidType()
}

func (self *CodeGenerator) codegenSintType(ir *hir.SintType) llvm.IntegerType {
	return self.builder.IntegerType(uint32(ir.Bits))
}

func (self *CodeGenerator) codegenUintType(ir *hir.UintType) llvm.IntegerType {
	return self.builder.IntegerType(uint32(ir.Bits))
}

func (self *CodeGenerator) codegenFloatType(ir *hir.FloatType) llvm.FloatType {
	var ft llvm.FloatTypeKind
	switch ir.Bits {
	case 16:
		ft = llvm.FloatTypeKindHalf
	case 32:
		ft = llvm.FloatTypeKindFloat
	case 64:
		ft = llvm.FloatTypeKindDouble
	case 128:
		ft = llvm.FloatTypeKindFP128
	default:
		panic("unreachable")
	}
	return self.builder.FloatType(ft)
}

func (self *CodeGenerator) codegenFuncType(ir *hir.FuncType) llvm.FunctionType {
	ft, _, _ := self.codegenCallableType(ir)
	return ft
}

func (self *CodeGenerator) codegenCallableType(ir hir.CallableType) (llvm.FunctionType, llvm.StructType, llvm.FunctionType) {
	if hir.IsType[*hir.FuncType](ir) {
		ft := hir.AsType[*hir.FuncType](ir)
		ret := self.codegenType(ft.Ret)
		params := stlslices.Map(ft.Params, func(_ int, e hir.Type) llvm.Type {
			return self.codegenType(e)
		})
		return self.builder.FunctionType(false, ret, params...), llvm.StructType{}, llvm.FunctionType{}
	} else {
		lbdt := hir.AsType[*hir.LambdaType](ir)
		ret := self.codegenType(lbdt.Ret)
		params := stlslices.Map(lbdt.Params, func(_ int, e hir.Type) llvm.Type {
			return self.codegenType(e)
		})
		ft1 := self.builder.FunctionType(false, ret, params...)
		ft2 := self.builder.FunctionType(false, ret, append([]llvm.Type{self.builder.OpaquePointerType()}, params...)...)
		return ft1, self.builder.StructType(false, self.builder.OpaquePointerType(), self.builder.OpaquePointerType(), self.builder.OpaquePointerType()), ft2
	}
}

func (self *CodeGenerator) codegenArrayType(ir *hir.ArrayType) llvm.ArrayType {
	elem := self.codegenType(ir.Elem)
	return self.builder.ArrayType(elem, uint32(ir.Size))
}

func (self *CodeGenerator) codegenTupleType(ir *hir.TupleType) llvm.StructType {
	elems := stlslices.Map(ir.Elems, func(_ int, e hir.Type) llvm.Type {
		return self.codegenType(e)
	})
	return self.builder.StructType(false, elems...)
}

func (self *CodeGenerator) codegenCustomType(ir *hir.CustomType) llvm.Type {
	return self.codegenType(ir.Target)
}

func (self *CodeGenerator) codegenStructType(ir *hir.StructType) llvm.StructType {
	if self.types.Contain(ir.Def) {
		return self.types.Get(ir.Def)
	}
	st := self.builder.NamedStructType("", false)
	self.types.Set(ir.Def, st)
	st.SetElems(false, stlslices.Map(ir.Fields.Values(), func(_ int, e hir.Field) llvm.Type {
		return self.codegenType(e.Type)
	})...)
	return st
}

func (self *CodeGenerator) codegenRefType() llvm.PointerType {
	return self.builder.OpaquePointerType()
}

func (self *CodeGenerator) codegenLambdaType(ir *hir.LambdaType) llvm.StructType {
	_, st, _ := self.codegenCallableType(ir)
	return st
}

func (self *CodeGenerator) codegenEnumType(ir *hir.EnumType) llvm.Type {
	if ir.IsSimple() {
		return self.builder.IntegerType(8)
	}

	if self.types.Contain(ir.Def) {
		return self.types.Get(ir.Def)
	}
	st := self.builder.NamedStructType("", false)
	self.types.Set(ir.Def, st)
	var maxSizeType llvm.Type
	var maxSize uint
	for iter := ir.Fields.Iterator(); iter.Next(); {
		if iter.Value().E2().Elem.IsNone() {
			continue
		}
		et := self.codegenType(iter.Value().E2().Elem.MustValue())
		if esize := self.builder.GetStoreSizeOfType(et); esize > maxSize {
			maxSizeType, maxSize = et, esize
		}
	}
	st.SetElems(false, maxSizeType, self.builder.IntegerType(8))
	return st
}
