package cir

import (
	"github.com/kkkunny/stl/container/optional"
)

type Global interface {
	Stmt
	global()
}

type FuncDecl struct {
	Name       string
	Params     []*Param
	ReturnType Type
	Body       optional.Optional[*Block]
}

func NewFuncDecl(name string, rt Type, params ...*Param) *FuncDecl {
	return &FuncDecl{
		Name:       name,
		Params:     params,
		ReturnType: rt,
	}
}

func (*FuncDecl) stmt()   {}
func (*FuncDecl) global() {}

func (g *FuncDecl) print(p *printer) {
	p.WriteString("static ")
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

func (g *FuncDecl) SetName(s string) {
	g.Name = s
}

func (g *FuncDecl) GetName() string {
	return g.Name
}

type Typedef struct {
	Type Type
	Name string
}

func NewTypedef(t Type, name string) *Typedef {
	return &Typedef{
		Type: t,
		Name: name,
	}
}

func (*Typedef) stmt()   {}
func (*Typedef) global() {}

func (g *Typedef) print(p *printer) {
	p.WriteString("typedef ")
	g.Type.printWithName(p, g.Name)
	p.WriteString(";")
}

func (g *Typedef) SetName(s string) {
	g.Name = s
}

func (g *Typedef) GetName() string {
	return g.Name
}
