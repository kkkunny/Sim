package codegen

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/ast"
)

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
	c.buf.WriteString(")")

	if body, ok := fn.Body.Value(); ok {
		c.buf.WriteString(" ")
		c.genBlock(body)
	} else {
		c.buf.WriteString(";")
	}
}
