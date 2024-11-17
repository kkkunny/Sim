package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (self *CodeGenerator) codegenTypeDefDecl(pkg *hir.Package) {
	for iter := pkg.Globals().Iterator(); iter.Next(); {
		switch ir := iter.Value().(type) {
		case global.CustomTypeDef:
			self.declCustomType(ir)
		}
	}
}

func (self *CodeGenerator) declCustomType(ir global.CustomTypeDef) {
	if !self.typeIsStruct(ir.Target()) {
		return
	}

	t := self.builder.NamedStructType("", false)
	self.types.Set(ir, t)
}

func (self *CodeGenerator) codegenTypeDefDef(pkg *hir.Package) {
	for iter := pkg.Globals().Iterator(); iter.Next(); {
		switch ir := iter.Value().(type) {
		case global.CustomTypeDef:
			self.defCustomType(ir)
		}
	}
}

func (self *CodeGenerator) defCustomType(ir global.CustomTypeDef) {
	tObj := self.types.Get(ir)
	if tObj == nil {
		return
	}
	t := tObj.(llvm.StructType)

	target := self.codegenType(ir.Target())
	if tt, ok := target.(llvm.StructType); ok {
		t.SetElems(false, tt.Elems()...)
	} else {
		t.SetElems(true, target)
	}
}

func (self *CodeGenerator) codegenGlobalVarDecl(pkg *hir.Package) {
	for iter := pkg.Globals().Iterator(); iter.Next(); {
		switch ir := iter.Value().(type) {
		case *global.FuncDef:
			self.declFuncDef(ir)
		case *global.MethodDef:
			self.declMethodDef(ir)
		case *global.VarDef:
			self.declGlobalVarDef(ir)
		}
	}
}

func (self *CodeGenerator) declFuncDef(ir *global.FuncDef) {
	var inline *bool
	var vararg, link bool
	for _, attr := range ir.Attrs() {
		switch attr := attr.(type) {
		case *global.FuncAttrLinkName:
			link = true
		case *global.FuncAttrInline:
			inline = stlval.Ptr(attr.Inline())
		case *global.FuncAttrVararg:
			vararg = true
		}
	}

	name := self.getIdentName(ir)
	ftIr := ir.CallableType().(types.FuncType)
	ft := self.codegenFuncType(ftIr)
	if vararg {
		ft = self.builder.FunctionType(true, ft.ReturnType(), ft.Params()...)
	}

	f := self.builder.NewFunction(name, ft)
	f.SetLinkage(stlval.Ternary(link, llvm.ExternalLinkage, llvm.LinkOnceODRAutoHideLinkage))
	if !link {
		f.SetDSOLocal(true)
	}
	if types.Is[types.NoReturnType](ftIr.Ret(), true) {
		f.AddAttribute(llvm.FuncAttributeNoReturn)
	}
	if inline != nil {
		f.AddAttribute(stlval.Ternary(*inline, llvm.FuncAttributeAlwaysInline, llvm.FuncAttributeNoInline))
	}
	self.values.Set(ir, f)
}

func (self *CodeGenerator) declMethodDef(ir *global.MethodDef) {
	var inline *bool
	var vararg, link bool
	for _, attr := range ir.Attrs() {
		switch attr := attr.(type) {
		case *global.FuncAttrLinkName:
			link = true
		case *global.FuncAttrInline:
			inline = stlval.Ptr(attr.Inline())
		case *global.FuncAttrVararg:
			vararg = true
		}
	}

	name := self.getIdentName(ir)
	ftIr := ir.CallableType().(types.FuncType)
	ft := self.codegenFuncType(ftIr)
	if vararg {
		ft = self.builder.FunctionType(true, ft.ReturnType(), ft.Params()...)
	}

	f := self.builder.NewFunction(name, ft)
	f.SetLinkage(stlval.Ternary(link, llvm.ExternalLinkage, llvm.LinkOnceODRAutoHideLinkage))
	if !link {
		f.SetDSOLocal(true)
	}
	if types.Is[types.NoReturnType](ftIr.Ret(), true) {
		f.AddAttribute(llvm.FuncAttributeNoReturn)
	}
	if inline != nil {
		f.AddAttribute(stlval.Ternary(*inline, llvm.FuncAttributeAlwaysInline, llvm.FuncAttributeNoInline))
	}
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
		case *global.MethodDef:
			self.defMethodDef(ir)
		case *global.VarDef:
			self.defGlobalVarDef(ir)
		}
	}
}

func (self *CodeGenerator) defFuncDef(ir *global.FuncDef) {
	body, ok := ir.Body()
	if !ok {
		return
	}

	f := self.values.Get(ir).(llvm.Function)
	self.builder.MoveToAfter(f.NewBlock(""))
	for i, pIr := range ir.Params() {
		p := self.builder.CreateAlloca("", self.codegenType(pIr.Type()))
		self.builder.CreateStore(f.Params()[i], p)
		self.values.Set(pIr, p)
	}

	block, _ := self.codegenBlock(body, nil)
	self.builder.CreateBr(block)
}

func (self *CodeGenerator) defMethodDef(ir *global.MethodDef) {
	body, ok := ir.Body()
	if !ok {
		return
	}

	f := self.values.Get(ir).(llvm.Function)
	self.builder.MoveToAfter(f.NewBlock(""))
	for i, pIr := range ir.Params() {
		p := self.builder.CreateAlloca("", self.codegenType(pIr.Type()))
		self.builder.CreateStore(f.Params()[i], p)
		self.values.Set(pIr, p)
	}

	block, _ := self.codegenBlock(body, nil)
	self.builder.CreateBr(block)
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
