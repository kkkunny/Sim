package codegen_ir

import (
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/queue"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	hir *hir.Result

	target  mir.Target
	ctx     *mir.Context
	module  *mir.Module
	builder *mir.Builder

	values           hashmap.HashMap[hir.Expr, mir.Value]
	types            hashmap.HashMap[*hir.CustomType, mir.Type]
	loops            hashmap.HashMap[hir.Loop, loop]
	strings          hashmap.HashMap[string, *mir.Constant]
	funcCache        hashmap.HashMap[string, *mir.Function]
	lambdaCaptureMap queue.Queue[hashmap.HashMap[hir.Ident, mir.Value]]
}

func New(target mir.Target, ir *hir.Result) *CodeGenerator {
	ctx := mir.NewContext(target)
	return &CodeGenerator{
		hir:     ir,
		target:  target,
		ctx:     ctx,
		module:  ctx.NewModule(),
		builder: ctx.NewBuilder(),
	}
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen() *mir.Module {
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
	for cursor := self.getInitFunction().Blocks().Front(); cursor != nil; cursor = cursor.Next() {
		if !cursor.Value.Terminated() {
			self.builder.MoveTo(cursor.Value)
			self.builder.BuildReturn()
		}
	}
	// 主函数
	stliter.Foreach[hir.Global](self.hir.Globals, func(v hir.Global) bool {
		if funcNode, ok := v.(*hir.FuncDef); ok && funcNode.Name == "main" {
			f := self.values.Get(funcNode).(*mir.Function)
			self.builder.MoveTo(self.getMainFunction().Blocks().Front().Value)
			self.builder.BuildCall(f)
			self.builder.BuildReturn(mir.NewUint(self.ctx.U8(), 0))
			return false
		}
		return true
	})
	return self.module
}
