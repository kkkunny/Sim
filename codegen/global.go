package codegen

import (
	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenGlobalDecl(node mean.Global) {
	switch globalNode := node.(type) {
	case *mean.FuncDef:
		self.declFuncDef(globalNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(node *mean.FuncDef) {
	ft := self.codegenFuncType(node.GetType().(*mean.FuncType))
	self.module.NewFunction(node.Name, ft)
}

func (self *CodeGenerator) codegenGlobalDef(node mean.Global) {
	switch globalNode := node.(type) {
	case *mean.FuncDef:
		self.defFuncDef(globalNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) defFuncDef(node *mean.FuncDef) {
	f := self.module.GetFunction(node.Name)

	self.builder.MoveToAfter(f.NewBlock("entry"))

	body := self.codegenBlock(node.Body)
	self.builder.CreateBr(body)
}
