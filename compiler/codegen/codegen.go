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
		c.genFunc(fn)
	}
	return c.buf.String()
}

func (c *CodeGenerator) genFunc(fn *ast.FuncDecl) {
	if rt, ok := fn.ReturnType.Value(); ok {
		c.buf.WriteString(c.genType(rt))
	} else {
		c.buf.WriteString("void")
	}
	c.buf.WriteString(" ")

	c.buf.WriteString(fn.Name.OriginText + "(")
	for i, p := range fn.Params {
		if i > 0 {
			c.buf.WriteString(", ")
		}
		c.buf.WriteString(fmt.Sprintf("%s %s", c.genType(p.Type), p.Name.OriginText))
	}
	c.buf.WriteString(") {\n")
	c.buf.WriteString("}\n\n")
}
