package hir

type Program struct {
	Globals []Global
}

func (p *Program) print(pr *printer) {
	for i, f := range p.Globals {
		pr.WriteBy(f)
		pr.NextLine()
		if i < len(p.Globals)-1 {
			pr.NextLine()
		}
	}
}

type ParamDecl struct {
	Name string
	Type Type
}

func (p *ParamDecl) print(pr *printer) {
	pr.WriteString(p.Name)
	pr.WriteString(": ")
	pr.WriteBy(p.Type)
}
