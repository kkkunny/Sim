package codegen

import (
	"strings"

	"github.com/kkkunny/Sim/compiler/ast"
)

type CodeGenerator struct {
	buf strings.Builder
}

func New() *CodeGenerator {
	return &CodeGenerator{}
}

func (c *CodeGenerator) Generate(program *ast.Program) string {
	c.buf.Reset()
	for _, fn := range program.Functions {
		c.genFunc(fn)
	}
	return c.buf.String()
}
