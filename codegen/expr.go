package codegen

import (
	"github.com/kkkunny/llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenExpr(node mean.Expr) llvm.Value {
	switch exprNode := node.(type) {
	case *mean.Integer:
		return self.codegenInteger(exprNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenInteger(node *mean.Integer) llvm.Value {
	t := self.codegenType(node.GetType())
	return llvm.ConstInt(t, node.Value.Uint64(), node.Type.HasSign())
}
