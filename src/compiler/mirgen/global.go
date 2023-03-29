package mirgen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/mir"
)

// 函数声明
func (self *MirGenerator) genFunctionDecl(ir *hir.Function) {
	ft := self.genType(ir.Type())
	var name string
	if ir.Name != "" {
		name = ir.Name
	}
	f := self.pkg.NewFunction(ft, name)
	if ir.Name != "" {
		f.Extern = true
	}
	if ir.NoReturn {
		f.NoReturn = true
	}
	if ir.MustInline {
		f.SetInline(true)
	}
	if ir.MustNoInline {
		f.SetNoInline(true)
	}
	if ir.Init {
		self.inits = append(self.inits, f)
	}
	if ir.Fini {
		self.finis = append(self.finis, f)
	}
	self.vars[ir] = f
}

// 方法声明
func (self *MirGenerator) genMethodDecl(ir *hir.Method) {
	ft := self.genType(ir.FunctionType())
	f := self.pkg.NewFunction(ft, "")
	if ir.NoReturn {
		f.NoReturn = true
	}
	if ir.MustInline {
		f.SetInline(true)
	}
	if ir.MustNoInline {
		f.SetNoInline(true)
	}
	self.vars[ir] = f
}

// 全局变量声明
func (self *MirGenerator) genGlobalValueDecl(ir *hir.GlobalValue) {
	var name string
	if ir.Name != "" {
		name = ir.Name
	}
	v := self.pkg.NewVariable(self.genType(ir.Type()), name, nil)
	if ir.Value != nil {
		v.Value = mir.NewEmpty(v.Type)
	}
	if ir.Name != "" {
		v.Extern = true
	}
	self.vars[ir] = v
}

// 函数定义
func (self *MirGenerator) genFunctionDef(ir *hir.Function) {
	if ir.Body == nil {
		return
	}
	f := self.vars[ir].(*mir.Function)
	self.block = f.NewBlock()
	for i, p := range ir.Params {
		self.vars[p] = f.Params[i]
	}
	self.genBlock(*ir.Body)
}

// 方法定义
func (self *MirGenerator) genMethodDef(ir *hir.Method) {
	f := self.vars[ir].(*mir.Function)
	self.block = f.NewBlock()
	for i, p := range ir.Params {
		self.vars[p] = f.Params[i]
	}
	self.genBlock(*ir.Body)
}

// 全局变量定义
func (self *MirGenerator) genGlobalValueDef(ir *hir.GlobalValue) {
	if ir.Value == nil && ir.Name != "" {
		return
	}
	v := self.vars[ir].(*mir.Variable)
	if ir.Value == nil {
		v.Value = mir.NewEmpty(v.Type)
	}
	if self.globalDefFunc == nil {
		self.globalDefFunc = self.pkg.NewFunction(
			mir.NewTypeFunc(false, mir.NewTypeVoid()),
			"",
		)
		self.globalDefFunc.SetInit(true)
		self.globalLastBlock = self.globalDefFunc.NewBlock()
	}
	self.block = self.globalLastBlock
	value := self.genExpr(ir.Value, true)
	self.block.NewStore(value, self.vars[ir])
	self.globalLastBlock = self.block
}
