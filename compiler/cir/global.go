package cir

import (
	"github.com/kkkunny/stl/container/optional"
)

type Global interface {
	printWriter
	global()
}

type FuncDecl struct {
	Name       string
	Params     []*ParamDecl
	ReturnType Type
	Body       optional.Optional[*Block]
}

func (b *Builder) BuildFuncDecl(name string, rt Type, params []*ParamDecl) *FuncDecl {
	g := &FuncDecl{
		Name:       name,
		Params:     params,
		ReturnType: rt,
	}
	b.Globals = append(b.Globals, g)
	return g
}

func (*FuncDecl) global() {}

func (g *FuncDecl) print(p *printer) {
	p.WriteBy(g.ReturnType)
	p.WriteString(" ")
	p.WriteString(g.Name + "(")
	for i, param := range g.Params {
		p.WriteBy(param)
		if i < len(g.Params)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")

	if body, ok := g.Body.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(body)
	} else {
		p.WriteString(";")
	}
}
