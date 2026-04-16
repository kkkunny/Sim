package codegen

import (
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

type CodeGenerator struct {
	builder *cir.Builder
}

func New() *CodeGenerator {
	return &CodeGenerator{
		builder: cir.NewBuilder(),
	}
}

func (c *CodeGenerator) Generate(program *hir.Program) string {
	c.buildProgram(program)
	return cir.NewEmitter().Emit(c.builder)
}

func (c *CodeGenerator) buildProgram(program *hir.Program) {
	for _, g := range program.Globals {
		c.buildGlobal(g)
	}
}
