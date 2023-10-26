package codegen

import (
	"github.com/kkkunny/go-llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenStmt(node mean.Stmt) {
	switch stmtNode := node.(type) {
	case *mean.Return:
		self.codegenReturn(stmtNode)
	case *mean.Variable:
		self.codegenVariable(stmtNode)
	case *mean.If:
		self.codegenIf(stmtNode)
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

func (self *CodeGenerator) codegenVariable(node *mean.Variable) {
	t := self.codegenType(node.Type)
	value := self.codegenExpr(node.Value)
	ptr := self.builder.CreateAlloca("", t)
	self.builder.CreateStore(value, ptr)
	self.values[node] = ptr
}

func (self *CodeGenerator) codegenIf(node *mean.If) {
	cond := self.codegenExpr(node.Cond)
	trueBlock := self.codegenBlock(node.Body)
	falseBlock := trueBlock.Belong().NewBlock("")
	self.builder.CreateCondBr(cond, trueBlock, falseBlock)
	self.builder.MoveToAfter(falseBlock)
}
