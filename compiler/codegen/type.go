package codegen

import (
	"github.com/kkkunny/go-llvm"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenType(node mean.Type) llvm.Type {
	switch typeNode := node.(type) {
	case *mean.EmptyType:
		return self.codegenEmptyType(typeNode)
	case mean.IntType:
		return self.codegenIntType(typeNode)
	case *mean.FloatType:
		return self.codegenFloatType(typeNode)
	case *mean.FuncType:
		return self.codegenFuncTypePtr(typeNode)
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

func (self *CodeGenerator) codegenEmptyType(_ *mean.EmptyType) llvm.VoidType {
	return self.ctx.VoidType()
}

func (self *CodeGenerator) codegenIntType(node mean.IntType) llvm.IntegerType {
	bits := node.GetBits()
	if bits == 0{
		return self.ctx.IntPtrType(self.target)
	}
	return self.ctx.IntegerType(uint32(bits))
}

func (self *CodeGenerator) codegenFloatType(node *mean.FloatType) llvm.FloatType {
	switch node.Bits {
	case 32:
		return self.ctx.FloatType(llvm.FloatTypeKindFloat)
	case 64:
		return self.ctx.FloatType(llvm.FloatTypeKindDouble)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFuncType(node *mean.FuncType) llvm.FunctionType {
	ret := self.codegenType(node.Ret)
	params := lo.Map(node.Params, func(item mean.Type, index int) llvm.Type {
		return self.codegenType(item)
	})
	return self.ctx.FunctionType(false, ret, params...)
}

func (self *CodeGenerator) codegenFuncTypePtr(node *mean.FuncType) llvm.PointerType {
	return self.ctx.PointerType(self.codegenFuncType(node))
}

func (self *CodeGenerator) codegenBoolType() llvm.IntegerType {
	return self.ctx.IntegerType(1)
}

func (self *CodeGenerator) codegenArrayType(node *mean.ArrayType) llvm.ArrayType {
	elem := self.codegenType(node.Elem)
	return self.ctx.ArrayType(elem, uint32(node.Size))
}

func (self *CodeGenerator) codegenTupleType(node *mean.TupleType) llvm.StructType {
	elems := lo.Map(node.Elems, func(item mean.Type, index int) llvm.Type {
		return self.codegenType(item)
	})
	return self.ctx.StructType(false, elems...)
}

func (self *CodeGenerator) codegenStructType(node *mean.StructType) llvm.StructType {
	return self.structs.Get(node)
}

func (self *CodeGenerator) codegenStringType() llvm.StructType {
	st := self.ctx.GetTypeByName("str")
	if st != nil {
		return *st
	}
	return self.ctx.NamedStructType("str", false, self.ctx.PointerType(self.ctx.IntegerType(8)), self.ctx.IntPtrType(self.target))
}

func (self *CodeGenerator) codegenUnionType(node *mean.UnionType) llvm.StructType {
	var maxSizeType llvm.Type
	var maxSize uint
	for _, e := range node.Elems{
		et := self.codegenType(e)
		if esize := self.target.GetSizeOfType(et); esize > maxSize {
			maxSizeType, maxSize = et, esize
		}
	}
	return self.ctx.StructType(true, maxSizeType, self.ctx.IntegerType(8))
}

func (self *CodeGenerator) codegenPtrType(node *mean.PtrType) llvm.PointerType {
	return self.ctx.PointerType(self.codegenType(node.Elem))
}

func (self *CodeGenerator) codegenRefType(node *mean.RefType) llvm.PointerType {
	return self.ctx.PointerType(self.codegenType(node.Elem))
}

func (self *CodeGenerator) codegenGenericParam(node *mean.GenericParam)llvm.Type{
	return self.codegenType(self.genericParams.Get(node))
}
