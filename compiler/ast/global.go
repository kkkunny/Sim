package ast

import "github.com/kkkunny/Sim/compiler/token"

type Global interface {
	global()
	printWriter
}

type Import struct {
	Pkgs []token.Token
}

func (*Import) global() {}

func (i *Import) print(p *printer) {
	p.WriteString("import ")
	for index, pkg := range i.Pkgs {
		p.WriteToken(pkg)
		if index < len(i.Pkgs)-1 {
			p.WriteString("::")
		}
	}
}

type TypeDef struct {
	Name token.Token
	Type Type
}

func (*TypeDef) global() {}

func (t *TypeDef) print(p *printer) {
	p.WriteString("type ")
	p.WriteToken(t.Name)
	p.WriteString(" ")
	p.WriteBy(t.Type)
}
