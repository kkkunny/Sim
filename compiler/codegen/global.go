package codegen

import (
	"github.com/kkkunny/go-llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) declStructDef(node *mean.StructDef) {
	self.ctx.NamedStructType(node.Name, false)
}

func (self *CodeGenerator) defStructDef(node *mean.StructDef) {
	st := self.ctx.GetTypeByName(node.Name)
	fields := make([]llvm.Type, node.Fields.Length())
	var i int
	for iter := node.Fields.Values().Iterator(); iter.Next(); i++ {
		fields[i] = self.codegenType(iter.Value())
	}
	st.SetElems(false, fields...)
}

func (self *CodeGenerator) codegenGlobalDecl(node mean.Global) {
	switch globalNode := node.(type) {
	case *mean.FuncDef:
		self.declFuncDef(globalNode)
	case *mean.Variable:
		self.declGlobalVariable(globalNode)
	case *mean.StructDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(node *mean.FuncDef) {
	ft := self.codegenFuncType(node.GetType().(*mean.FuncType))
	self.values[node] = self.newFunction("", ft)
}

func (self *CodeGenerator) declGlobalVariable(node *mean.Variable) {
	t := self.codegenType(node.Type)
	v := self.module.NewGlobal("", t)
	self.values[node] = v
	v.SetInitializer(self.codegenZero(&mean.Zero{Type: node.GetType()}))
}

func (self *CodeGenerator) codegenGlobalDef(node mean.Global) {
	switch globalNode := node.(type) {
	case *mean.FuncDef:
		if globalNode.Body.IsNone() {
			self.defFuncDecl(globalNode)
		} else {
			self.defFuncDef(globalNode)
		}
	case *mean.StructDef:
		self.defStructDef(globalNode)
	case *mean.Variable:
		self.defGlobalVariable(globalNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) defFuncDef(node *mean.FuncDef) {
	f := self.values[node].(llvm.Function)

	self.builder.MoveToAfter(f.NewBlock("entry"))
	for i, p := range f.Params() {
		np := node.Params[i]

		pt := self.codegenType(np.GetType())
		param := self.builder.CreateAlloca("", pt)

		self.builder.CreateStore(p, param)

		self.values[np] = param
	}

	block, _ := self.codegenBlock(node.Body.MustValue(), nil)
	self.builder.CreateBr(block)
}

func (self *CodeGenerator) defFuncDecl(node *mean.FuncDef) {
	_ = self.values[node].(llvm.Function)
}

func (self *CodeGenerator) defGlobalVariable(node *mean.Variable) {
	v := self.values[node].(llvm.GlobalValue)
	self.builder.MoveToAfter(self.getMainFunction().EntryBlock())
	self.builder.CreateStore(self.codegenExpr(node.Value, true), v)
}
