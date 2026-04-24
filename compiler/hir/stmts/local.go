package stmts

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type Local interface {
	hir.PrintWriter
	local()
}

type Block struct {
	Stmts []Local
}

func NewBlock() *Block {
	return &Block{}
}

func (*Block) local() {}

func (b *Block) Print(p *hir.Printer) {
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

func NewReturn(v ...Expr) *Return {
	return &Return{
		Value: optional.UnEmpty(stlslices.Last(v)),
	}
}

func (*Return) local() {}

func (r *Return) Print(p *hir.Printer) {
	if value, ok := r.Value.Value(); ok {
		p.WriteString("return ")
		p.WriteBy(value)
	} else {
		p.WriteString("return")
	}
}

type Let struct {
	Mut   bool
	Type  types.Type
	Name  string
	Value Expr
}

func (*Let) local()  {}
func (*Let) global() {}

func (l *Let) Print(p *hir.Printer) {
	p.WriteString("let ")
	p.WriteString(l.Name)
	p.WriteString(": ")
	p.WriteBy(l.Type)
	p.WriteString(" = ")
	p.WriteBy(l.Value)
}

func (l *Let) GetName() string {
	return l.Name
}

func (l *Let) GetType() types.Type {
	return l.Value.GetType()
}

func (e *Let) Mutable() bool {
	return e.Mut
}

type If struct {
	Condition Expr
	Body      *Block
	Else      optional.Optional[either.Either[*If, *Block]]
}

func NewIf(cond Expr, body *Block, next ...either.Either[*If, *Block]) *If {
	return &If{
		Condition: cond,
		Body:      body,
		Else:      optional.UnEmpty(stlslices.Last(next)),
	}
}

func (*If) local()  {}
func (*If) global() {}

func (l *If) Print(p *hir.Printer) {
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
	Condition Expr
	Body      *Block
}

func NewWhile(cond Expr, body *Block) *While {
	return &While{
		Condition: cond,
		Body:      body,
	}
}

func (*While) local()  {}
func (*While) global() {}

func (l *While) Print(p *hir.Printer) {
	p.WriteString("for ")
	p.WriteBy(l.Condition)
	p.WriteString(" ")
	p.WriteBy(l.Body)
}

type For struct {
	Var   *Param
	Range Expr
	Body  *Block
}

func NewFor(v *Param, rv Expr, body *Block) *For {
	return &For{
		Var:   v,
		Range: rv,
		Body:  body,
	}
}

func (*For) local()  {}
func (*For) global() {}

func (l *For) Print(p *hir.Printer) {
	p.WriteString("for ")
	if l.Var.Mut {
		p.WriteString("mut ")
	}
	p.WriteString(l.Var.Name)
	p.WriteString(" in ")
	p.WriteBy(l.Range)
	p.WriteString(" ")
	p.WriteBy(l.Body)
}
