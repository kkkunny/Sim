package codegen_ir

import (
	stlos "github.com/kkkunny/stl/os"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) codegenType(t hir.Type) mir.Type {
	switch t := hir.ToRuntimeType(t).(type) {
	case *hir.EmptyType:
		return self.codegenEmptyType()
	case *hir.SintType:
		return self.codegenSintType(t)
	case *hir.UintType:
		return self.codegenUintType(t)
	case *hir.FloatType:
		return self.codegenFloatType(t)
	case *hir.FuncType:
		return self.codegenFuncType(t)
	case *hir.ArrayType:
		return self.codegenArrayType(t)
	case *hir.TupleType:
		return self.codegenTupleType(t)
	case *hir.CustomType:
		return self.codegenCustomType(t)
	case *hir.UnionType:
		return self.codegenUnionType(t)
	case *hir.RefType:
		return self.codegenRefType(t)
	case *hir.StructType:
		return self.codegenStructType(t)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenEmptyType() mir.VoidType {
	return self.ctx.Void()
}

func (self *CodeGenerator) codegenSintType(ir *hir.SintType) mir.SintType {
	return self.ctx.NewSintType(stlos.Size(ir.Bits))
}

func (self *CodeGenerator) codegenUintType(ir *hir.UintType) mir.UintType {
	return self.ctx.NewUintType(stlos.Size(ir.Bits))
}

func (self *CodeGenerator) codegenFloatType(ir *hir.FloatType) mir.FloatType {
	switch ir.Bits {
	case 16:
		return self.ctx.F32()
	case 32:
		return self.ctx.F32()
	case 64:
		return self.ctx.F64()
	case 128:
		return self.ctx.F128()
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFuncType(ir *hir.FuncType) mir.FuncType {
	ret := self.codegenType(ir.Ret)
	params := make([]mir.Type, len(ir.Params))
	for i, p := range ir.Params{
		params[i] = self.codegenType(p)
	}
	return self.ctx.NewFuncType(false, ret, params...)
}

func (self *CodeGenerator) codegenArrayType(ir *hir.ArrayType) mir.ArrayType {
	elem := self.codegenType(ir.Elem)
	return self.ctx.NewArrayType(uint(ir.Size), elem)
}

func (self *CodeGenerator) codegenTupleType(ir *hir.TupleType) mir.StructType {
	elems := make([]mir.Type, len(ir.Elems))
	for i, e := range ir.Elems{
		elems[i] = self.codegenType(e)
	}
	return self.ctx.NewStructType(elems...)
}

func (self *CodeGenerator) codegenCustomType(ir *hir.CustomType) mir.Type {
	return self.types.Get(ir)
}

func (self *CodeGenerator) codegenStructType(ir *hir.StructType) mir.StructType {
	fields := stlslices.Map(ir.Fields.Values().ToSlice(), func(_ int, e hir.Field) mir.Type {
		return self.codegenType(e.Type)
	})
	return self.ctx.NewStructType(fields...)
}

func (self *CodeGenerator) codegenUnionType(ir *hir.UnionType) mir.StructType {
	var maxSizeType mir.Type
	var maxSize stlos.Size
	for _, e := range ir.Elems{
		et := self.codegenType(e)
		if esize := et.Size(); esize > maxSize {
			maxSizeType, maxSize = et, esize
		}
	}
	return self.ctx.NewStructType(maxSizeType, self.ctx.U8())
}

func (self *CodeGenerator) codegenRefType(ir *hir.RefType) mir.PtrType {
	elem := self.codegenType(ir.Elem)
	return self.ctx.NewPtrType(elem)
}
