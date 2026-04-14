package ast

import (
	"strings"

	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/token"
)

type Global interface {
	global()
	String() string
}

type FuncDecl struct {
	Name       token.Token
	Params     []*ParamDecl
	ReturnType optional.Optional[Type]
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
	if rt, ok := f.ReturnType.Value(); ok {
		b.WriteString(" -> ")
		b.WriteString(rt.String())
	}
	b.WriteString(" {\n")
	b.WriteString("}")
	return b.String()
}
