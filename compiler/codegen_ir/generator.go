package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/bimap"
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/queue"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	llvmUtil "github.com/kkkunny/Sim/compiler/codegen_ir/llvm"
	"github.com/kkkunny/Sim/compiler/hir"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	hir *hir.Result

	builder *llvmUtil.Builder

	values           bimap.BiMap[hir.Expr, llvm.Value]
	types            hashmap.HashMap[*hir.CustomType, llvm.StructType]
	loops            hashmap.HashMap[hir.Loop, loop]
	strings          hashmap.HashMap[string, llvm.Constant]
	funcCache        hashmap.HashMap[string, llvm.Function]
	lambdaCaptureMap queue.Queue[hashmap.HashMap[hir.Ident, llvm.Value]]
}

func New(target llvm.Target, ir *hir.Result) *CodeGenerator {
	return &CodeGenerator{
		hir:              ir,
		builder:          llvmUtil.NewBuilder(target),
		values:           bimap.StdWith[hir.Expr, llvm.Value](),
		types:            hashmap.StdWith[*hir.CustomType, llvm.StructType](),
		loops:            hashmap.StdWith[hir.Loop, loop](),
		strings:          hashmap.StdWith[string, llvm.Constant](),
		funcCache:        hashmap.StdWith[string, llvm.Function](),
		lambdaCaptureMap: queue.New[hashmap.HashMap[hir.Ident, llvm.Value]](),
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
	for _, b := range self.builder.GetInitFunction().Blocks() {
		if !b.IsTerminating() {
			self.builder.MoveToAfter(b)
			self.builder.CreateRet(nil)
		}
	}
	// 主函数
	stliter.Foreach[hir.Global](self.hir.Globals, func(v hir.Global) bool {
		if funcNode, ok := v.(*hir.FuncDef); ok && funcNode.Name == "main" {
			f := self.values.GetValue(funcNode).(llvm.Function)
			self.builder.MoveToAfter(stlslices.First(self.builder.GetMainFunction().Blocks()))
			self.builder.CreateCall("", f.FunctionType(), f)
			self.builder.CreateRet(stlval.Ptr[llvm.Value](self.builder.ConstInteger(self.builder.IntegerType(8), 0)))
			return false
		}
		return true
	})
	return self.builder.Module
}
