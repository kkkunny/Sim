package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/iterator"
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/mean"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	means linkedlist.LinkedList[mean.Global]

	target  *llvm.Target
	ctx     llvm.Context
	module  llvm.Module
	builder llvm.Builder

	values  map[mean.Expr]llvm.Value
	loops   hashmap.HashMap[mean.Loop, loop]
	strings hashmap.HashMap[string, *llvm.GlobalValue]
	structs hashmap.HashMap[*mean.StructDef, llvm.StructType]

	genericParams hashmap.HashMap[*mean.GenericParam, mean.Type]
}

func New(target *llvm.Target, means linkedlist.LinkedList[mean.Global]) *CodeGenerator {
	ctx := llvm.NewContext()
	module := ctx.NewModule("main")
	module.SetTarget(target)
	return &CodeGenerator{
		means: means,
		target:   target,
		ctx:      ctx,
		module:   module,
		builder:  ctx.NewBuilder(),
		values:   make(map[mean.Expr]llvm.Value),
		structs: hashmap.NewHashMap[*mean.StructDef, llvm.StructType](),
	}
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen() llvm.Module {
	// 类型声明
	iterator.Foreach(self.means, func(v mean.Global) bool {
		st, ok := v.(*mean.StructDef)
		if ok {
			self.declStructDef(st)
		}
		return true
	})
	// 值声明
	iterator.Foreach(self.means, func(v mean.Global) bool {
		self.codegenGlobalDecl(v)
		return true
	})
	// 值定义
	iterator.Foreach(self.means, func(v mean.Global) bool {
		self.codegenGlobalDef(v)
		return true
	})
	// 初始化函数
	// FIXME: jit无法运行llvm.global_ctors
	self.builder.MoveToAfter(self.getInitFunction().EntryBlock())
	self.builder.CreateRet(nil)
	// 主函数
	var hasMain bool
	iterator.Foreach(self.means, func(v mean.Global) bool {
		if funcNode, ok := v.(*mean.FuncDef); ok && funcNode.Name == "main" {
			hasMain = true
			f := self.values[funcNode].(llvm.Function)
			self.builder.MoveToAfter(self.getMainFunction().EntryBlock())
			var ret llvm.Value = self.builder.CreateCall("", self.codegenFuncType(funcNode.GetType().(*mean.FuncType)), f)
			self.builder.CreateRet(&ret)
			return false
		}
		return true
	})
	if !hasMain {
		self.builder.MoveToAfter(self.getMainFunction().EntryBlock())
		var ret llvm.Value = self.ctx.ConstInteger(self.ctx.IntegerType(8), 0)
		self.builder.CreateRet(&ret)
	}
	return self.module
}
