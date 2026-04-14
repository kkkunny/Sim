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

func (c *CodeGenerator) Generate(a ast.Ast) string {
	c.buf.Reset()
	switch node := a.(type) {
	case *ast.FuncDecl:
		c.buf.WriteString(fmt.Sprintf("void %s() {}\n", node.Name))
	}
	return c.buf.String()
}
