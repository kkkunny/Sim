package llgen

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/hir/locals"
)

func (c *CodeGenerator) genLocal(local locals.Local) {
	switch local := local.(type) {
	case *locals.Block:
		for _, s := range local.Stmts {
			c.genLocal(s)
		}
	case *locals.Return:
		c.ensureBlock()
		c.genReturn(local)
	case *locals.Let:
		c.ensureBlock()
		c.genLocalLet(local)
	case locals.Expr:
		c.ensureBlock()
		c.genExpr(local)
	default:
		panic(fmt.Errorf("llgen: 暂不支持的语句 %T", local))
	}
}

func (c *CodeGenerator) genReturn(l *locals.Return) {
	if v, ok := l.Value.Value(); ok {
		c.builder.Ret(c.genExpr(v))
	} else {
		c.builder.RetVoid()
	}
	c.terminated = true
}

func (c *CodeGenerator) genLocalLet(l *locals.Let) {
	ptr := c.builder.Alloca(c.genType(l.GetType()), l.Name)
	c.ctx.idents[l] = &Ident{Name: l.Name, Local: ptr.Value}
	if v, ok := l.Value.Value(); ok {
		c.builder.Store(c.genExpr(v), ptr)
	}
}
