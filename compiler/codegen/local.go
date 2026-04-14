package codegen

import (
	"github.com/kkkunny/Sim/compiler/ast"
)

func (c *CodeGenerator) genLocal(local ast.Local) {
	switch local := local.(type) {
	case *ast.Block:
		c.genBlock(local)
	case *ast.Return:
		c.genReturn(local)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) genBlock(b *ast.Block) {
	c.buf.WriteString("{\n")
	for _, stmt := range b.Stmts {
		c.buf.WriteString("    ")
		c.genLocal(stmt)
		c.buf.WriteString("\n")
	}
	c.buf.WriteString("}")
}

func (c *CodeGenerator) genReturn(_ *ast.Return) {
	c.buf.WriteString("return;")
}
