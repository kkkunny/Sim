package llvm

import "github.com/kkkunny/Sim/mir"

func (self *LLVMOutputer) codegenBlock(ir *mir.Block){
	block := self.blocks.Get(ir)
	self.builder.MoveToAfter(block)
	for iter:=ir.Stmts().Iterator(); iter.Next(); {
		self.codegenStmt(iter.Value())
	}
}
