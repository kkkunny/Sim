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
	case *mean.MethodDef:
		self.declMethodDef(globalNode)
	case *mean.Variable:
		self.declGlobalVariable(globalNode)
	case *mean.StructDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(node *mean.FuncDef) {
	ft := self.codegenFuncType(node.GetType().(*mean.FuncType))
	f := self.newFunction(node.ExternName, ft)
	if node.ExternName == "" {
		f.SetLinkage(llvm.InternalLinkage)
	} else {
		f.SetLinkage(llvm.ExternalLinkage)
	}
	self.values[node] = f
}

func (self *CodeGenerator) declMethodDef(node *mean.MethodDef) {
	ft := self.codegenFuncType(node.GetType().(*mean.FuncType))
	f := self.newFunction("", ft)
	f.SetLinkage(llvm.InternalLinkage)
	self.values[node] = f
}

func (self *CodeGenerator) declGlobalVariable(node *mean.Variable) {
	t := self.codegenType(node.Type)
	v := self.module.NewGlobal("", t)
	if node.ExternName == "" {
		v.SetLinkage(llvm.InternalLinkage)
	} else {
		v.SetLinkage(llvm.ExternalLinkage)
	}
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
	case *mean.MethodDef:
		self.defMethodDef(globalNode)
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
	self.enterFunction(self.codegenFuncType(node.GetType().(*mean.FuncType)), f, node.Params)
	block, _ := self.codegenBlock(node.Body.MustValue(), nil)
	self.builder.CreateBr(block)
}

func (self *CodeGenerator) defMethodDef(node *mean.MethodDef) {
	f := self.values[node].(llvm.Function)
	self.builder.MoveToAfter(f.NewBlock("entry"))
	self.enterFunction(self.codegenFuncType(node.GetType().(*mean.FuncType)), f, append([]*mean.Param{node.SelfParam}, node.Params...))
	block, _ := self.codegenBlock(node.Body, nil)
	self.builder.CreateBr(block)
}

func (self *CodeGenerator) defFuncDecl(node *mean.FuncDef) {
	_ = self.values[node].(llvm.Function)
}

func (self *CodeGenerator) defGlobalVariable(node *mean.Variable) {
	gv := self.values[node].(llvm.GlobalValue)
	self.builder.MoveToAfter(self.getInitFunction().EntryBlock())
	value := self.codegenExpr(node.Value, true)
	if constValue, ok := value.(llvm.Constant); ok {
		gv.SetInitializer(constValue)
	} else {
		self.builder.CreateStore(value, gv)
	}
}
