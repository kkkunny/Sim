package ast

import (
	"github.com/kkkunny/stl/container/either"
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
	Attributes []Attribute
	Public     bool
	Mut        bool
	Name       token.Token
	Type       optional.Optional[Type] // Type和Value必有一个不为空
	Value      optional.Optional[Expr] // Type和Value必有一个不为空

	Bind optional.Optional[Type] // 方法绑定的类型
}

func (*Let) local()  {}
func (*Let) global() {}

func (l *Let) print(p *printer) {
	for _, a := range l.Attributes {
		p.WriteBy(a)
		p.NextLine()
	}
	if l.Public {
		p.WriteString("pub ")
	}
	p.WriteString("let ")
	if l.Mut {
		p.WriteString("mut ")
	}
	p.WriteToken(l.Name)
	if forType, ok := l.Bind.Value(); ok {
		p.WriteString(" | ")
		p.WriteBy(forType)
	}
	if t, ok := l.Type.Value(); ok {
		p.WriteString(": ")
		p.WriteBy(t)
	}
	if v, ok := l.Value.Value(); ok {
		p.WriteString(" = ")
		p.WriteBy(v)
	}
}

type If struct {
	Condition Expr
	Body      *Block
	Else      optional.Optional[either.Either[*If, *Block]]
}

func (*If) local() {}

func (l *If) print(p *printer) {
	p.WriteString("if ")
	p.WriteBy(l.Condition)
	p.WriteString(" ")
	p.WriteBy(l.Body)

	if next, ok := l.Else.Value(); ok {
		p.WriteString(" else ")
		if elseif, ok := next.TryLeft(); ok {
			p.WriteBy(elseif)
		} else {
			p.WriteBy(next.Right())
		}
	}
}

type While struct {
	Condition optional.Optional[Expr]
	Body      *Block
}

func (*While) local() {}

func (l *While) print(p *printer) {
	p.WriteString("for ")
	if cond, ok := l.Condition.Value(); ok {
		p.WriteBy(cond)
	}
	p.WriteString(" ")
	p.WriteBy(l.Body)
}

type For struct {
	Mut      bool
	Variable token.Token
	Range    Expr
	Body     *Block
}

func (*For) local() {}

func (l *For) print(p *printer) {
	p.WriteString("for ")
	if l.Mut {
		p.WriteString("mut ")
	}
	p.WriteString(l.Variable.OriginText)
	p.WriteString(" in ")
	p.WriteBy(l.Range)
	p.WriteString(" ")
	p.WriteBy(l.Body)
}
