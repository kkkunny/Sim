package ast

import (
	"strings"
)

type Program struct {
	Functions []*FuncDecl
}

func (p *Program) String() string {
	var b strings.Builder
	for i, f := range p.Functions {
		b.WriteString(f.String())
		if i < len(p.Functions)-1 {
			b.WriteString("\n\n")
		}
	}
	return b.String()
}

type ParamDecl struct {
	Name string
	Type Type
}

func (p *ParamDecl) String() string {
	return p.Name + ": " + p.Type.String()
}
