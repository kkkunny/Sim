package codegen_ir

import (
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/util"
)

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
	case *hir.TypeDef, *hir.TypeAliasDef, *hir.Trait, *hir.GenericFuncDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(ir *hir.FuncDef) {
	ftObj := self.codegenType(ir.GetType())
	ft := ftObj.(mir.FuncType)
	if ir.VarArg {
		ft.SetVarArg(true)
	}
	f := self.module.NewFunction(ir.ExternName, ft)
	if ir.Ret.EqualTo(hir.NoReturn) {
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.InlineControl.Value(); ok {
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.values.Set(ir, f)
}

func (self *CodeGenerator) declMethodDef(ir *hir.MethodDef) {
	self.declFuncDef(ir.FuncDef)
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
	case *hir.GlobalVarDef:
		if global.Value.IsNone() {
			self.defGlobalVariableDecl(global)
		} else {
			self.defGlobalVariableDef(global)
		}
	case *hir.MultiGlobalVarDef:
		self.defMultiGlobalVariable(global)
	case *hir.TypeAliasDef, *hir.Trait, *hir.TypeDef, *hir.GenericFuncDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) defFuncDef(ir *hir.FuncDef) {
	f := self.values.Get(ir).(*mir.Function)
	self.builder.MoveTo(f.NewBlock())
	for i, pir := range ir.Params {
		self.values.Set(pir, f.Params()[i])
	}
	block, _ := self.codegenBlock(ir.Body.MustValue(), nil)
	self.builder.BuildUnCondJump(block)
}

func (self *CodeGenerator) defMethodDef(ir *hir.MethodDef) {
	self.defFuncDef(ir.FuncDef)
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
		self.codegenUnTuple(ir.Value, stlslices.Map(ir.Vars, func(_ int, item *hir.GlobalVarDef) hir.Expr { return item }))
	}
}

func (self *CodeGenerator) codegenGenericFuncInst(ir *hir.GenericFuncInst) *mir.Function {
	args := stlslices.Map(ir.Args, func(_ int, e hir.Type) mir.Type {
		return self.codegenType(e)
	})
	key := strings.Join(stlslices.Map(args, func(_ int, e mir.Type) string {
		return e.String()
	}), "|")
	if self.genericFuncCache.ContainKey(pair.NewPair(ir.Def, key)) {
		return self.genericFuncCache.Get(pair.NewPair(ir.Def, key))
	}

	preGenericParamMap := self.genericParamMap
	var i int
	self.genericParamMap = stliter.Map[pair.Pair[string, *hir.GenericParam], pair.Pair[*hir.GenericParam, mir.Type], hashmap.HashMap[*hir.GenericParam, mir.Type]](ir.Def.GenericParams, func(p pair.Pair[string, *hir.GenericParam]) pair.Pair[*hir.GenericParam, mir.Type] {
		i++
		return pair.NewPair(p.Second, args[i-1])
	})
	defer func() {
		self.genericParamMap = preGenericParamMap
	}()

	ft := self.codegenType(ir.Def.GetFuncType()).(mir.FuncType)
	f := self.module.NewFunction("", ft)
	if ir.Def.Ret.EqualTo(hir.NoReturn) {
		f.SetAttribute(mir.FunctionAttributeNoReturn)
	}
	if inline, ok := ir.Def.InlineControl.Value(); ok {
		f.SetAttribute(stlbasic.Ternary(inline, mir.FunctionAttributeInline, mir.FunctionAttributeNoInline))
	}
	self.genericFuncCache.Set(pair.NewPair(ir.Def, key), f)

	curBlock := self.builder.Current()
	defer func() {
		self.builder.MoveTo(curBlock)
	}()

	self.builder.MoveTo(f.NewBlock())
	for i, pir := range ir.Def.Params {
		self.values.Set(pir, f.Params()[i])
	}
	block, _ := self.codegenBlock(ir.Def.Body.MustValue(), nil)
	self.builder.BuildUnCondJump(block)
	return f
}
