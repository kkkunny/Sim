package stmts

import (
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
