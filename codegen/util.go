package codegen

import (
	"github.com/kkkunny/go-llvm"
)

func (self *CodeGenerator) buildArrayIndex(t llvm.ArrayType, from, index llvm.Value, load bool) llvm.Value {
	if _, ptr := from.Type().(llvm.PointerType); ptr {
		elemPtr := self.builder.CreateInBoundsGEP("", t, from, self.ctx.ConstInteger(index.Type().(llvm.IntegerType), 0), index)
		if !load {
			return elemPtr
		}
		return self.builder.CreateLoad("", t.Element(), elemPtr)
	} else {
		if !load {
			panic("unreachable")
		}
		return self.builder.CreateExtractElement("", from, index)
	}
}

func (self *CodeGenerator) buildArrayIndexWith(t llvm.ArrayType, from llvm.Value, index uint, load bool) llvm.Value {
	return self.buildArrayIndex(t, from, self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(index)), load)
}

func (self *CodeGenerator) buildStructIndex(t llvm.StructType, from llvm.Value, index uint, load bool) llvm.Value {
	if _, ptr := from.Type().(llvm.PointerType); ptr {
		elemPtr := self.builder.CreateStructGEP("", t, from, index)
		if !load {
			return elemPtr
		}
		return self.builder.CreateLoad("", t.GetElem(uint32(index)), elemPtr)
	} else {
		if !load {
			panic("unreachable")
		}
		return self.builder.CreateExtractValue("", from, index)
	}
}

func (self *CodeGenerator) buildEqual(l, r llvm.Value) llvm.Value {
	t := l.Type()

	switch t.(type) {
	case llvm.IntegerType, llvm.PointerType:
		return self.builder.CreateIntCmp("", llvm.IntEQ, l, r)
	case llvm.FloatType:
		return self.builder.CreateFloatCmp("", llvm.FloatOEQ, l, r)
	case llvm.ArrayType:
		// TODO
		panic("unreachable")
	case llvm.StructType:
		// TODO
		panic("unreachable")
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) buildNotEqual(l, r llvm.Value) llvm.Value {
	t := l.Type()

	switch t.(type) {
	case llvm.IntegerType, llvm.PointerType:
		return self.builder.CreateIntCmp("", llvm.IntNE, l, r)
	case llvm.FloatType:
		return self.builder.CreateFloatCmp("", llvm.FloatUNE, l, r)
	case llvm.ArrayType, llvm.StructType:
		return self.builder.CreateNot("", self.buildEqual(l, r))
	default:
		panic("unreachable")
	}
}
