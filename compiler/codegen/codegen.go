package codegen

import (
	"fmt"
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
		c.generateFunc(fn)
	}
	return c.buf.String()
}

func (c *CodeGenerator) generateFunc(fn *ast.FuncDecl) {
	if fn.ReturnType != nil {
		c.buf.WriteString(c.genType(fn.ReturnType))
	} else {
		c.buf.WriteString("void")
	}
	c.buf.WriteString(" ")

	c.buf.WriteString(fn.Name + "(")
	for i, p := range fn.Params {
		if i > 0 {
			c.buf.WriteString(", ")
		}
		c.buf.WriteString(fmt.Sprintf("%s %s", c.genType(p.Type), p.Name))
	}
	c.buf.WriteString(") {\n")

	if fn.ReturnType != nil {
		c.buf.WriteString("    return 0;\n")
	}

	c.buf.WriteString("}\n\n")
}
