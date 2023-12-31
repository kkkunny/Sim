package codegen_ir

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) declStructDef(ir *hir.StructDef) {
	self.structs.Set(ir.String(), self.module.NewNamedStructType(""))
}

func (self *CodeGenerator) defStructDef(ir *hir.StructDef) {
	st := self.structs.Get(ir.String())
	fields := make([]mir.Type, ir.Fields.Length())
	var i int
	for iter := ir.Fields.Values().Iterator(); iter.Next(); i++ {
		fields[i] = self.codegenType(iter.Value().Second)
	}
	st.SetElems(fields...)
}

func (self *CodeGenerator) declGenericStructDef(ir *hir.GenericStructInst) mir.StructType {
	return self.module.NewNamedStructType("")
}

func (self *CodeGenerator) defGenericStructDef(ir *hir.GenericStructInst, st mir.StructType) {
	stIr := ir.StructType()
	fields := make([]mir.Type, stIr.Fields.Length())
	var i int
	for iter := stIr.Fields.Values().Iterator(); iter.Next(); i++ {
		fields[i] = self.codegenType(iter.Value().Second)
	}
	st.SetElems(fields...)
}

func (self *CodeGenerator) codegenGlobalDecl(ir hir.Global) {
	switch global := ir.(type) {
	case *hir.FuncDef:
		self.declFuncDef(global)
	case *hir.MethodDef:
		self.declMethodDef(global)
	case *hir.VarDef:
		self.declGlobalVariable(global)
	case *hir.MultiVarDef:
		self.declMultiGlobalVariable(global)
	case *hir.StructDef, *hir.TypeAliasDef, *hir.GenericFuncDef, *hir.GenericStructDef, *hir.GenericStructMethodDef, *hir.GenericMethodDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(ir *hir.FuncDef) {
	ft := self.codegenType(ir.GetType()).(mir.FuncType)
	f := self.module.NewFunction(ir.ExternName, ft)
	if ir.NoReturn{
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.InlineControl.Value(); ok{
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
}

func (self *CodeGenerator) declMethodDef(ir *hir.MethodDef) {
	ft := self.codegenType(ir.GetType()).(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.NoReturn{
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.InlineControl.Value(); ok{
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
}

func (self *CodeGenerator) declGlobalVariable(ir *hir.VarDef) {
	t := self.codegenType(ir.Type)
	v := self.module.NewGlobalVariable("", t, mir.NewZero(self.codegenType(ir.GetType())))
	self.values.Set(ir, v)
}

func (self *CodeGenerator) declMultiGlobalVariable(ir *hir.MultiVarDef) {
	for _, varDef := range ir.Vars{
		self.declGlobalVariable(varDef)
	}
}

func (self *CodeGenerator) declGenericFuncDef(ir *hir.GenericFuncInst) *mir.Function {
	ft := self.codegenType(ir.GetType()).(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Define.NoReturn{
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.Define.InlineControl.Value(); ok{
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
	return f
}

func (self *CodeGenerator) declGenericStructMethodDef(ir *hir.GenericStructMethodInst) *mir.Function {
	ft := self.codegenType(ir.GetFuncType()).(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Define.NoReturn{
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.Define.InlineControl.Value(); ok{
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
	return f
}

func (self *CodeGenerator) declGenericMethodDef(ir *hir.GenericMethodInst) *mir.Function {
	ft := self.codegenType(ir.GetFuncType()).(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Define.NoReturn{
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.Define.InlineControl.Value(); ok{
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
	return f
}

func (self *CodeGenerator) codegenGlobalDef(ir hir.Global) {
	switch global := ir.(type) {
	case *hir.FuncDef:
		if global.Body.IsNone() {
			self.defFuncDecl(global)
		} else {
			self.defFuncDef(global)
		}
	case *hir.MethodDef:
		self.defMethodDef(global)
	case *hir.StructDef:
		self.defStructDef(global)
	case *hir.VarDef:
		self.defGlobalVariable(global)
	case *hir.MultiVarDef:
		self.defMultiGlobalVariable(global)
	case *hir.TypeAliasDef, *hir.GenericFuncDef, *hir.GenericStructDef, *hir.GenericStructMethodDef, *hir.GenericMethodDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) defFuncDef(ir *hir.FuncDef) {
	f := self.values.Get(ir).(*mir.Function)
	self.builder.MoveTo(f.NewBlock())
	for i, p := range f.Params() {
		self.values.Set(ir.Params[i], p)
	}
	block, _ := self.codegenBlock(ir.Body.MustValue(), nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defMethodDef(ir *hir.MethodDef) {
	f := self.values.Get(ir).(*mir.Function)
	self.builder.MoveTo(f.NewBlock())
	paramNodes := append([]*hir.Param{ir.SelfParam}, ir.Params...)
	for i, p := range f.Params() {
		self.values.Set(paramNodes[i], p)
	}
	block, _ := self.codegenBlock(ir.Body, nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defFuncDecl(ir *hir.FuncDef) {
	_ = self.values.Get(ir).(*mir.Function)
}

func (self *CodeGenerator) defGlobalVariable(ir *hir.VarDef) {
	gv := self.values.Get(ir).(*mir.GlobalVariable)
	self.builder.MoveTo(self.getInitFunction().Blocks().Front().Value)
	value := self.codegenExpr(ir.Value, true)
	if constValue, ok := value.(mir.Const); ok {
		gv.SetValue(constValue)
	} else {
		self.builder.BuildStore(value, gv)
	}
}

func (self *CodeGenerator) defMultiGlobalVariable(ir *hir.MultiVarDef) {
	if constant, ok := ir.Value.(*hir.Tuple); ok && len(constant.Elems) == len(ir.Vars){
		for i, varNode := range ir.Vars{
			varNode.Value = constant.Elems[i]
			self.defGlobalVariable(varNode)
		}
	}else{
		self.builder.MoveTo(self.getInitFunction().Blocks().Front().Value)
		self.codegenUnTuple(ir.Value, lo.Map(ir.Vars, func(item *hir.VarDef, _ int) hir.Expr {
			return item
		}))
	}
}

func (self *CodeGenerator) defGenericFuncDef(ir *hir.GenericFuncInst, f *mir.Function) {
	self.builder.MoveTo(f.NewBlock())
	for i, p := range f.Params() {
		self.values.Set(ir.Define.Params[i], p)
	}
	var maps hashmap.HashMap[*hir.GenericIdentType, mir.Type]
	var i int
	for iter:=ir.Define.GenericParams.Iterator(); iter.Next(); {
		maps.Set(iter.Value().Second, self.codegenType(ir.Params[i]))
		i++
	}
	self.genericIdentMapStack.Push(maps)
	defer self.genericIdentMapStack.Pop()
	block, _ := self.codegenBlock(ir.Define.Body, nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defGenericStructMethodDef(ir *hir.GenericStructMethodInst, f *mir.Function) {
	self.builder.MoveTo(f.NewBlock())
	paramNodes := append([]*hir.Param{ir.Define.SelfParam}, ir.Define.Params...)
	for i, p := range f.Params() {
		self.values.Set(paramNodes[i], p)
	}
	var maps hashmap.HashMap[*hir.GenericIdentType, mir.Type]
	genericParams := ir.GetGenericParams()
	for i, iter:=0, ir.Define.Scope.GenericParams.Iterator(); iter.Next(); i++{
		maps.Set(iter.Value().Second, self.codegenType(genericParams[i]))
	}
	self.genericIdentMapStack.Push(maps)
	defer self.genericIdentMapStack.Pop()
	block, _ := self.codegenBlock(ir.Define.Body, nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defGenericMethodDef(ir *hir.GenericMethodInst, f *mir.Function) {
	self.builder.MoveTo(f.NewBlock())
	paramNodes := append([]*hir.Param{ir.Define.SelfParam}, ir.Define.Params...)
	for i, p := range f.Params() {
		self.values.Set(paramNodes[i], p)
	}
	var maps hashmap.HashMap[*hir.GenericIdentType, mir.Type]
	for i, iter:=0, ir.Define.GenericParams.Iterator(); iter.Next(); i++{
		maps.Set(iter.Value().Second, self.codegenType(ir.Params[i]))
	}
	self.genericIdentMapStack.Push(maps)
	defer self.genericIdentMapStack.Pop()
	block, _ := self.codegenBlock(ir.Define.Body, nil)
	self.builder.BuildUnCondJump(block)
}
