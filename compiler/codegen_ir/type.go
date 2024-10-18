package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/oldhir"
)

func (self *CodeGenerator) codegenType(t oldhir.Type) llvm.Type {
	switch t := oldhir.ToRuntimeType(t).(type) {
	case *oldhir.NoThingType, *oldhir.NoReturnType:
		return self.codegenEmptyType()
	case *oldhir.SintType:
		return self.codegenSintType(t)
	case *oldhir.UintType:
		return self.codegenUintType(t)
	case *oldhir.FloatType:
		return self.codegenFloatType(t)
	case *oldhir.ArrayType:
		return self.codegenArrayType(t)
	case *oldhir.TupleType:
		return self.codegenTupleType(t)
	case *oldhir.CustomType:
		return self.codegenCustomType(t)
	case *oldhir.FuncType, *oldhir.RefType:
		return self.codegenRefType()
	case *oldhir.StructType:
		return self.codegenStructType(t)
	case *oldhir.LambdaType:
		return self.codegenLambdaType(t)
	case *oldhir.EnumType:
		return self.codegenEnumType(t)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenEmptyType() llvm.VoidType {
	return self.builder.VoidType()
}

func (self *CodeGenerator) codegenSintType(ir *oldhir.SintType) llvm.IntegerType {
	return self.builder.IntegerType(uint32(ir.Bits))
}

func (self *CodeGenerator) codegenUintType(ir *oldhir.UintType) llvm.IntegerType {
	return self.builder.IntegerType(uint32(ir.Bits))
}

func (self *CodeGenerator) codegenFloatType(ir *oldhir.FloatType) llvm.FloatType {
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

func (self *CodeGenerator) codegenFuncType(ir *oldhir.FuncType) llvm.FunctionType {
	ft, _, _ := self.codegenCallableType(ir)
	return ft
}

func (self *CodeGenerator) codegenCallableType(ir oldhir.CallableType) (llvm.FunctionType, llvm.StructType, llvm.FunctionType) {
	if oldhir.IsType[*oldhir.FuncType](ir) {
		ft := oldhir.AsType[*oldhir.FuncType](ir)
		ret := self.codegenType(ft.Ret)
		params := stlslices.Map(ft.Params, func(_ int, e oldhir.Type) llvm.Type {
			return self.codegenType(e)
		})
		return self.builder.FunctionType(false, ret, params...), llvm.StructType{}, llvm.FunctionType{}
	} else {
		lbdt := oldhir.AsType[*oldhir.LambdaType](ir)
		ret := self.codegenType(lbdt.Ret)
		params := stlslices.Map(lbdt.Params, func(_ int, e oldhir.Type) llvm.Type {
			return self.codegenType(e)
		})
		ft1 := self.builder.FunctionType(false, ret, params...)
		ft2 := self.builder.FunctionType(false, ret, append([]llvm.Type{self.builder.OpaquePointerType()}, params...)...)
		return ft1, self.builder.StructType(false, self.builder.OpaquePointerType(), self.builder.OpaquePointerType(), self.builder.OpaquePointerType()), ft2
	}
}

func (self *CodeGenerator) codegenArrayType(ir *oldhir.ArrayType) llvm.ArrayType {
	elem := self.codegenType(ir.Elem)
	return self.builder.ArrayType(elem, uint32(ir.Size))
}

func (self *CodeGenerator) codegenTupleType(ir *oldhir.TupleType) llvm.StructType {
	elems := stlslices.Map(ir.Elems, func(_ int, e oldhir.Type) llvm.Type {
		return self.codegenType(e)
	})
	return self.builder.StructType(false, elems...)
}

func (self *CodeGenerator) codegenCustomType(ir *oldhir.CustomType) llvm.Type {
	return self.codegenType(ir.Target)
}

func (self *CodeGenerator) codegenStructType(ir *oldhir.StructType) llvm.StructType {
	if self.types.Contain(ir.Def) {
		return self.types.Get(ir.Def)
	}
	st := self.builder.NamedStructType("", false)
	self.types.Set(ir.Def, st)
	st.SetElems(false, stlslices.Map(ir.Fields.Values(), func(_ int, e oldhir.Field) llvm.Type {
		return self.codegenType(e.Type)
	})...)
	return st
}

func (self *CodeGenerator) codegenRefType() llvm.PointerType {
	return self.builder.OpaquePointerType()
}

func (self *CodeGenerator) codegenLambdaType(ir *oldhir.LambdaType) llvm.StructType {
	_, st, _ := self.codegenCallableType(ir)
	return st
}

func (self *CodeGenerator) codegenEnumType(ir *oldhir.EnumType) llvm.Type {
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
