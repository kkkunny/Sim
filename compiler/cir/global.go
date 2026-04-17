package cir

import (
	"fmt"

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

type Typedef struct {
	Type Type
	Name string
}

func (b *Builder) BuildTypedef(t Type, name string) *Typedef {
	b.typedefCount++
	if name == "" {
		name = fmt.Sprintf("_t%d", b.typedefCount)
	}
	g := &Typedef{
		Type: t,
		Name: name,
	}
	b.Globals = append(b.Globals, g)
	return g
}

func (*Typedef) global() {}

func (g *Typedef) print(p *printer) {
	p.WriteString("typedef ")
	g.Type.printWithName(p, g.Name)
	p.WriteString(";")
}
