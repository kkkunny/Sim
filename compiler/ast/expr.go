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
	Pkg  optional.Optional[token.Token]
	Name token.Token
}

func (e *IdentExpr) local() {}
func (e *IdentExpr) expr()  {}

func (e *IdentExpr) print(p *printer) {
	if pkg, ok := e.Pkg.Value(); ok {
		p.WriteToken(pkg)
		p.WriteString("::")
	}
	p.WriteToken(e.Name)
}

func (e *IdentExpr) Position() reader.Position {
	if pkg, ok := e.Pkg.Value(); ok {
		return reader.MixPosition(pkg.Position, e.Name.Position)
	}
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

type Char struct {
	Value token.Token
}

func (e *Char) local() {}
func (e *Char) expr()  {}

func (e *Char) print(p *printer) {
	p.WriteToken(e.Value)
}

func (e *Char) Position() reader.Position {
	return e.Value.Position
}

type String struct {
	Value token.Token
}

func (e *String) local() {}
func (e *String) expr()  {}

func (e *String) print(p *printer) {
	p.WriteToken(e.Value)
}

func (e *String) Position() reader.Position {
	return e.Value.Position
}

type Unary struct {
	Op   token.Token
	Expr Expr
}

func (e *Unary) local() {}
func (e *Unary) expr()  {}

func (e *Unary) print(p *printer) {
	p.WriteToken(e.Op)
	p.WriteBy(e.Expr)
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
	p.WriteBy(e.Left)
	p.WriteString(" ")
	p.WriteToken(e.Op)
	p.WriteString(" ")
	p.WriteBy(e.Right)
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

type As struct {
	Left  Expr
	Right Type
}

func (e *As) local() {}
func (e *As) expr()  {}

func (e *As) print(p *printer) {
	p.WriteBy(e.Left)
	p.WriteString(" as ")
	p.WriteBy(e.Right)
}

func (e *As) Position() reader.Position {
	return reader.MixPosition(e.Left.Position(), e.Right.Position())
}

type Boolean struct {
	Value token.Token
}

func (e *Boolean) local() {}
func (e *Boolean) expr()  {}

func (e *Boolean) print(p *printer) {
	p.WriteString(e.Value.OriginText)
}

func (e *Boolean) Position() reader.Position {
	return e.Value.Position
}

type GetReference struct {
	BeginPosition reader.Position
	Mut           bool
	Value         Expr
}

func (e *GetReference) local() {}
func (e *GetReference) expr()  {}

func (e *GetReference) print(p *printer) {
	p.WriteString("&")
	if e.Mut {
		p.WriteString("mut ")
	}
	p.WriteBy(e.Value)
}

func (e *GetReference) Position() reader.Position {
	return reader.MixPosition(e.BeginPosition, e.Value.Position())
}

type Ternary struct {
	Condition Expr
	TrueExpr  Expr
	FalseExpr Expr
}

func (e *Ternary) local() {}
func (e *Ternary) expr()  {}

func (e *Ternary) print(p *printer) {
	p.WriteBy(e.Condition)
	p.WriteString(" ? ")
	p.WriteBy(e.TrueExpr)
	p.WriteString(" : ")
	p.WriteBy(e.FalseExpr)
}

func (e *Ternary) Position() reader.Position {
	return reader.MixPosition(e.Condition.Position(), e.FalseExpr.Position())
}

type StructFieldInit struct {
	Name  token.Token
	Value Expr
}

type Struct struct {
	Type        Type
	Fields      []*StructFieldInit
	EndPosition reader.Position
}

func (e *Struct) local() {}
func (e *Struct) expr()  {}

func (e *Struct) print(p *printer) {
	p.WriteBy(e.Type)
	p.WriteString("{")
	if len(e.Fields) > 0 {
		p.NextLine(+1)
		for i, f := range e.Fields {
			p.WriteToken(f.Name)
			p.WriteString(": ")
			p.WriteBy(f.Value)
			p.WriteString(",")
			if i < len(e.Fields)-1 {
				p.NextLine()
			} else {
				p.NextLine(-1)
			}
		}
	}
	p.WriteString("}")
}

func (e *Struct) Position() reader.Position {
	return reader.MixPosition(e.Type.Position(), e.EndPosition)
}

type Member struct {
	From Expr
	Name token.Token
}

func (e *Member) local() {}
func (e *Member) expr()  {}

func (e *Member) print(p *printer) {
	p.WriteBy(e.From)
	p.WriteString(".")
	p.WriteToken(e.Name)
}

func (e *Member) Position() reader.Position {
	return reader.MixPosition(e.From.Position(), e.Name.Position)
}
