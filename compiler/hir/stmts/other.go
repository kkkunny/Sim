package stmts

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type Ident interface {
	Public() bool
	GetType() types.Type
	GetName() string
	Mutable() bool
}

type Package struct {
	Name         string
	Dependencies []*Package
	Globals      []Global
}

func (p *Package) Print(pr *hir.Printer) {
	for i, f := range p.Globals {
		pr.WriteBy(f)
		pr.NextLine()
		if i < len(p.Globals)-1 {
			pr.NextLine()
		}
	}
}

type Param struct {
	Mut  bool
	Name string
	Type types.Type
}

func NewParam(mut bool, t types.Type, name string) *Param {
	return &Param{
		Mut:  mut,
		Name: name,
		Type: t,
	}
}

func (p *Param) Public() bool {
	return false
}

func (p *Param) Print(pr *hir.Printer) {
	pr.WriteString(p.Name)
	pr.WriteString(": ")
	pr.WriteBy(p.Type)
}

func (p *Param) GetName() string {
	return p.Name
}

func (p *Param) GetType() types.Type {
	return p.Type
}

func (e *Param) Mutable() bool {
	return e.Mut
}
