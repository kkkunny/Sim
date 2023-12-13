package codegen_ir

import (
	stlos "github.com/kkkunny/stl/os"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) codegenType(node hir.Type) mir.Type {
	switch typeNode := node.(type) {
	case *hir.EmptyType:
		return self.codegenEmptyType(typeNode)
	case *hir.SintType:
		return self.codegenSintType(typeNode)
	case *hir.UintType:
		return self.codegenUintType(typeNode)
	case *hir.FloatType:
		return self.codegenFloatType(typeNode)
	case *hir.FuncType:
		return self.codegenFuncType(typeNode)
	case *hir.BoolType:
		return self.codegenBoolType()
	case *hir.ArrayType:
		return self.codegenArrayType(typeNode)
	case *hir.TupleType:
		return self.codegenTupleType(typeNode)
	case *hir.StructType:
		return self.codegenStructType(typeNode)
	case *hir.StringType:
		return self.codegenStringType()
	case *hir.UnionType:
		return self.codegenUnionType(typeNode)
	case *hir.PtrType:
		return self.codegenPtrType(typeNode)
	case *hir.RefType:
		return self.codegenRefType(typeNode)
	case *hir.GenericParam:
		return self.codegenGenericParam(typeNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenEmptyType(_ *hir.EmptyType) mir.VoidType {
	return self.ctx.Void()
}

func (self *CodeGenerator) codegenSintType(node *hir.SintType) mir.SintType {
	bits := node.GetBits()
	if bits == 0{
		return self.ctx.Isize()
	}
	return self.ctx.NewSintType(stlos.Size(bits))
}

func (self *CodeGenerator) codegenUintType(node *hir.UintType) mir.UintType {
	bits := node.GetBits()
	if bits == 0{
		return self.ctx.Usize()
	}
	return self.ctx.NewUintType(stlos.Size(bits))
}

func (self *CodeGenerator) codegenIntType(node hir.IntType) mir.IntType {
	switch intNode := node.(type) {
	case *hir.SintType:
		return self.codegenSintType(intNode)
	case *hir.UintType:
		return self.codegenUintType(intNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFloatType(node *hir.FloatType) mir.FloatType {
	switch node.Bits {
	case 32:
		return self.ctx.F32()
	case 64:
		return self.ctx.F64()
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFuncType(node *hir.FuncType) mir.FuncType {
	ret := self.codegenType(node.Ret)
	params := lo.Map(node.Params, func(item hir.Type, index int) mir.Type {
		return self.codegenType(item)
	})
	return self.ctx.NewFuncType(ret, params...)
}

func (self *CodeGenerator) codegenBoolType() mir.UintType {
	return self.ctx.NewUintType(1)
}

func (self *CodeGenerator) codegenArrayType(node *hir.ArrayType) mir.ArrayType {
	return self.ctx.NewArrayType(node.Size, self.codegenType(node.Elem))
}

func (self *CodeGenerator) codegenTupleType(node *hir.TupleType) mir.StructType {
	elems := lo.Map(node.Elems, func(item hir.Type, index int) mir.Type {
		return self.codegenType(item)
	})
	return self.ctx.NewStructType(elems...)
}

func (self *CodeGenerator) codegenStructType(node *hir.StructType) mir.StructType {
	return self.structs.Get(node)
}

func (self *CodeGenerator) codegenStringType() mir.StructType {
	st, ok := self.module.NamedStructType("str")
	if ok {
		return st
	}
	return self.module.NewNamedStructType("str", self.ctx.NewPtrType(self.ctx.U8()), self.ctx.Usize())
}

func (self *CodeGenerator) codegenUnionType(node *hir.UnionType) mir.StructType {
	var maxSizeType mir.Type
	var maxSize stlos.Size
	for _, e := range node.Elems{
		et := self.codegenType(e)
		if esize := et.Size(); esize > maxSize {
			maxSizeType, maxSize = et, esize
		}
	}
	return self.ctx.NewStructType(maxSizeType, self.ctx.U8())
}

func (self *CodeGenerator) codegenPtrType(node *hir.PtrType) mir.PtrType {
	return self.ctx.NewPtrType(self.codegenType(node.Elem))
}

func (self *CodeGenerator) codegenRefType(node *hir.RefType) mir.PtrType {
	return self.ctx.NewPtrType(self.codegenType(node.Elem))
}

func (self *CodeGenerator) codegenGenericParam(node *hir.GenericParam)mir.Type{
	return self.codegenType(self.genericParams.Get(node))
}
