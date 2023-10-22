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
		return self.codegenBoolType(typeNode)
	case *mean.ArrayType:
		return self.codegenArrayType(typeNode)
	case *mean.TupleType:
		return self.codegenTupleType(typeNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenEmptyType(node *mean.EmptyType) llvm.VoidType {
	return self.ctx.VoidType()
}

func (self *CodeGenerator) codegenIntType(node mean.IntType) llvm.IntegerType {
	return self.ctx.IntegerType(uint32(node.GetBits()))
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
	return self.ctx.FunctionType(ret, nil, false)
}

func (self *CodeGenerator) codegenFuncTypePtr(node *mean.FuncType) llvm.PointerType {
	return self.ctx.PointerType(self.codegenFuncType(node))
}

func (self *CodeGenerator) codegenBoolType(node *mean.BoolType) llvm.IntegerType {
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
