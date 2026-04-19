package hir

import (
	"github.com/kkkunny/stl/container/optional"
)

type Local interface {
	printWriter
	local()
}

type Block struct {
	Stmts []Local
}

func (*Block) local() {}

func (b *Block) print(p *printer) {
	p.WriteString("{")
	if len(b.Stmts) == 0 {
		p.WriteString("}")
		return
	}
	p.NextLine(+1)
	for i, stmt := range b.Stmts {
		p.WriteBy(stmt)
		if i < len(b.Stmts)-1 {
			p.NextLine()
		} else {
			p.NextLine(-1)
		}
	}
	p.WriteString("}")
}

type Return struct {
	Value optional.Optional[Expr]
}

func (*Return) local() {}

func (r *Return) print(p *printer) {
	if value, ok := r.Value.Value(); ok {
		p.WriteString("return ")
		p.WriteBy(value)
	} else {
		p.WriteString("return")
	}
}

type Let struct {
	Mut   bool
	Name  string
	Value Expr
}

func (*Let) local()  {}
func (*Let) global() {}

func (l *Let) print(p *printer) {
	p.WriteString("let ")
	p.WriteString(l.Name)
	p.WriteString(" = ")
	p.WriteBy(l.Value)
}

func (l *Let) GetName() string {
	return l.Name
}

func (l *Let) GetType() Type {
	return l.Value.GetType()
}

func (e *Let) Mutable() bool {
	return e.Mut
}
