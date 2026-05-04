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
	Pub  bool
	Type types.CustomType
}

func NewTypeDef(pub bool) *TypeDef {
	return &TypeDef{Pub: pub}
}

func (*TypeDef) global() {}

func (t *TypeDef) Public() bool {
	return t.Pub
}

func (t *TypeDef) GetName() string {
	return t.Type.GetName()
}

func (t *TypeDef) Print(p *hir.Printer) {
	p.WriteString("type ")
	p.WriteString(t.Type.GetName())
	p.WriteString(" ")
	p.WriteBy(t.Type.GetUnderlying())
}
