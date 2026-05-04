package cir

import (
	"github.com/kkkunny/stl/container/optional"
)

type Global interface {
	Stmt
	global()
}

type Func struct {
	Static     bool
	Name       string
	Params     []*Param
	ReturnType Type
	Body       optional.Optional[*Block]
}

func NewFunc(name string, rt Type, params ...*Param) *Func {
	return &Func{
		Name:       name,
		Params:     params,
		ReturnType: rt,
	}
}

func (*Func) stmt()   {}
func (*Func) global() {}

func (g *Func) print(p *printer) {
	if g.Static {
		p.WriteString("static ")
	}
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

func (g *Func) printDecl(p *printer) {
	if g.Static {
		p.WriteString("static ")
	}
	p.WriteBy(g.ReturnType)
	p.WriteString(" ")
	p.WriteString(g.Name + "(")
	for i, param := range g.Params {
		p.WriteBy(param)
		if i < len(g.Params)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(");")
}

func (g *Func) SetName(s string) {
	g.Name = s
}

func (g *Func) GetName() string {
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

type StructTypeDef struct {
	Type *StructType
}

func NewStructTypeDef(t *StructType) *StructTypeDef {
	return &StructTypeDef{
		Type: t,
	}
}

func (*StructTypeDef) stmt()   {}
func (*StructTypeDef) global() {}

func (g *StructTypeDef) print(p *printer) {
	p.WriteBy(g.Type)
	p.WriteString(";")
}

func (g *StructTypeDef) SetName(s string) {
	g.Type.Name = s
}

func (g *StructTypeDef) GetName() string {
	return g.Type.Name
}
