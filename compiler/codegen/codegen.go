package codegen

import (
	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir"
)

type CodeGenerator struct{}

func New() *CodeGenerator {
	return &CodeGenerator{}
}

func (c *CodeGenerator) Generate(program *hir.Program) string {
	ir := c.buildProgram(program)
	return cir.NewEmitter().Emit(ir)
}

func (c *CodeGenerator) buildProgram(program *hir.Program) *cir.Program {
	globals := make([]cir.Global, len(program.Functions))
	for i, fn := range program.Functions {
		globals[i] = c.buildFunc(fn)
	}
	return &cir.Program{Globals: globals}
}
