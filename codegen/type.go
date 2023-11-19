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
	case *mean.StructType:
		return self.codegenStructType(typeNode)
	case *mean.StringType:
		return self.codegenStringType(typeNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenEmptyType(_ *mean.EmptyType) llvm.VoidType {
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
	params := lo.Map(node.Params, func(item mean.Type, index int) llvm.Type {
		return self.codegenType(item)
	})
	return self.ctx.FunctionType(ret, params, false)
}

func (self *CodeGenerator) codegenFuncTypePtr(node *mean.FuncType) llvm.PointerType {
	return self.ctx.PointerType(self.codegenFuncType(node))
}

func (self *CodeGenerator) codegenBoolType(_ *mean.BoolType) llvm.IntegerType {
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
	fields := make([]llvm.Type, node.Fields.Length())
	var i int
	for iter := node.Fields.Iterator(); iter.Next(); i++ {
		fields[i] = self.codegenType(iter.Value().Second)
	}
	return self.ctx.StructType(false, fields...)
}

func (self *CodeGenerator) codegenStringType(_ *mean.StringType) llvm.StructType {
	st := self.ctx.GetTypeByName("str")
	if st != nil {
		return *st
	}
	return self.ctx.NamedStructType("str", false, self.ctx.PointerType(self.ctx.IntegerType(8)), self.ctx.IntPtrType(self.target))
}
