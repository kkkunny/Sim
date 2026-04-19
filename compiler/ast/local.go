package ast

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/token"
)

type Local interface {
	local()
	printWriter
}

type Block struct {
	BeginPosition reader.Position
	Stmts         []Local
	EndPosition   reader.Position
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

func (b *Block) Position() reader.Position {
	return reader.MixPosition(b.BeginPosition, b.EndPosition)
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
	Name  token.Token
	Value Expr
}

func (*Let) local()  {}
func (*Let) global() {}

func (l *Let) print(p *printer) {
	p.WriteString("let ")
	if l.Mut {
		p.WriteString("mut ")
	}
	p.WriteToken(l.Name)
	p.WriteString(" = ")
	p.WriteBy(l.Value)
}
