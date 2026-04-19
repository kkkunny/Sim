package hir

type Ident interface {
	GetType() Type
	GetName() string
	Mutable() bool
}

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

type Param struct {
	Mut  bool
	Name string
	Type Type
}

func NewParam(mut bool, t Type, name string) *Param {
	return &Param{
		Mut:  mut,
		Name: name,
		Type: t,
	}
}

func (p *Param) print(pr *printer) {
	pr.WriteString(p.Name)
	pr.WriteString(": ")
	pr.WriteBy(p.Type)
}

func (p *Param) GetName() string {
	return p.Name
}

func (p *Param) GetType() Type {
	return p.Type
}

func (e *Param) Mutable() bool {
	return e.Mut
}
