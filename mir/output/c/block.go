package c

import (
	"github.com/kkkunny/Sim/mir"
)

func (self *COutputer) codegenBlock(ir *mir.Block){
	name := self.blocks.Get(ir)
	self.varDefBuffer.WriteString(name)
	self.varDefBuffer.WriteString(":;\n")

	for cursor:=ir.Stmts().Front(); cursor!=nil; cursor=cursor.Next(){
		self.varDefBuffer.WriteByte('\t')
		self.codegenStmt(cursor.Value)
		self.varDefBuffer.WriteString(";\n")
	}
}
