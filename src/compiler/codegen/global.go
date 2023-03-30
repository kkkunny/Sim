package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/mir"
	"github.com/kkkunny/llvm"
	"github.com/kkkunny/stl/util"
)

// 类型别名声明
func (self *CodeGenerator) codegenAliasDecl(ir *mir.Alias) {
	namedStruct := self.ctx.StructCreateNamed(ir.Name)
	self.types[ir] = namedStruct
}

// 类型别名定义
func (self *CodeGenerator) codegenAliasDef(ir *mir.Alias) {
	elemStruct := self.codegenType(ir.Target)
	self.types[ir].StructSetBody(elemStruct.StructElementTypes(), false)
}

// 函数声明
func (self *CodeGenerator) codegenFunctionDecl(ir *mir.Function) llvm.Value {
	fnPtrType := self.codegenType(ir.Type)
	fn := llvm.AddFunction(self.module, ir.Name, fnPtrType.ElementType())
	self.globals[ir] = fn
	fn.SetLinkage(util.Ternary(ir.Extern, llvm.ExternalLinkage, llvm.InternalLinkage))
	if ir.NoReturn {
		fn.AddFunctionAttr(self.ctx.CreateEnumAttribute(llvm.AttributeKindID("noreturn"), 0))
	}
	if ir.GetInline() {
		fn.AddFunctionAttr(self.ctx.CreateEnumAttribute(llvm.AttributeKindID("alwaysinline"), 0))
	}
	if ir.GetNoInline() {
		fn.AddFunctionAttr(self.ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0))
	}
	return fn
}

// 函数定义
func (self *CodeGenerator) codegenFunctionDef(ir *mir.Function) {
	if ir.Blocks.Len() == 0 {
		return
	}

	fn := self.globals[ir]

	block := llvm.AddBasicBlock(fn, "entry")
	self.builder.SetInsertPointAtEnd(block)
	self.vars = make(map[mir.Value]llvm.Value)
	for i, p := range ir.Params {
		param := fn.Param(i)
		alloca := self.builder.CreateAlloca(self.codegenType(p.Type), "")
		alloca.SetAlignment(int(p.Type.Align()))
		store := self.builder.CreateStore(param, alloca)
		store.SetAlignment(int(p.Type.Align()))
		self.vars[p] = alloca
	}

	self.blocks = make(map[*mir.Block]llvm.BasicBlock)
	for cursor := ir.Blocks.Front(); cursor != nil; cursor = cursor.Next() {
		self.blocks[cursor.Value] = llvm.AddBasicBlock(fn, "")
	}
	self.builder.CreateBr(self.blocks[ir.Blocks.Front().Value])
	for cursor := ir.Blocks.Front(); cursor != nil; cursor = cursor.Next() {
		self.codegenBlock(cursor.Value)
	}
}

// 变量声明
func (self *CodeGenerator) codegenVariableDecl(ir *mir.Variable) {
	typ := self.codegenType(ir.Type)
	v := llvm.AddGlobal(self.module, typ, ir.Name)
	self.globals[ir] = v
	v.SetLinkage(util.Ternary(ir.Extern, llvm.ExternalLinkage, llvm.InternalLinkage))
}

// 变量定义
func (self *CodeGenerator) codegenVariableDef(ir *mir.Variable) {
	if ir.Value == nil {
		return
	}
	v := self.globals[ir]
	value := self.codegenConstant(ir.Value)
	v.SetInitializer(value)
}
