package codegen

import (
	"github.com/kkkunny/llvm"

	"github.com/kkkunny/Sim/analyse"
	"github.com/kkkunny/Sim/mean"
)

// CodeGenerator 代码生成器
type CodeGenerator struct {
	analyser *analyse.Analyser // 语法分析器

	ctx     llvm.Context
	module  llvm.Module
	builder llvm.Builder
}

func New(analyser *analyse.Analyser) *CodeGenerator {
	module := llvm.NewModule("main")
	return &CodeGenerator{
		analyser: analyser,
		ctx:      module.Context(),
		module:   module,
		builder:  llvm.NewBuilder(),
	}
}

// Codegen 代码生成
func (self *CodeGenerator) Codegen() llvm.Module {
	nodes := self.analyser.Analyse()
	nodes.Iterator().Foreach(func(v mean.Global) bool {
		self.codegenGlobal(v)
		return true
	})
	return self.module
}
