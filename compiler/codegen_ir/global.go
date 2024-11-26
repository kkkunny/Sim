package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (self *CodeGenerator) codegenGlobalVarDecl(pkg *hir.Package) {
	for iter := pkg.Globals().Iterator(); iter.Next(); {
		switch ir := iter.Value().(type) {
		case *global.FuncDef:
			self.declFuncDef(ir)
		case *global.OriginMethodDef:
			self.declMethodDef(ir)
		case *global.VarDef:
			self.declGlobalVarDef(ir)
		}
	}
}

func (self *CodeGenerator) declFuncDef(ir *global.FuncDef) {
	if len(ir.GenericParams()) > 0 {
		return
	}
	f := self.declFunc(self.getIdentName(ir), ir.CallableType().(types.FuncType), ir.Attrs()...)
	self.values.Set(ir, f)
}

func (self *CodeGenerator) declMethodDef(ir *global.OriginMethodDef) {
	if len(ir.From().GenericParams()) > 0 || len(ir.GenericParams()) > 0 {
		return
	}
	f := self.declFunc(self.getIdentName(ir), ir.CallableType().(types.FuncType), ir.Attrs()...)
	self.values.Set(ir, f)
}

func (self *CodeGenerator) declGlobalVarDef(ir *global.VarDef) {
	var link bool
	for _, attr := range ir.Attrs() {
		switch attr.(type) {
		case *global.VarAttrLinkName:
			link = true
		}
	}

	name := self.getIdentName(ir)
	t := self.codegenType(ir.Type())
	v := self.builder.NewGlobal(name, t, nil)
	v.SetLinkage(stlval.Ternary(link, llvm.ExternalLinkage, llvm.LinkOnceODRAutoHideLinkage))
	if !link {
		v.SetInitializer(self.builder.ConstZero(t))
	}
	self.values.Set(ir, v)
}

func (self *CodeGenerator) codegenGlobalVarDef(pkg *hir.Package) {
	for iter := pkg.Globals().Iterator(); iter.Next(); {
		switch ir := iter.Value().(type) {
		case *global.FuncDef:
			self.defFuncDef(ir)
		case *global.OriginMethodDef:
			self.defMethodDef(ir)
		case *global.VarDef:
			self.defGlobalVarDef(ir)
		}
	}
}

func (self *CodeGenerator) defFuncDef(ir *global.FuncDef) {
	if len(ir.GenericParams()) > 0 {
		return
	}
	body, ok := ir.Body()
	if !ok {
		return
	}
	f := self.values.Get(ir).(llvm.Function)
	self.defFunc(f, ir.Params(), body)
}

func (self *CodeGenerator) defMethodDef(ir *global.OriginMethodDef) {
	if len(ir.From().GenericParams()) > 0 || len(ir.GenericParams()) > 0 {
		return
	}
	body, ok := ir.Body()
	if !ok {
		return
	}
	f := self.values.Get(ir).(llvm.Function)
	self.defFunc(f, ir.Params(), body)
}

func (self *CodeGenerator) defGlobalVarDef(ir *global.VarDef) {
	vIr, ok := ir.Value()
	if !ok {
		return
	}

	gv := self.values.Get(ir).(llvm.GlobalValue)
	self.builder.MoveToAfter(stlslices.First(self.builder.GetInitFunction().Blocks()))
	value := self.codegenValue(vIr, true)
	if constValue, ok := value.(llvm.Constant); ok {
		gv.SetInitializer(constValue)
	} else {
		self.builder.CreateStore(value, gv)
	}
}
