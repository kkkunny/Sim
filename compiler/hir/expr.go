package hir

import (
	"math/big"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/enum"
)

type Expr interface {
	Local
	expr()
	GetType() Type
}

type Integer struct {
	Type  Type
	Value *big.Int
}

func (*Integer) expr()  {}
func (*Integer) local() {}

func (e *Integer) print(p *printer) {
	p.WriteString(e.Value.String())
}

func (e *Integer) GetType() Type {
	return e.Type
}

type Float struct {
	Type  Type
	Value *big.Float
}

func (*Float) expr()  {}
func (*Float) local() {}

func (e *Float) print(p *printer) {
	p.WriteString(e.Value.String())
}

func (e *Float) GetType() Type {
	return e.Type
}

type UnaryOp string

var UnaryOpEnum = enum.New[struct {
	Not UnaryOp `enum:"!"`
}]()

type Unary struct {
	Op   UnaryOp
	Expr Expr
}

func (*Unary) expr()  {}
func (*Unary) local() {}

func (e *Unary) print(p *printer) {
	p.WriteString("(")
	p.WriteString(string(e.Op))
	p.WriteBy(e.Expr)
	p.WriteString(")")
}

func (e *Unary) GetType() Type {
	return e.Expr.GetType()
}

type BinaryOp string

var BinaryOpEnum = enum.New[struct {
	Add BinaryOp `enum:"+"`
	Sub BinaryOp `enum:"-"`
	Mul BinaryOp `enum:"*"`
	Quo BinaryOp `enum:"/"`
	Rem BinaryOp `enum:"%"`
	And BinaryOp `enum:"&"`
	Or  BinaryOp `enum:"|"`
	Xor BinaryOp `enum:"^"`
}]()

type Binary struct {
	Op    BinaryOp
	Left  Expr
	Right Expr
}

func (*Binary) expr()  {}
func (*Binary) local() {}

func (e *Binary) print(p *printer) {
	p.WriteString("(")
	p.WriteBy(e.Left)
	p.WriteString(" ")
	p.WriteString(string(e.Op))
	p.WriteString(" ")
	p.WriteBy(e.Right)
	p.WriteString(")")
}

func (e *Binary) GetType() Type {
	return e.Left.GetType()
}

type Func struct {
	Params     []*Param
	ReturnType Type
	Body       optional.Optional[*Block]

	UsedExternalVariables []Ident
}

func (*Func) expr()  {}
func (*Func) local() {}

func (e *Func) print(p *printer) {
	p.WriteString("(")
	for i, param := range e.Params {
		p.WriteBy(param)
		if i < len(e.Params)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
	p.WriteString(" -> ")
	p.WriteBy(e.ReturnType)
	if body, ok := e.Body.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(body)
	}
}

func (e *Func) GetType() Type {
	params := stlslices.Map(e.Params, func(_ int, param *Param) Type {
		return param.Type
	})
	return NewFuncType(e.ReturnType, params...)
}

type Call struct {
	Func Expr
	Args []Expr
}

func (*Call) expr()  {}
func (*Call) local() {}

func (e *Call) print(p *printer) {
	p.WriteBy(e.Func)
	p.WriteString("(")
	for i, param := range e.Args {
		p.WriteBy(param)
		if i < len(e.Args)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}

func (e *Call) GetType() Type {
	return e.Func.GetType().(*FuncType).Return
}

type Tuple struct {
	Elems []Expr
}

func NewTuple(elems ...Expr) *Tuple {
	return &Tuple{Elems: elems}
}

func (*Tuple) expr()  {}
func (*Tuple) local() {}

func (e *Tuple) print(p *printer) {
	p.WriteString("(")
	for i, param := range e.Elems {
		p.WriteBy(param)
		if i < len(e.Elems)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}

func (e *Tuple) GetType() Type {
	return NewTupleType(stlslices.Map(e.Elems, func(_ int, e Expr) Type {
		return e.GetType()
	})...)
}
