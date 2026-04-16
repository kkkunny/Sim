package cir

type ParamDecl struct {
	Name string
	Type Type
}

func (p *ParamDecl) print(pr *printer) {
	p.Type.printWithName(pr, p.Name)
}
