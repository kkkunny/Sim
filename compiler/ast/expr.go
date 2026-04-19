package ast

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/token"
)

type Expr interface {
	Local
	expr()
	Position() reader.Position
}

type IdentExpr struct {
	Name token.Token
}

func (e *IdentExpr) local() {}
func (e *IdentExpr) expr()  {}

func (e *IdentExpr) print(p *printer) {
	p.WriteToken(e.Name)
}

func (e *IdentExpr) Position() reader.Position {
	return e.Name.Position
}

type Integer struct {
	Value token.Token
}

func (e *Integer) local() {}
func (e *Integer) expr()  {}

func (e *Integer) print(p *printer) {
	p.WriteToken(e.Value)
}

func (e *Integer) Position() reader.Position {
	return e.Value.Position
}

type Unary struct {
	Op   token.Token
	Expr Expr
}

func (e *Unary) local() {}
func (e *Unary) expr()  {}

func (e *Unary) print(p *printer) {
	p.WriteString("(")
	p.WriteToken(e.Op)
	p.WriteBy(e.Expr)
	p.WriteString(")")
}

func (e *Unary) Position() reader.Position {
	return reader.MixPosition(e.Op.Position, e.Expr.Position())
}

type Binary struct {
	Op    token.Token
	Left  Expr
	Right Expr
}

func (e *Binary) local() {}
func (e *Binary) expr()  {}

func (e *Binary) print(p *printer) {
	p.WriteString("(")
	p.WriteBy(e.Left)
	p.WriteString(" ")
	p.WriteToken(e.Op)
	p.WriteString(" ")
	p.WriteBy(e.Right)
	p.WriteString(")")
}

func (e *Binary) Position() reader.Position {
	return reader.MixPosition(e.Left.Position(), e.Right.Position())
}

type Func struct {
	BeginPosition reader.Position
	Params        []*ParamDecl
	ReturnType    optional.Optional[Type]
	Body          optional.Optional[*Block]
	EndPosition   reader.Position
}

func (e *Func) local() {}
func (e *Func) expr()  {}

func (e *Func) print(p *printer) {
	p.WriteString("(")
	for i, param := range e.Params {
		p.WriteBy(param)
		if i < len(e.Params)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
	if rt, ok := e.ReturnType.Value(); ok {
		p.WriteString(" -> ")
		p.WriteBy(rt)
	}
	if body, ok := e.Body.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(body)
	}
}

func (e *Func) Position() reader.Position {
	return reader.MixPosition(e.BeginPosition, e.EndPosition)
}

type Call struct {
	Func        Expr
	Args        []Expr
	EndPosition reader.Position
}

func (e *Call) local() {}
func (e *Call) expr()  {}

func (e *Call) print(p *printer) {
	p.WriteBy(e.Func)
	p.WriteString("(")
	for i, arg := range e.Args {
		p.WriteBy(arg)
		if i < len(e.Args)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}

func (e *Call) Position() reader.Position {
	return reader.MixPosition(e.Func.Position(), e.EndPosition)
}

type Tuple struct {
	BeginPosition reader.Position
	Elems         []Expr
	EndPosition   reader.Position
}

func (e *Tuple) local() {}
func (e *Tuple) expr()  {}

func (e *Tuple) print(p *printer) {
	p.WriteString("(")
	for i, arg := range e.Elems {
		p.WriteBy(arg)
		if i < len(e.Elems)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}

func (e *Tuple) Position() reader.Position {
	return reader.MixPosition(e.BeginPosition, e.EndPosition)
}

type Index struct {
	From        Expr
	Index       Expr
	EndPosition reader.Position
}

func (e *Index) local() {}
func (e *Index) expr()  {}

func (e *Index) print(p *printer) {
	p.WriteBy(e.From)
	p.WriteString("[")
	p.WriteBy(e.Index)
	p.WriteString("]")
}

func (e *Index) Position() reader.Position {
	return reader.MixPosition(e.From.Position(), e.EndPosition)
}

type Array struct {
	BeginPosition reader.Position
	Elems         []Expr
	EndPosition   reader.Position
}

func (e *Array) local() {}
func (e *Array) expr()  {}

func (e *Array) print(p *printer) {
	p.WriteString("[")
	for i, arg := range e.Elems {
		p.WriteBy(arg)
		if i < len(e.Elems)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString("]")
}

func (e *Array) Position() reader.Position {
	return reader.MixPosition(e.BeginPosition, e.EndPosition)
}
