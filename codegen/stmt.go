package codegen

import (
	"github.com/kkkunny/llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenStmt(node mean.Stmt) {
	switch stmtNode := node.(type) {
	case *mean.Return:
		self.codegenReturn(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenBlock(node *mean.Block) llvm.BasicBlock {
	from := self.builder.GetInsertBlock()
	defer self.builder.SetInsertPointAtEnd(from)

	block := llvm.AddBasicBlock(from.Parent(), "")
	self.builder.SetInsertPointAtEnd(block)

	for iter := node.Stmts.Iterator(); iter.Next(); {
		self.codegenStmt(iter.Value())
	}

	return block
}

func (self *CodeGenerator) codegenReturn(node *mean.Return) {
	if v, ok := node.Value.Value(); ok {
		self.builder.CreateRet(self.codegenExpr(v))
	} else {
		self.builder.CreateRetVoid()
	}
}
