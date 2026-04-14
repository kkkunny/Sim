package codegen

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/cir"
)

type CodeGenerator struct{}

func New() *CodeGenerator {
	return &CodeGenerator{}
}

func (c *CodeGenerator) Generate(program *ast.Program) string {
	ir := c.buildProgram(program)
	return cir.NewEmitter().Emit(ir)
}

func (c *CodeGenerator) buildProgram(program *ast.Program) *cir.Program {
	globals := make([]cir.Global, len(program.Functions))
	for i, fn := range program.Functions {
		globals[i] = c.buildFunc(fn)
	}
	return &cir.Program{Globals: globals}
}
