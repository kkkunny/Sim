package codegen

import (
	"github.com/kkkunny/stl/container/tuple"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

type Context struct {
	idents    map[stmts.Ident]string
	typeCache map[string]*cir.AliasType
}

func NewContext() *Context {
	return &Context{
		idents:    make(map[stmts.Ident]string),
		typeCache: make(map[string]*cir.AliasType),
	}
}

type CodeGenerator struct {
	pkg     *stmts.Package
	builder *cir.Builder

	ctx *Context

	captureVarsMap map[tuple.Tuple2[*stmts.Func, stmts.Ident]]*cir.GetMember
	currentFunc    *stmts.Func
}

func New(ctx *Context, pkg *stmts.Package) *CodeGenerator {
	return &CodeGenerator{
		pkg:     pkg,
		builder: cir.NewBuilder(),

		ctx: ctx,

		captureVarsMap: make(map[tuple.Tuple2[*stmts.Func, stmts.Ident]]*cir.GetMember),
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
