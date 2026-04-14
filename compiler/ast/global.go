package ast

import (
	"strings"

	"github.com/kkkunny/Sim/compiler/token"
)

type Global interface {
	global()
	String() string
}

type FuncDecl struct {
	Name       token.Token
	Params     []*ParamDecl
	ReturnType Type
}

func (f *FuncDecl) global() {}

func (f *FuncDecl) String() string {
	var b strings.Builder
	b.WriteString("let ")
	b.WriteString(f.Name.OriginText)
	b.WriteString(" = (")
	for i, p := range f.Params {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(p.String())
	}
	b.WriteString(")")
	b.WriteString(" -> ")
	b.WriteString(f.ReturnType.String())
	b.WriteString(" {\n")
	b.WriteString("}")
	return b.String()
}
