package ast

import "github.com/kkkunny/Sim/compiler/token"

type Global interface {
	global()
	printWriter
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
