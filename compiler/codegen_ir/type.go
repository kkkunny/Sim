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
	case *hir.FuncType:
		return self.codegenFuncTypePtr(t)
	case *hir.ArrayType:
		return self.codegenArrayType(t)
	case *hir.TupleType:
		return self.codegenTupleType(t)
	case *hir.CustomType:
		return self.codegenCustomType(t)
	case *hir.RefType:
		return self.codegenRefType(t)
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
	return self.ctx.VoidType()
}

func (self *CodeGenerator) codegenSintType(ir *hir.SintType) llvm.IntegerType {
	return self.ctx.IntegerType(uint32(ir.Bits))
}

func (self *CodeGenerator) codegenUintType(ir *hir.UintType) llvm.IntegerType {
	return self.ctx.IntegerType(uint32(ir.Bits))
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
	return self.ctx.FloatType(ft)
}

func (self *CodeGenerator) codegenFuncTypePtr(ir *hir.FuncType) llvm.PointerType {
	return self.ctx.PointerType(self.codegenFuncType(ir))
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
		return self.ctx.FunctionType(false, ret, params...), llvm.StructType{}, llvm.FunctionType{}
	} else {
		lbdt := hir.AsType[*hir.LambdaType](ir)
		ret := self.codegenType(lbdt.Ret)
		params := stlslices.Map(lbdt.Params, func(_ int, e hir.Type) llvm.Type {
			return self.codegenType(e)
		})
		ft1 := self.ctx.FunctionType(false, ret, params...)
		ft2 := self.ctx.FunctionType(false, ret, append([]llvm.Type{self.ptrType()}, params...)...)
		return ft1, self.ctx.StructType(false, self.ctx.PointerType(ft1), self.ctx.PointerType(ft2), self.ptrType()), ft2
	}
}

func (self *CodeGenerator) codegenArrayType(ir *hir.ArrayType) llvm.ArrayType {
	elem := self.codegenType(ir.Elem)
	return self.ctx.ArrayType(elem, uint32(ir.Size))
}

func (self *CodeGenerator) codegenTupleType(ir *hir.TupleType) llvm.StructType {
	elems := stlslices.Map(ir.Elems, func(_ int, e hir.Type) llvm.Type {
		return self.codegenType(e)
	})
	return self.ctx.StructType(false, elems...)
}

func (self *CodeGenerator) codegenCustomType(ir *hir.CustomType) llvm.Type {
	// TODO: 支持除结构体之外的类型循环
	return self.codegenType(ir.Target)
}

func (self *CodeGenerator) codegenStructType(ir *hir.StructType) llvm.StructType {
	if self.types.ContainKey(ir.Def) {
		return self.types.Get(ir.Def)
	}
	st := self.ctx.NamedStructType("", false)
	self.types.Set(ir.Def, st)
	st.SetElems(false, stlslices.Map(ir.Fields.Values().ToSlice(), func(_ int, e hir.Field) llvm.Type {
		return self.codegenType(e.Type)
	})...)
	return st
}

func (self *CodeGenerator) codegenRefType(ir *hir.RefType) llvm.PointerType {
	elem := self.codegenType(ir.Elem)
	return self.ctx.PointerType(elem)
}

func (self *CodeGenerator) codegenLambdaType(ir *hir.LambdaType) llvm.StructType {
	_, st, _ := self.codegenCallableType(ir)
	return st
}

func (self *CodeGenerator) codegenEnumType(ir *hir.EnumType) llvm.Type {
	if ir.IsSimple() {
		return self.ctx.IntegerType(8)
	}

	if self.types.ContainKey(ir.Def) {
		return self.types.Get(ir.Def)
	}
	st := self.ctx.NamedStructType("", false)
	self.types.Set(ir.Def, st)
	var maxSizeType llvm.Type
	var maxSize uint
	for iter := ir.Fields.Iterator(); iter.Next(); {
		if len(iter.Value().Second.Elems) == 0 {
			continue
		}
		et := self.codegenTupleType(hir.NewTupleType(iter.Value().Second.Elems...))
		if esize := self.target.GetStoreSizeOfType(et); esize > maxSize {
			maxSizeType, maxSize = et, esize
		}
	}
	st.SetElems(false, maxSizeType, self.ctx.IntegerType(8))
	return st
}
