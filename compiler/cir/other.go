package cir

type Namer interface {
	SetName(s string)
	GetName() string
}

type Param struct {
	Name string
	Type Type
}

func NewParam(name string, t Type) *Param {
	return &Param{
		Name: name,
		Type: t,
	}
}

func (p *Param) print(pr *printer) {
	p.Type.printWithName(pr, p.Name)
}

func (p *Param) SetName(s string) {
	p.Name = s
}

func (p *Param) GetName() string {
	return p.Name
}
