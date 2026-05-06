package hir

type Ident interface {
	Public() bool
	GetType() Type
	GetName() string
	Mutable() bool
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

func (p *Param) Public() bool {
	return false
}

func (p *Param) Print(pr *Printer) {
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
