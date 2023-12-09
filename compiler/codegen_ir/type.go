package codegen_ir

import (
	stlos "github.com/kkkunny/stl/os"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) codegenType(node mean.Type) mir.Type {
	switch typeNode := node.(type) {
	case *mean.EmptyType:
		return self.codegenEmptyType(typeNode)
	case *mean.SintType:
		return self.codegenSintType(typeNode)
	case *mean.UintType:
		return self.codegenUintType(typeNode)
	case *mean.FloatType:
		return self.codegenFloatType(typeNode)
	case *mean.FuncType:
		return self.codegenFuncType(typeNode)
	case *mean.BoolType:
		return self.codegenBoolType()
	case *mean.ArrayType:
		return self.codegenArrayType(typeNode)
	case *mean.TupleType:
		return self.codegenTupleType(typeNode)
	case *mean.StructType:
		return self.codegenStructType(typeNode)
	case *mean.StringType:
		return self.codegenStringType()
	case *mean.UnionType:
		return self.codegenUnionType(typeNode)
	case *mean.PtrType:
		return self.codegenPtrType(typeNode)
	case *mean.RefType:
		return self.codegenRefType(typeNode)
	case *mean.GenericParam:
		return self.codegenGenericParam(typeNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenEmptyType(_ *mean.EmptyType) mir.VoidType {
	return self.ctx.Void()
}

func (self *CodeGenerator) codegenSintType(node *mean.SintType) mir.SintType {
	bits := node.GetBits()
	if bits == 0{
		return self.ctx.Isize()
	}
	return self.ctx.NewSintType(stlos.Size(bits))
}

func (self *CodeGenerator) codegenUintType(node *mean.UintType) mir.UintType {
	bits := node.GetBits()
	if bits == 0{
		return self.ctx.Usize()
	}
	return self.ctx.NewUintType(stlos.Size(bits))
}

func (self *CodeGenerator) codegenIntType(node mean.IntType) mir.IntType {
	switch intNode := node.(type) {
	case *mean.SintType:
		return self.codegenSintType(intNode)
	case *mean.UintType:
		return self.codegenUintType(intNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFloatType(node *mean.FloatType) mir.FloatType {
	switch node.Bits {
	case 32:
		return self.ctx.F32()
	case 64:
		return self.ctx.F64()
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFuncType(node *mean.FuncType) mir.FuncType {
	ret := self.codegenType(node.Ret)
	params := lo.Map(node.Params, func(item mean.Type, index int) mir.Type {
		return self.codegenType(item)
	})
	return self.ctx.NewFuncType(ret, params...)
}

func (self *CodeGenerator) codegenBoolType() mir.UintType {
	return self.ctx.NewUintType(1)
}

func (self *CodeGenerator) codegenArrayType(node *mean.ArrayType) mir.ArrayType {
	return self.ctx.NewArrayType(node.Size, self.codegenType(node.Elem))
}

func (self *CodeGenerator) codegenTupleType(node *mean.TupleType) mir.StructType {
	elems := lo.Map(node.Elems, func(item mean.Type, index int) mir.Type {
		return self.codegenType(item)
	})
	return self.ctx.NewStructType(elems...)
}

func (self *CodeGenerator) codegenStructType(node *mean.StructType) mir.StructType {
	return self.structs.Get(node)
}

func (self *CodeGenerator) codegenStringType() mir.StructType {
	st, ok := self.module.NamedStructType("str")
	if ok {
		return st
	}
	return self.module.NewNamedStructType("str", self.ctx.NewPtrType(self.ctx.U8()), self.ctx.Usize())
}

func (self *CodeGenerator) codegenUnionType(node *mean.UnionType) mir.StructType {
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

func (self *CodeGenerator) codegenPtrType(node *mean.PtrType) mir.PtrType {
	return self.ctx.NewPtrType(self.codegenType(node.Elem))
}

func (self *CodeGenerator) codegenRefType(node *mean.RefType) mir.PtrType {
	return self.ctx.NewPtrType(self.codegenType(node.Elem))
}

func (self *CodeGenerator) codegenGenericParam(node *mean.GenericParam)mir.Type{
	return self.codegenType(self.genericParams.Get(node))
}
