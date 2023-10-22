package codegen

import (
	"github.com/kkkunny/go-llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenStmt(node mean.Stmt) {
	switch stmtNode := node.(type) {
	case *mean.Return:
		self.codegenReturn(stmtNode)
	case mean.Expr:
		self.codegenExpr(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenBlock(node *mean.Block) llvm.Block {
	from := self.builder.CurrentBlock()
	defer self.builder.MoveToAfter(from)

	block := from.Belong().NewBlock("")
	self.builder.MoveToAfter(block)

	for iter := node.Stmts.Iterator(); iter.Next(); {
		self.codegenStmt(iter.Value())
	}

	return block
}

func (self *CodeGenerator) codegenReturn(node *mean.Return) {
	if v, ok := node.Value.Value(); ok {
		v := self.codegenExpr(v)
		self.builder.CreateRet(&v)
	} else {
		self.builder.CreateRet(nil)
	}
}
