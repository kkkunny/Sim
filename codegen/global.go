package codegen

import (
	"github.com/kkkunny/llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenGlobal(node mean.Global) {
	switch globalNode := node.(type) {
	case *mean.FuncDef:
		self.codegenFuncDef(globalNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFuncDef(node *mean.FuncDef) {
	ft := self.codegenType(node.GetType())
	f := llvm.AddFunction(self.module, node.Name, ft)

	self.builder.SetInsertPointAtEnd(llvm.AddBasicBlock(f, "entry"))

	body := self.codegenBlock(node.Body)
	self.builder.CreateBr(body)
}
