package codegen

import (
	"github.com/kkkunny/stl/container/tuple"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

type CodeGenerator struct {
	builder *cir.Builder

	idents         map[stmts.Ident]cir.Namer
	typeCache      map[string]*cir.AliasType
	captureVarsMap map[tuple.Tuple2[*stmts.Func, stmts.Ident]]*cir.GetMember
	currentFunc    *stmts.Func
}

func New() *CodeGenerator {
	return &CodeGenerator{
		builder:        cir.NewBuilder(),
		idents:         make(map[stmts.Ident]cir.Namer),
		typeCache:      make(map[string]*cir.AliasType),
		captureVarsMap: make(map[tuple.Tuple2[*stmts.Func, stmts.Ident]]*cir.GetMember),
	}
}

func (c *CodeGenerator) Generate(program *stmts.Program) *cir.Builder {
	c.genProgram(program)
	return c.builder
}

func (c *CodeGenerator) genProgram(program *stmts.Program) {
	for _, g := range program.Globals {
		td, ok := g.(*stmts.TypeDef)
		if !ok {
			continue
		}
		c.genCustomType(td.Type)
	}

	for _, g := range program.Globals {
		if stlval.Is[*stmts.TypeDef](g) {
			continue
		}
		c.genGlobalDecl(g)
	}
	for _, g := range program.Globals {
		if stlval.Is[*stmts.TypeDef](g) {
			continue
		}
		c.genGlobalDef(g)
	}
}
