package globals

import (
	"github.com/kkkunny/Sim/compiler/hir"
)

type Global interface {
	hir.PrintWriter
	Global()
}

type TypeDef struct {
	Pub        bool
	Name       string
	Underlying hir.Type
}

func NewTypeDef(pub bool, name string) *TypeDef {
	return &TypeDef{
		Pub:  pub,
		Name: name,
	}
}

func (*TypeDef) Global() {}

func (t *TypeDef) Public() bool {
	return t.Pub
}

func (t *TypeDef) Print(p *hir.Printer) {
	p.WriteString("type ")
	p.WriteString(t.Name)
	p.WriteString(" ")
	p.WriteBy(t.Underlying)
}
