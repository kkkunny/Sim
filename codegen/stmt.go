package codegen

import (
	"github.com/kkkunny/llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenStmt(node mean.Stmt) {

}

func (self *CodeGenerator) codegenBlock(node *mean.Block) llvm.BasicBlock {
	from := self.builder.GetInsertBlock()
	defer self.builder.SetInsertPointAtEnd(from)

	block := llvm.AddBasicBlock(from.Parent(), "")
	self.builder.SetInsertPointAtEnd(block)

	self.builder.CreateRetVoid()

	return block
}
