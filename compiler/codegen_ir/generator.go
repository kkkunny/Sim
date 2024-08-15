package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/bimap"
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/queue"
	stlslices "github.com/kkkunny/stl/container/slices"

	llvmUtil "github.com/kkkunny/Sim/compiler/codegen_ir/llvm"
	"github.com/kkkunny/Sim/compiler/hir"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	hir *oldhir.Result

	builder *llvmUtil.Builder

	values           bimap.BiMap[oldhir.Expr, llvm.Value]
	types            hashmap.HashMap[*oldhir.CustomType, llvm.StructType]
	loops            hashmap.HashMap[oldhir.Loop, loop]
	strings          hashmap.HashMap[string, llvm.Constant]
	funcCache        hashmap.HashMap[string, llvm.Function]
	lambdaCaptureMap queue.Queue[hashmap.HashMap[oldhir.Ident, llvm.Value]]
}

func New(target llvm.Target, ir *oldhir.Result) *CodeGenerator {
	return &CodeGenerator{
		hir:     ir,
		builder: llvmUtil.NewBuilder(target),
	}
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen() llvm.Module {
	// 值声明
	stliter.Foreach[oldhir.Global](self.hir.Globals, func(v oldhir.Global) bool {
		self.codegenGlobalDecl(v)
		return true
	})
	// 值定义
	stliter.Foreach[oldhir.Global](self.hir.Globals, func(v oldhir.Global) bool {
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
	stliter.Foreach[oldhir.Global](self.hir.Globals, func(v oldhir.Global) bool {
		if funcNode, ok := v.(*oldhir.FuncDef); ok && funcNode.Name == "main" {
			f := self.values.GetValue(funcNode).(llvm.Function)
			self.builder.MoveToAfter(stlslices.First(self.builder.GetMainFunction().Blocks()))
			self.builder.CreateCall("", f.FunctionType(), f)
			self.builder.CreateRet(stlbasic.Ptr[llvm.Value](self.builder.ConstInteger(self.builder.IntegerType(8), 0)))
			return false
		}
		return true
	})
	return self.builder.Module
}
