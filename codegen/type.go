package codegen

import (
	"github.com/kkkunny/llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenType(node mean.Type) llvm.Type {
	switch typeNode := node.(type) {
	case *mean.EmptyType:
		return self.ctx.VoidType()
	case mean.IntType:
		return self.ctx.IntType(int(typeNode.GetBits()))
	case *mean.FloatType:
		switch typeNode.Bits {
		case 32:
			return self.ctx.FloatType()
		case 64:
			return self.ctx.DoubleType()
		default:
			panic("unreachable")
		}
	case *mean.FuncType:
		ret := self.codegenType(typeNode.Ret)
		return llvm.FunctionType(ret, nil, false)
	case *mean.BoolType:
		return self.ctx.Int1Type()
	default:
		panic("unreachable")
	}
}
