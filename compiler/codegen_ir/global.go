package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/oldhir"
)

func (self *CodeGenerator) codegenGlobalDecl(ir oldhir.Global) {
	switch global := ir.(type) {
	case *oldhir.FuncDef:
		self.declFuncDef(global)
	case *oldhir.MethodDef:
		self.declMethodDef(global)
	case *oldhir.GlobalVarDef:
		self.declGlobalVariable(global)
	case *oldhir.MultiGlobalVarDef:
		self.declMultiGlobalVariable(global)
	case *oldhir.TypeDef, *oldhir.TypeAliasDef, *oldhir.Trait:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) declFuncDef(ir *oldhir.FuncDef) {
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
	if ir.Ret.EqualTo(oldhir.NoReturn) {
		f.AddAttribute(llvm.FuncAttributeNoReturn)
	}
	if inline, ok := ir.InlineControl.Value(); ok {
		f.AddAttribute(stlval.Ternary(inline, llvm.FuncAttributeAlwaysInline, llvm.FuncAttributeNoInline))
	}
	self.values.Set(ir, f)
}

func (self *CodeGenerator) declMethodDef(ir *oldhir.MethodDef) {
	self.declFuncDef(&ir.FuncDef)
}

func (self *CodeGenerator) declGlobalVariable(ir *oldhir.GlobalVarDef) {
	t := self.codegenType(ir.Type)
	v := self.builder.NewGlobal(ir.ExternName, t, nil)
	if ir.ExternName == "" {
		v.SetLinkage(llvm.PrivateLinkage)
	} else {
		v.SetLinkage(llvm.ExternalLinkage)
	}
	self.values.Set(ir, v)
}

func (self *CodeGenerator) declMultiGlobalVariable(ir *oldhir.MultiGlobalVarDef) {
	for _, varDef := range ir.Vars {
		self.declGlobalVariable(varDef)
	}
}

func (self *CodeGenerator) codegenGlobalDef(ir oldhir.Global) {
	switch global := ir.(type) {
	case *oldhir.FuncDef:
		if global.Body.IsNone() {
			self.defFuncDecl(global)
		} else {
			self.defFuncDef(global)
		}
	case *oldhir.MethodDef:
		self.defMethodDef(global)
	case *oldhir.GlobalVarDef:
		if global.Value.IsNone() {
			self.defGlobalVariableDecl(global)
		} else {
			self.defGlobalVariableDef(global)
		}
	case *oldhir.MultiGlobalVarDef:
		self.defMultiGlobalVariable(global)
	case *oldhir.TypeAliasDef, *oldhir.Trait, *oldhir.TypeDef:
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) defFuncDef(ir *oldhir.FuncDef) {
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

func (self *CodeGenerator) defMethodDef(ir *oldhir.MethodDef) {
	self.defFuncDef(&ir.FuncDef)
}

func (self *CodeGenerator) defFuncDecl(ir *oldhir.FuncDef) {
	_ = self.values.GetValue(ir).(llvm.Function)
}

func (self *CodeGenerator) defGlobalVariableDef(ir *oldhir.GlobalVarDef) {
	gv := self.values.GetValue(ir).(llvm.GlobalValue)
	self.builder.MoveToAfter(stlslices.First(self.builder.GetInitFunction().Blocks()))
	value := self.codegenExpr(ir.Value.MustValue(), true)
	if constValue, ok := value.(llvm.Constant); ok {
		gv.SetInitializer(constValue)
	} else {
		self.builder.CreateStore(value, gv)
	}
}

func (self *CodeGenerator) defGlobalVariableDecl(ir *oldhir.GlobalVarDef) {
	_ = self.values.GetValue(ir).(llvm.GlobalValue)
}

func (self *CodeGenerator) defMultiGlobalVariable(ir *oldhir.MultiGlobalVarDef) {
	if constant, ok := ir.Value.(*oldhir.Tuple); ok && len(constant.Elems) == len(ir.Vars) {
		for i, varNode := range ir.Vars {
			varNode.Value = optional.Some(constant.Elems[i])
			self.defGlobalVariableDef(varNode)
		}
	} else {
		self.builder.MoveToAfter(stlslices.First(self.builder.GetInitFunction().Blocks()))
		self.codegenUnTuple(ir.Value, stlslices.Map(ir.Vars, func(_ int, item *oldhir.GlobalVarDef) oldhir.Expr { return item }))
	}
}
