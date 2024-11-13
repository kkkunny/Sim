package llvmUtil

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/hashmap"
)

type Builder struct {
	llvm.Context
	llvm.Target
	llvm.Module
	llvm.Builder

	stringMap hashmap.HashMap[string, llvm.Constant]
}

func NewBuilder(target llvm.Target) *Builder {
	ctx := llvm.GlobalContext
	module := ctx.NewModule("main")
	module.SetTarget(target)
	return &Builder{
		Context:   ctx,
		Target:    target,
		Module:    module,
		Builder:   ctx.NewBuilder(),
		stringMap: hashmap.StdWith[string, llvm.Constant](),
	}
}
