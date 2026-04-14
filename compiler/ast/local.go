package ast

import (
	"github.com/kkkunny/stl/container/optional"
)

type Local interface {
	local()
	printWriter
}

type Block struct {
	Stmts []Local
}

func (b *Block) local() {}

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

func (r *Return) local() {}

func (r *Return) print(p *printer) {
	if value, ok := r.Value.Value(); ok {
		p.WriteString("return ")
		p.WriteBy(value)
	} else {
		p.WriteString("return")
	}
}
