package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/analyse"
	"github.com/kkkunny/llvm"
	stlutil "github.com/kkkunny/stl/util"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	ctx      llvm.Context
	module   llvm.Module
	builder  llvm.Builder
	function llvm.Value

	vars  map[analyse.Expr]llvm.Value
	types map[string]llvm.Type

	// loop
	cb, eb llvm.BasicBlock
	// string
	stringPool map[string]llvm.Value
	// cstring
	cstringPool map[string]llvm.Value
}

// NewCodeGenerator 新建代码生成器
func NewCodeGenerator() *CodeGenerator {
	ctx := llvm.NewContext()
	cg := &CodeGenerator{
		ctx:         ctx,
		module:      ctx.NewModule(""),
		builder:     ctx.NewBuilder(),
		vars:        make(map[analyse.Expr]llvm.Value),
		types:       make(map[string]llvm.Type),
		stringPool:  make(map[string]llvm.Value),
		cstringPool: make(map[string]llvm.Value),
	}
	cg.init()
	return cg
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen(mean analyse.ProgramContext) llvm.Module {
	// 声明
	for _, g := range mean.Globals {
		switch global := g.(type) {
		case *analyse.Function:
			ft := self.codegenType(global.GetType()).ElementType()
			f := llvm.AddFunction(self.module, global.ExternName, ft)
			if global.NoReturn {
				f.AddFunctionAttr(self.ctx.CreateEnumAttribute(31, 0))
			}
			if global.Inline != nil {
				f.AddFunctionAttr(self.ctx.CreateEnumAttribute(stlutil.Ternary[uint](*global.Inline, 1, 26), 0))
			}
			self.vars[global] = f
		case *analyse.GlobalVariable:
			vt := self.codegenType(global.GetType())
			self.vars[global] = llvm.AddGlobal(self.module, vt, global.ExternName)
		default:
			panic("")
		}
	}
	// 定义
	for _, g := range mean.Globals {
		switch global := g.(type) {
		case *analyse.Function:
			if global.Body != nil {
				f := self.vars[global]
				self.function = f
				entry := llvm.AddBasicBlock(f, "")
				self.builder.SetInsertPointAtEnd(entry)

				for i, p := range global.Params {
					param := self.builder.CreateAlloca(self.codegenType(p.GetType()), "")
					self.builder.CreateStore(f.Param(i), param)
					self.vars[p] = param
				}

				self.codegenBlock(*global.Body)
			}
		case *analyse.GlobalVariable:
			if global.Value != nil {
				self.vars[global].SetInitializer(self.codegenConstantExpr(global.Value))
			}
		default:
			panic("")
		}
	}
	return self.module
}
