package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/llvm"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	ctx      llvm.Context
	module   llvm.Module
	builder  llvm.Builder
	function llvm.Value

	vars  map[hir.Ident]llvm.Value
	types map[string]llvm.Type

	// loop
	cb, eb llvm.BasicBlock
	// string
	stringPool map[string]llvm.Value
	// cstring
	cstringPool map[string]llvm.Value
	// global
	globalDefFunc   llvm.Value
	globalLastBlock llvm.BasicBlock
	// init fini
	inits, finis []llvm.Value
}

// NewCodeGenerator 新建代码生成器
func NewCodeGenerator() *CodeGenerator {
	ctx := llvm.NewContext()
	cg := &CodeGenerator{
		ctx:         ctx,
		module:      ctx.NewModule(""),
		builder:     ctx.NewBuilder(),
		vars:        make(map[hir.Ident]llvm.Value),
		types:       make(map[string]llvm.Type),
		stringPool:  make(map[string]llvm.Value),
		cstringPool: make(map[string]llvm.Value),
	}
	cg.init()
	return cg
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen(pkg hir.Package) llvm.Module {
	// 声明
	for iter := pkg.Globals.Iterator(); iter.HasValue(); iter.Next() {
		switch global := iter.Value().(type) {
		case *hir.Typedef:
			// TODO:
		case *hir.Function:
			ft := self.codegenType(global.Type()).ElementType()
			f := llvm.AddFunction(self.module, global.Name, ft)
			if global.Name == "" {
				f.SetLinkage(llvm.InternalLinkage)
			} else {
				f.SetLinkage(llvm.ExternalLinkage)
			}
			if global.NoReturn {
				f.AddFunctionAttr(self.ctx.CreateEnumAttribute(31, 0))
			}
			if global.MustInline {
				f.AddFunctionAttr(self.ctx.CreateEnumAttribute(1, 0))
			}
			if global.MustNoInline {
				f.AddFunctionAttr(self.ctx.CreateEnumAttribute(26, 0))
			}
			if global.Init {
				self.inits = append(self.inits, f)
			}
			if global.Fini {
				self.finis = append(self.finis, f)
			}
			self.vars[global] = f
		case *hir.Method:
			ft := self.codegenType(global.FunctionType()).ElementType()
			f := llvm.AddFunction(self.module, "", ft)
			f.SetLinkage(llvm.InternalLinkage)
			if global.NoReturn {
				f.AddFunctionAttr(self.ctx.CreateEnumAttribute(31, 0))
			}
			if global.MustInline {
				f.AddFunctionAttr(self.ctx.CreateEnumAttribute(1, 0))
			}
			if global.MustNoInline {
				f.AddFunctionAttr(self.ctx.CreateEnumAttribute(26, 0))
			}
			self.vars[global] = f
		case *hir.GlobalValue:
			vt := self.codegenType(global.Type())
			v := llvm.AddGlobal(self.module, vt, global.Name)
			if global.Name == "" {
				v.SetLinkage(llvm.InternalLinkage)
			} else if global.Value == nil {
				// TODO: 全局变量声明
				v.SetLinkage(llvm.ExternalLinkage)
			}
			if global.Value != nil {
				v.SetInitializer(llvm.ConstZero(vt))
			}
			self.vars[global] = v
		default:
			panic("unreachable")
		}
	}
	// 定义
	for iter := pkg.Globals.Iterator(); iter.HasValue(); iter.Next() {
		switch global := iter.Value().(type) {
		case *hir.Typedef:
		case *hir.Function:
			if global.Body == nil {
				continue
			}
			f := self.vars[global]
			self.function = f
			entry := llvm.AddBasicBlock(f, "")
			self.builder.SetInsertPointAtEnd(entry)

			for i, p := range global.Params {
				param := self.builder.CreateAlloca(self.codegenType(p.Type()), "")
				self.builder.CreateStore(f.Param(i), param)
				self.vars[p] = param
			}

			self.codegenBlock(*global.Body)
		case *hir.Method:
			f := self.vars[global]
			self.function = f
			entry := llvm.AddBasicBlock(f, "")
			self.builder.SetInsertPointAtEnd(entry)

			for i, p := range global.Params {
				param := self.builder.CreateAlloca(self.codegenType(p.Type()), "")
				self.builder.CreateStore(f.Param(i), param)
				self.vars[p] = param
			}

			self.codegenBlock(*global.Body)
		case *hir.GlobalValue:
			if global.Value == nil {
				continue
			}
			if self.globalDefFunc.IsNil() {
				self.globalDefFunc = llvm.AddFunction(
					self.module,
					"",
					llvm.FunctionType(self.ctx.VoidType(), nil, false),
				)
				self.globalDefFunc.SetLinkage(llvm.InternalLinkage)
				self.inits = append(self.inits, self.globalDefFunc)
				self.globalLastBlock = llvm.AddBasicBlock(self.globalDefFunc, "entry")
			}
			self.builder.SetInsertPointAtEnd(self.globalLastBlock)
			value := self.codegenExpr(global.Value, true)
			if value.IsConstant() {
				self.vars[global].SetInitializer(value)
			} else {
				self.builder.CreateStore(value, self.vars[global])
			}
			self.globalLastBlock = self.builder.GetInsertBlock()
		default:
			panic("unreachable")
		}
	}
	if !self.globalLastBlock.IsNil() {
		self.builder.SetInsertPointAtEnd(self.globalLastBlock)
		self.builder.CreateRetVoid()
	}
	// init
	structType := self.ctx.StructType(
		[]llvm.Type{
			self.ctx.Int32Type(),
			llvm.PointerType(llvm.FunctionType(self.ctx.VoidType(), nil, false), 0),
			llvm.PointerType(self.ctx.Int8Type(), 0),
		},
		false,
	)
	if len(self.inits) > 0 {
		init := llvm.AddGlobal(self.module, llvm.ArrayType(structType, len(self.inits)), "llvm.global_ctors")
		init.SetLinkage(llvm.AppendingLinkage)
		values := make([]llvm.Value, len(self.inits))
		for i, f := range self.inits {
			values[i] = llvm.ConstStruct(
				[]llvm.Value{
					llvm.ConstInt(structType.StructElementTypes()[0], 65535, true),
					f,
					llvm.ConstZero(structType.StructElementTypes()[2]),
				},
				false,
			)
		}
		init.SetInitializer(llvm.ConstArray(structType, values))
	}
	// fini
	if len(self.finis) > 0 {
		fini := llvm.AddGlobal(self.module, llvm.ArrayType(structType, len(self.finis)), "llvm.global_dtors")
		fini.SetLinkage(llvm.AppendingLinkage)
		values := make([]llvm.Value, len(self.finis))
		for i, f := range self.finis {
			values[len(self.finis)-i-1] = llvm.ConstStruct(
				[]llvm.Value{
					llvm.ConstInt(structType.StructElementTypes()[0], 65535, true),
					f,
					llvm.ConstZero(structType.StructElementTypes()[2]),
				},
				false,
			)
		}
		fini.SetInitializer(llvm.ConstArray(structType, values))
	}
	return self.module
}
