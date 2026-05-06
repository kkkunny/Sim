package codegen

import (
	"github.com/kkkunny/stl/container/tuple"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/hir/locals"
)

type CodeGenerator struct {
	pkg     *globals.Package
	builder *cir.Builder

	ctx *Context

	captureVarsMap map[tuple.Tuple2[*locals.Func, hir.Ident]]*cir.GetField
	currentFunc    *locals.Func
}

func New(ctx *Context, pkg *globals.Package) *CodeGenerator {
	return &CodeGenerator{
		pkg:     pkg,
		builder: cir.NewBuilder(),

		ctx: ctx,

		captureVarsMap: make(map[tuple.Tuple2[*locals.Func, hir.Ident]]*cir.GetField),
	}
}

func (c *CodeGenerator) Generate() *cir.Builder {
	for _, g := range c.pkg.Globals {
		c.genTypeDecl(g)
	}
	for _, g := range c.pkg.Globals {
		c.genTypeDef(g)
	}

	for _, g := range c.pkg.Globals {
		c.genGlobalValue(g)
	}
	return c.builder
}
