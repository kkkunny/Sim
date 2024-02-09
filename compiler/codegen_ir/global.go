package codegen_ir

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/util"
)

func (self *CodeGenerator) declTypeDef(ir *hir.TypeDef) {
	if stlbasic.Is[*hir.StructType](ir.Target){
		self.types.Set(ir, self.module.NewNamedStructType(""))
	}else{
		// TODO: 支持除结构体之外的类型循环
		self.types.Set(ir, self.codegenType(ir.Target))
	}
}

func (self *CodeGenerator) defTypeDef(ir *hir.TypeDef) {
	if !stlbasic.Is[*hir.StructType](ir.Target){
		return
	}
	td := self.types.Get(ir)
	target := self.codegenType(ir.Target).(mir.StructType)
	td.(mir.StructType).SetElems(target.Elems()...)
}

func (self *CodeGenerator) codegenGlobalDecl(ir hir.Global) {
	switch global := ir.(type) {
	case *hir.FuncDef:
		self.declFuncDef(global)
	case *hir.MethodDef:
		self.declMethodDef(global)
	case *hir.GlobalVarDef:
		self.declGlobalVariable(global)
	case *hir.MultiGlobalVarDef:
		self.declMultiGlobalVariable(global)
	case *hir.TypeDef, *hir.TypeAliasDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(ir *hir.FuncDef) {
	ftObj := self.codegenType(ir.GetType())
	ft := ftObj.(mir.FuncType)
	if ir.VarArg{
		ft.SetVarArg(true)
	}
	f := self.module.NewFunction(ir.ExternName, ft)
	if hir.IsType[*hir.NoReturnType](ir.Ret) {
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.InlineControl.Value(); ok {
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
}

func (self *CodeGenerator) declMethodDef(ir *hir.MethodDef) {
	self.declFuncDef(&ir.FuncDef)
}

func (self *CodeGenerator) declGlobalVariable(ir *hir.GlobalVarDef) {
	t := self.codegenType(ir.Type)
	v := self.module.NewGlobalVariable(ir.ExternName, t, nil)
	self.values.Set(ir, v)
}

func (self *CodeGenerator) declMultiGlobalVariable(ir *hir.MultiGlobalVarDef) {
	for _, varDef := range ir.Vars {
		self.declGlobalVariable(varDef)
	}
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
	case *hir.TypeDef:
		self.defTypeDef(global)
	case *hir.GlobalVarDef:
		if global.Value.IsNone(){
			self.defGlobalVariableDecl(global)
		}else{
			self.defGlobalVariableDef(global)
		}
	case *hir.MultiGlobalVarDef:
		self.defMultiGlobalVariable(global)
	case *hir.TypeAliasDef:
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
	self.defFuncDef(&ir.FuncDef)
}

func (self *CodeGenerator) defFuncDecl(ir *hir.FuncDef) {
	_ = self.values.Get(ir).(*mir.Function)
}

func (self *CodeGenerator) defGlobalVariableDef(ir *hir.GlobalVarDef) {
	gv := self.values.Get(ir).(*mir.GlobalVariable)
	self.builder.MoveTo(self.getInitFunction().Blocks().Front().Value)
	value := self.codegenExpr(ir.Value.MustValue(), true)
	if constValue, ok := value.(mir.Const); ok {
		gv.SetValue(constValue)
	} else {
		self.builder.BuildStore(value, gv)
	}
}

func (self *CodeGenerator) defGlobalVariableDecl(ir *hir.GlobalVarDef) {
	_ = self.values.Get(ir).(*mir.GlobalVariable)
}

func (self *CodeGenerator) defMultiGlobalVariable(ir *hir.MultiGlobalVarDef) {
	if constant, ok := ir.Value.(*hir.Tuple); ok && len(constant.Elems) == len(ir.Vars) {
		for i, varNode := range ir.Vars {
			varNode.Value = util.Some(constant.Elems[i])
			self.defGlobalVariableDef(varNode)
		}
	} else {
		self.builder.MoveTo(self.getInitFunction().Blocks().Front().Value)
		self.codegenUnTuple(ir.Value, lo.Map(ir.Vars, func(item *hir.GlobalVarDef, _ int) hir.Expr {
			return item
		}))
	}
}
