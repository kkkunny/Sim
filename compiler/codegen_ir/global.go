package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
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
	case *hir.TypeDef, *hir.TypeAliasDef, *hir.Trait:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(ir *hir.FuncDef) {
	ft := self.codegenFuncType(ir.GetFuncType())
	if ir.VarArg {
		ft = self.builder.FunctionType(true, ft.ReturnType(), ft.Params()...)
	}
	f := self.builder.NewFunction(ir.ExternName, ft)
	if ir.ExternName == "" {
		f.SetLinkage(llvm.PrivateLinkage)
	} else {
		f.SetLinkage(llvm.ExternalLinkage)
	}
	if ir.Ret.EqualTo(hir.NoReturn) {
		f.AddAttribute(llvm.FuncAttributeNoReturn)
	}
	if inline, ok := ir.InlineControl.Value(); ok {
		f.AddAttribute(stlbasic.Ternary(inline, llvm.FuncAttributeAlwaysInline, llvm.FuncAttributeNoInline))
	}
	self.values.Set(ir, f)
}

func (self *CodeGenerator) declMethodDef(ir *hir.MethodDef) {
	self.declFuncDef(&ir.FuncDef)
}

func (self *CodeGenerator) declGlobalVariable(ir *hir.GlobalVarDef) {
	t := self.codegenType(ir.Type)
	v := self.builder.NewGlobal(ir.ExternName, t, nil)
	if ir.ExternName == "" {
		v.SetLinkage(llvm.PrivateLinkage)
	} else {
		v.SetLinkage(llvm.ExternalLinkage)
	}
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
	case *hir.TypeAliasDef, *hir.Trait, *hir.TypeDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) defFuncDef(ir *hir.FuncDef) {
	f := self.values.GetValue(ir).(llvm.Function)
	self.builder.MoveToAfter(f.NewBlock(""))
	for i, pir := range ir.Params {
		p := self.builder.CreateAlloca("", self.codegenType(pir.Type))
		self.builder.CreateStore(f.Params()[i], p)
		self.values.Set(pir, p)
	}
	block, _ := self.codegenBlock(ir.Body.MustValue(), nil)
	self.builder.CreateBr(block)
}

func (self *CodeGenerator) defMethodDef(ir *hir.MethodDef) {
	self.defFuncDef(&ir.FuncDef)
}

func (self *CodeGenerator) defFuncDecl(ir *hir.FuncDef) {
	_ = self.values.GetValue(ir).(llvm.Function)
}

func (self *CodeGenerator) defGlobalVariableDef(ir *hir.GlobalVarDef) {
	gv := self.values.GetValue(ir).(llvm.GlobalValue)
	self.builder.MoveToAfter(stlslices.First(self.builder.GetInitFunction().Blocks()))
	value := self.codegenExpr(ir.Value.MustValue(), true)
	if constValue, ok := value.(llvm.Constant); ok {
		gv.SetInitializer(constValue)
	} else {
		self.builder.CreateStore(value, gv)
	}
}

func (self *CodeGenerator) defGlobalVariableDecl(ir *hir.GlobalVarDef) {
	_ = self.values.GetValue(ir).(llvm.GlobalValue)
}

func (self *CodeGenerator) defMultiGlobalVariable(ir *hir.MultiGlobalVarDef) {
	if constant, ok := ir.Value.(*hir.Tuple); ok && len(constant.Elems) == len(ir.Vars) {
		for i, varNode := range ir.Vars {
			varNode.Value = optional.Some(constant.Elems[i])
			self.defGlobalVariableDef(varNode)
		}
	} else {
		self.builder.MoveToAfter(stlslices.First(self.builder.GetInitFunction().Blocks()))
		self.codegenUnTuple(ir.Value, stlslices.Map(ir.Vars, func(_ int, item *hir.GlobalVarDef) hir.Expr { return item }))
	}
}
