package ast

import "fmt"

type Program struct {
	Functions []*FuncDecl
}

func (p *Program) String() string {
	s := "Program:\n"
	for _, f := range p.Functions {
		s += "  " + f.String() + "\n"
	}
	return s
}

type ParamDecl struct {
	Name string
	Type Type
}

func (p *ParamDecl) String() string {
	return fmt.Sprintf("%s: %s", p.Name, p.Type.String())
}
