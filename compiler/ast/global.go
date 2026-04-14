package ast

import (
	"strings"
)

type Global interface {
	global()
	String() string
}

type FuncDecl struct {
	Name       string
	Params     []*ParamDecl
	ReturnType Type
}

func (f *FuncDecl) global() {}

func (f *FuncDecl) String() string {
	var b strings.Builder
	b.WriteString("func ")
	b.WriteString(f.Name)
	b.WriteString("(")
	for i, p := range f.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(p.String())
	}
	b.WriteString(")")
	if f.ReturnType != nil {
		b.WriteString(": ")
		b.WriteString(f.ReturnType.String())
	}
	b.WriteString(" {\n")
	b.WriteString("}")
	return b.String()
}
