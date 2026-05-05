package ast

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/token"
)

type Global interface {
	global()
	printWriter
}

type Import struct {
	Alias optional.Optional[token.Token]
	Pkgs  []token.Token
}

func (*Import) global() {}

func (i *Import) print(p *printer) {
	p.WriteString("import ")
	for index, pkg := range i.Pkgs {
		p.WriteToken(pkg)
		if index < len(i.Pkgs)-1 {
			p.WriteString("::")
		}
	}
	if alias, ok := i.Alias.Value(); ok {
		p.WriteString(" as ")
		p.WriteToken(alias)
	}
}

type TypeDef struct {
	Public bool
	Name   token.Token
	Type   Type
}

func (*TypeDef) global() {}

func (t *TypeDef) print(p *printer) {
	if t.Public {
		p.WriteString("pub ")
	}
	p.WriteString("type ")
	p.WriteToken(t.Name)
	p.WriteString(" ")
	p.WriteBy(t.Type)
}
