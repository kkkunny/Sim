package ast

import (
	"github.com/kkkunny/Sim/compiler/token"
)

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
	Name token.Token
	Type Type
}

func (p *ParamDecl) print(pr *printer) {
	pr.WriteToken(p.Name)
	pr.WriteString(": ")
	pr.WriteBy(p.Type)
}
