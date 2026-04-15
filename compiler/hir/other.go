package hir

type Program struct {
	Functions []*FuncDecl
}

func (p *Program) print(pr *printer) {
	for i, f := range p.Functions {
		pr.WriteBy(f)
		pr.NextLine()
		if i < len(p.Functions)-1 {
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
