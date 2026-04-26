package stmts

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type Global interface {
	hir.PrintWriter
	global()
}

type TypeDef struct {
	Type types.CustomType
}

func NewTypeDef(name string, underlying types.Type) *TypeDef {
	return &TypeDef{
		Type: types.NewCustomType(name, underlying),
	}
}

func (*TypeDef) global() {}

func (t *TypeDef) Print(p *hir.Printer) {
	p.WriteString("type ")
	p.WriteString(t.Type.GetName())
	p.WriteString(" ")
	p.WriteBy(t.Type.GetUnderlying())
}
