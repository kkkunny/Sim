package codegen_ir

import (
	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) declStructDef(node *mean.StructDef) {
	self.structs.Set(node, self.module.NewNamedStructType(""))
}

func (self *CodeGenerator) defStructDef(node *mean.StructDef) {
	st := self.structs.Get(node)
	fields := make([]mir.Type, node.Fields.Length())
	var i int
	for iter := node.Fields.Values().Iterator(); iter.Next(); i++ {
		fields[i] = self.codegenType(iter.Value().Second)
	}
	st.SetElems(fields...)
}

func (self *CodeGenerator) codegenGlobalDecl(node mean.Global) {
	switch globalNode := node.(type) {
	case *mean.FuncDef:
		self.declFuncDef(globalNode)
	case *mean.MethodDef:
		self.declMethodDef(globalNode)
	case *mean.Variable:
		self.declGlobalVariable(globalNode)
	case *mean.GenericFuncDef:
		self.declGenericFuncDef(globalNode)
	case *mean.StructDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(node *mean.FuncDef) {
	ft := self.codegenFuncType(node.GetType().(*mean.FuncType))
	f := self.module.NewFunction(node.ExternName, ft)
	self.values.Set(node, f)
}

func (self *CodeGenerator) declMethodDef(node *mean.MethodDef) {
	ft := self.codegenFuncType(node.GetType().(*mean.FuncType))
	f := self.module.NewFunction("", ft)
	self.values.Set(node, f)
}

func (self *CodeGenerator) declGlobalVariable(node *mean.Variable) {
	t := self.codegenType(node.Type)
	v := self.module.NewGlobalVariable("", t, mir.NewZero(self.codegenType(node.GetType())))
	self.values.Set(node, v)
}

func (self *CodeGenerator) declGenericFuncDef(node *mean.GenericFuncDef) {
	for iter:=node.Instances.Values().Iterator(); iter.Next(); {
		inst := iter.Value()
		ft := self.codegenFuncType(inst.GetType().(*mean.FuncType))
		f := self.module.NewFunction("", ft)
		self.values.Set(inst, f)
	}
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
	case *mean.GenericFuncDef:
		self.defGenericFuncDef(globalNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) defFuncDef(node *mean.FuncDef) {
	f := self.values.Get(node).(*mir.Function)
	self.builder.MoveTo(f.NewBlock())
	for i, p := range f.Params() {
		self.values.Set(node.Params[i], p)
	}
	block, _ := self.codegenBlock(node.Body.MustValue(), nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defMethodDef(node *mean.MethodDef) {
	f := self.values.Get(node).(*mir.Function)
	self.builder.MoveTo(f.NewBlock())
	paramNodes := append([]*mean.Param{node.SelfParam}, node.Params...)
	for i, p := range f.Params() {
		self.values.Set(paramNodes[i], p)
	}
	block, _ := self.codegenBlock(node.Body, nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defFuncDecl(node *mean.FuncDef) {
	_ = self.values.Get(node).(*mir.Function)
}

func (self *CodeGenerator) defGlobalVariable(node *mean.Variable) {
	gv := self.values.Get(node).(*mir.GlobalVariable)
	self.builder.MoveTo(self.getInitFunction().Blocks().Front().Value)
	value := self.codegenExpr(node.Value, true)
	if constValue, ok := value.(mir.Const); ok {
		gv.SetValue(constValue)
	} else {
		self.builder.BuildStore(value, gv)
	}
}

func (self *CodeGenerator) defGenericFuncDef(node *mean.GenericFuncDef) {
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
