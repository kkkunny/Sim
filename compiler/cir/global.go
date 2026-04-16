package cir

import "github.com/kkkunny/stl/container/optional"

type Global interface {
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
