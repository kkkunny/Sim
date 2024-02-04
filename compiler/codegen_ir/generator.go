package codegen_ir

import (
	"github.com/kkkunny/stl/container/hashmap"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/runtime/types"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	irs linkedlist.LinkedList[hir.Global]

	target  mir.Target
	ctx     *mir.Context
	module  *mir.Module
	builder *mir.Builder

	values               hashmap.HashMap[hir.Expr, mir.Value]
	loops                hashmap.HashMap[hir.Loop, loop]
	strings              hashmap.HashMap[string, *mir.Constant]
	structs              hashmap.HashMap[hir.Type, pair.Pair[mir.StructType, *types.StructType]]
	funcCache            hashmap.HashMap[string, *mir.Function]
}

func New(target mir.Target, irs linkedlist.LinkedList[hir.Global]) *CodeGenerator {
	ctx := mir.NewContext(target)
	return &CodeGenerator{
		irs:     irs,
		target:  target,
		ctx:     ctx,
		module:  ctx.NewModule(),
		builder: ctx.NewBuilder(),
	}
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen() *mir.Module {
	// 类型声明
	stliter.Foreach[hir.Global](self.irs, func(v hir.Global) bool {
		st, ok := v.(*hir.StructDef)
		if ok {
			self.declStructDef(st)
		}
		return true
	})
	// 值声明
	stliter.Foreach[hir.Global](self.irs, func(v hir.Global) bool {
		self.codegenGlobalDecl(v)
		return true
	})
	// 值定义
	stliter.Foreach[hir.Global](self.irs, func(v hir.Global) bool {
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
	stliter.Foreach[hir.Global](self.irs, func(v hir.Global) bool {
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
