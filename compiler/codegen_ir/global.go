package codegen_ir

import (
	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) declStructDef(node *hir.StructDef) {
	self.structs.Set(node, self.module.NewNamedStructType(""))
}

func (self *CodeGenerator) defStructDef(node *hir.StructDef) {
	st := self.structs.Get(node)
	fields := make([]mir.Type, node.Fields.Length())
	var i int
	for iter := node.Fields.Values().Iterator(); iter.Next(); i++ {
		fields[i] = self.codegenType(iter.Value().Second)
	}
	st.SetElems(fields...)
}

func (self *CodeGenerator) codegenGlobalDecl(node hir.Global) {
	switch globalNode := node.(type) {
	case *hir.FuncDef:
		self.declFuncDef(globalNode)
	case *hir.MethodDef:
		self.declMethodDef(globalNode)
	case *hir.VarDef:
		self.declGlobalVariable(globalNode)
	case *hir.GenericFuncDef:
		self.declGenericFuncDef(globalNode)
	case *hir.StructDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(node *hir.FuncDef) {
	ft := self.codegenFuncType(node.GetType().(*hir.FuncType))
	f := self.module.NewFunction(node.ExternName, ft)
	self.values.Set(node, f)
}

func (self *CodeGenerator) declMethodDef(node *hir.MethodDef) {
	ft := self.codegenFuncType(node.GetType().(*hir.FuncType))
	f := self.module.NewFunction("", ft)
	self.values.Set(node, f)
}

func (self *CodeGenerator) declGlobalVariable(node *hir.VarDef) {
	t := self.codegenType(node.Type)
	v := self.module.NewGlobalVariable("", t, mir.NewZero(self.codegenType(node.GetType())))
	self.values.Set(node, v)
}

func (self *CodeGenerator) declGenericFuncDef(node *hir.GenericFuncDef) {
	for iter:=node.Instances.Values().Iterator(); iter.Next(); {
		inst := iter.Value()
		ft := self.codegenFuncType(inst.GetType().(*hir.FuncType))
		f := self.module.NewFunction("", ft)
		self.values.Set(inst, f)
	}
}

func (self *CodeGenerator) codegenGlobalDef(node hir.Global) {
	switch globalNode := node.(type) {
	case *hir.FuncDef:
		if globalNode.Body.IsNone() {
			self.defFuncDecl(globalNode)
		} else {
			self.defFuncDef(globalNode)
		}
	case *hir.MethodDef:
		self.defMethodDef(globalNode)
	case *hir.StructDef:
		self.defStructDef(globalNode)
	case *hir.VarDef:
		self.defGlobalVariable(globalNode)
	case *hir.GenericFuncDef:
		self.defGenericFuncDef(globalNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) defFuncDef(node *hir.FuncDef) {
	f := self.values.Get(node).(*mir.Function)
	self.builder.MoveTo(f.NewBlock())
	for i, p := range f.Params() {
		self.values.Set(node.Params[i], p)
	}
	block, _ := self.codegenBlock(node.Body.MustValue(), nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defMethodDef(node *hir.MethodDef) {
	f := self.values.Get(node).(*mir.Function)
	self.builder.MoveTo(f.NewBlock())
	paramNodes := append([]*hir.Param{node.SelfParam}, node.Params...)
	for i, p := range f.Params() {
		self.values.Set(paramNodes[i], p)
	}
	block, _ := self.codegenBlock(node.Body, nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defFuncDecl(node *hir.FuncDef) {
	_ = self.values.Get(node).(*mir.Function)
}

func (self *CodeGenerator) defGlobalVariable(node *hir.VarDef) {
	gv := self.values.Get(node).(*mir.GlobalVariable)
	self.builder.MoveTo(self.getInitFunction().Blocks().Front().Value)
	value := self.codegenExpr(node.Value, true)
	if constValue, ok := value.(mir.Const); ok {
		gv.SetValue(constValue)
	} else {
		self.builder.BuildStore(value, gv)
	}
}

func (self *CodeGenerator) defGenericFuncDef(node *hir.GenericFuncDef) {
	for iter:=node.Instances.Values().Iterator(); iter.Next(); {
		inst := iter.Value()
		f := self.values.Get(inst).(*mir.Function)
		self.builder.MoveTo(f.NewBlock())

		var i int
		for iter:=node.GenericParams.Values().Iterator(); iter.Next(); {
			self.genericParams.Set(iter.Value(), inst.Params[i])
			i++
		}

		for i, p := range f.Params() {
			self.values.Set(node.Params[i], p)
		}
		block, _ := self.codegenBlock(node.Body, nil)
		self.builder.BuildUnCondJump(block)

		self.genericParams.Clear()
	}
}
