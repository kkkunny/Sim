package codegen

import (
	"github.com/kkkunny/llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenType(node mean.Type) llvm.Type {
	switch typeNode := node.(type) {
	case *mean.EmptyType:
		return self.ctx.VoidType()
	case *mean.SintType:
		return self.ctx.IntType(int(typeNode.Bits))
	case *mean.FloatType:
		switch typeNode.Bits {
		case 64:
			return self.ctx.DoubleType()
		default:
			panic("unreachable")
		}
	case *mean.FuncType:
		ret := self.codegenType(typeNode.Ret)
		return llvm.FunctionType(ret, nil, false)
	default:
		panic("unreachable")
	}
}
