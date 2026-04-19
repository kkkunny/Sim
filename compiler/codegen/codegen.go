package codegen

import (
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

type CodeGenerator struct {
	builder *cir.Builder

	idents    map[hir.Ident]cir.Namer
	typeCache map[string]*cir.AliasType
}

func New() *CodeGenerator {
	return &CodeGenerator{
		builder:   cir.NewBuilder(),
		idents:    make(map[hir.Ident]cir.Namer),
		typeCache: make(map[string]*cir.AliasType),
	}
}

func (c *CodeGenerator) Generate(program *hir.Program) *cir.Builder {
	c.buildProgram(program)
	return c.builder
}

func (c *CodeGenerator) buildProgram(program *hir.Program) {
	for _, g := range program.Globals {
		c.buildGlobal(g)
	}
}
