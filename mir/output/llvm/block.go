package llvm

import "github.com/kkkunny/Sim/mir"

func (self *LLVMOutputer) codegenBlock(ir *mir.Block){
	block := self.blocks.Get(ir)
	self.builder.MoveToAfter(block)
	for cursor:=ir.Stmts().Front(); cursor!=nil; cursor=cursor.Next(){
		self.codegenStmt(cursor.Value)
	}
}
