package llvmUtil

import "github.com/kkkunny/go-llvm"

type Builder struct {
	llvm.Context
	llvm.Target
	llvm.Module
	llvm.Builder
}

func NewBuilder(target llvm.Target) *Builder {
	ctx := llvm.GlobalContext
	module := ctx.NewModule("main")
	module.SetTarget(target)
	return &Builder{
		Context: ctx,
		Target:  target,
		Module:  module,
		Builder: ctx.NewBuilder(),
	}
}
