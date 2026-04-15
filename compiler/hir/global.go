package hir

import "github.com/kkkunny/stl/container/optional"

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

func (f *FuncDecl) global() {}

func (f *FuncDecl) print(p *printer) {
	p.WriteString("let ")
	p.WriteString(f.Name)
	p.WriteString(" = (")
	for i, param := range f.Params {
		if i > 0 {
			p.WriteString(", ")
		}
		p.WriteBy(param)
	}
	p.WriteString(")")
	p.WriteString(" -> ")
	p.WriteBy(f.ReturnType)
	if body, ok := f.Body.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(body)
	}
}
