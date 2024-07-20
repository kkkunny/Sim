package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/queue"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	hir *hir.Result

	target  *llvm.Target
	ctx     llvm.Context
	module  llvm.Module
	builder llvm.Builder

	values           hashmap.HashMap[hir.Expr, llvm.Value]
	types            hashmap.HashMap[*hir.CustomType, llvm.StructType]
	loops            hashmap.HashMap[hir.Loop, loop]
	strings          hashmap.HashMap[string, llvm.Constant]
	funcCache        hashmap.HashMap[string, llvm.Function]
	lambdaCaptureMap queue.Queue[hashmap.HashMap[hir.Ident, llvm.Value]]
}

func New(target *llvm.Target, ir *hir.Result) *CodeGenerator {
	ctx := llvm.NewContext()
	module := ctx.NewModule("main")
	module.SetTarget(target)
	return &CodeGenerator{
		hir:     ir,
		target:  target,
		ctx:     ctx,
		module:  module,
		builder: ctx.NewBuilder(),
	}
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen() llvm.Module {
	// 值声明
	stliter.Foreach[hir.Global](self.hir.Globals, func(v hir.Global) bool {
		self.codegenGlobalDecl(v)
		return true
	})
	// 值定义
	stliter.Foreach[hir.Global](self.hir.Globals, func(v hir.Global) bool {
		self.codegenGlobalDef(v)
		return true
	})
	// 初始化函数
	for _, b := range self.getInitFunction().Blocks() {
		if !b.IsTerminating() {
			self.builder.MoveToAfter(b)
			self.builder.CreateRet(nil)
		}
	}
	// 主函数
	stliter.Foreach[hir.Global](self.hir.Globals, func(v hir.Global) bool {
		if funcNode, ok := v.(*hir.FuncDef); ok && funcNode.Name == "main" {
			f := self.values.Get(funcNode).(llvm.Function)
			self.builder.MoveToAfter(stlslices.First(self.getMainFunction().Blocks()))
			self.builder.CreateCall("", f.FunctionType(), f)
			self.builder.CreateRet(stlbasic.Ptr[llvm.Value](self.ctx.ConstInteger(self.ctx.IntegerType(8), 0)))
			return false
		}
		return true
	})
	return self.module
}
