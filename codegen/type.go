package codegen

import (
	"github.com/kkkunny/llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenType(node mean.Type) llvm.Type {
	switch node.(type) {
	case *mean.EmptyType:
		return self.ctx.VoidType()
	default:
		panic("unreachable")
	}
}
