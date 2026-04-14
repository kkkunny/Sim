package ast

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/token"
)

type Global interface {
	global()
	printWriter
}

type FuncDecl struct {
	Name       token.Token
	Params     []*ParamDecl
	ReturnType optional.Optional[Type]
	Body       optional.Optional[*Block]
}

func (f *FuncDecl) global() {}

func (f *FuncDecl) print(p *printer) {
	p.WriteString("let ")
	p.WriteToken(f.Name)
	p.WriteString(" = (")
	for i, param := range f.Params {
		if i > 0 {
			p.WriteString(", ")
		}
		p.WriteBy(param)
	}
	p.WriteString(")")
	if rt, ok := f.ReturnType.Value(); ok {
		p.WriteString(" -> ")
		p.WriteBy(rt)
	}
	if body, ok := f.Body.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(body)
	}
}
