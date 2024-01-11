package codegen_ir

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/pair"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/runtime/types"
	"github.com/kkkunny/Sim/util"
)

func (self *CodeGenerator) declStructDef(ir *hir.StructDef) {
	self.structs.Set(ir.String(), pair.NewPair[mir.StructType, *types.StructType](self.module.NewNamedStructType(""), types.NewStructType(ir.Pkg.String(), ir.Name, nil, nil)))
}

func (self *CodeGenerator) defStructDef(ir *hir.StructDef) {
	stPair := self.structs.Get(ir.String())
	fields, fieldRts := make([]mir.Type, ir.Fields.Length()), make([]types.Field, ir.Fields.Length())
	for i, iter := 0, ir.Fields.Values().Iterator(); iter.Next(); i++ {
		f, fRt := self.codegenType(iter.Value().Type)
		fields[i], fieldRts[i] = f, types.NewField(fRt, ir.Name)
	}
	stPair.First.SetElems(fields...)
	stPair.Second.Fields = fieldRts
	methodRts := make([]types.Method, 0, ir.Methods.Length())
	for iter := ir.Methods.Values().Iterator(); iter.Next(); {
		switch method := iter.Value().(type) {
		case *hir.MethodDef:
			_, ft := self.codegenFuncType(method.GetFuncType())
			methodRts = append(methodRts, types.NewMethod(ft, method.Name))
		}
	}
	stPair.Second.Methods = methodRts
}

func (self *CodeGenerator) declGenericStructDef(ir *hir.GenericStructInst) (mir.StructType, *types.StructType) {
	return self.module.NewNamedStructType(""), types.NewStructType(ir.Define.Pkg.String(), ir.Define.GetName(), nil, nil)
}

func (self *CodeGenerator) defGenericStructDef(ir *hir.GenericStructInst, st mir.StructType, stRt *types.StructType) {
	defer self.mapGenericParams2GenericArgs(ir)()

	stIr := ir.StructType()
	fields := make([]mir.Type, stIr.Fields.Length())
	fieldRts := make([]types.Field, stIr.Fields.Length())
	for i, iter := 0, stIr.Fields.Iterator(); iter.Next(); i++ {
		f, fRt := self.codegenType(iter.Value().Second.Type)
		fields[i], fieldRts[i] = f, types.NewField(fRt, iter.Value().First)
	}
	st.SetElems(fields...)
	stRt.Fields = fieldRts
	methodRts := make([]types.Method, 0, stIr.Methods.Length())
	for iter := stIr.Methods.Values().Iterator(); iter.Next(); {
		switch method := iter.Value().(type) {
		case *hir.GenericStructMethodDef:
			if method.GenericParams.Empty(){
				_, ft := self.codegenFuncType(method.GetFuncType())
				methodRts = append(methodRts, types.NewMethod(ft, method.Name))
			}
		}
	}
	stRt.Methods = methodRts
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
	case *hir.StructDef, *hir.TypeAliasDef, *hir.GenericFuncDef, *hir.GenericStructDef, *hir.GenericStructMethodDef, *hir.GenericMethodDef, *hir.TraitDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(ir *hir.FuncDef) {
	ftObj, _ := self.codegenType(ir.GetType())
	ft := ftObj.(mir.FuncType)
	if ir.VarArg{
		ft.SetVarArg(true)
	}
	f := self.module.NewFunction(ir.ExternName, ft)
	if ir.NoReturn {
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
	t, _ := self.codegenType(ir.Type)
	v := self.module.NewGlobalVariable(ir.ExternName, t, nil)
	self.values.Set(ir, v)
}

func (self *CodeGenerator) declMultiGlobalVariable(ir *hir.MultiGlobalVarDef) {
	for _, varDef := range ir.Vars {
		self.declGlobalVariable(varDef)
	}
}

func (self *CodeGenerator) declGenericFuncDef(ir *hir.GenericFuncInst) *mir.Function {
	ftObj, _ := self.codegenType(ir.GetType())
	ft := ftObj.(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Define.NoReturn {
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.Define.InlineControl.Value(); ok {
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
	return f
}

func (self *CodeGenerator) declGenericStructMethodDef(ir *hir.GenericStructMethodInst) *mir.Function {
	ftObj, _ := self.codegenType(ir.GetFuncType())
	ft := ftObj.(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Define.NoReturn {
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.Define.InlineControl.Value(); ok {
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
	return f
}

func (self *CodeGenerator) declGenericMethodDef(ir *hir.GenericMethodInst) *mir.Function {
	ftObj, _ := self.codegenType(ir.GetFuncType())
	ft := ftObj.(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Define.NoReturn {
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.Define.InlineControl.Value(); ok {
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
	return f
}

func (self *CodeGenerator) declGenericStructGenericMethodDef(ir *hir.GenericStructGenericMethodInst) *mir.Function {
	ftObj, _ := self.codegenType(ir.GetFuncType())
	ft := ftObj.(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Define.NoReturn {
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.Define.InlineControl.Value(); ok {
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
	case *hir.GlobalVarDef:
		if global.Value.IsNone(){
			self.defGlobalVariableDecl(global)
		}else{
			self.defGlobalVariableDef(global)
		}
	case *hir.MultiGlobalVarDef:
		self.defMultiGlobalVariable(global)
	case *hir.TypeAliasDef, *hir.GenericFuncDef, *hir.GenericStructDef, *hir.GenericStructMethodDef, *hir.GenericMethodDef, *hir.TraitDef:
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

func (self *CodeGenerator) defGenericFuncDef(ir *hir.GenericFuncInst, f *mir.Function) {
	self.builder.MoveTo(f.NewBlock())
	for i, p := range f.Params() {
		self.values.Set(ir.Define.Params[i], p)
	}

	defer self.mapGenericParams2GenericArgs(ir)()

	block, _ := self.codegenBlock(ir.Define.Body, nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defGenericStructMethodDef(ir *hir.GenericStructMethodInst, f *mir.Function) {
	self.builder.MoveTo(f.NewBlock())
	for i, p := range f.Params() {
		self.values.Set(ir.Define.Params[i], p)
	}

	defer self.mapGenericParams2GenericArgs(ir)()

	block, _ := self.codegenBlock(ir.Define.Body, nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defGenericMethodDef(ir *hir.GenericMethodInst, f *mir.Function) {
	self.builder.MoveTo(f.NewBlock())
	for i, p := range f.Params() {
		self.values.Set(ir.Define.Params[i], p)
	}

	defer self.mapGenericParams2GenericArgs(ir)()

	block, _ := self.codegenBlock(ir.Define.Body, nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defGenericStructGenericMethodDef(ir *hir.GenericStructGenericMethodInst, f *mir.Function) {
	self.builder.MoveTo(f.NewBlock())
	for i, p := range f.Params() {
		self.values.Set(ir.Define.Params[i], p)
	}

	defer self.mapGenericParams2GenericArgs(ir)()

	block, _ := self.codegenBlock(ir.Define.Body, nil)
	self.builder.BuildUnCondJump(block)
}
