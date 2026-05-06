package globals

import (
	"github.com/kkkunny/Sim/compiler/hir"
)

type Package struct {
	Name         string
	Path         string
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
