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
	for i, p := range f.Params() {
		np := node.Params[i]

		pt := self.codegenType(np.GetType())
		param := self.builder.CreateAlloca("", pt)

		self.builder.CreateStore(p, param)

		self.values[np] = param
	}

	body := self.codegenBlock(node.Body)
	self.builder.CreateBr(body)
}
