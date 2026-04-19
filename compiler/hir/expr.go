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
	Mutable() bool
}

type IdentExpr struct {
	Define Ident
}

func NewIdentExpr(ident Ident) *IdentExpr {
	return &IdentExpr{
		Define: ident,
	}
}

func (*IdentExpr) expr()  {}
func (*IdentExpr) local() {}

func (e *IdentExpr) print(p *printer) {
	p.WriteString(e.Define.GetName())
}

func (e *IdentExpr) GetType() Type {
	return e.Define.GetType()
}

func (e *IdentExpr) Mutable() bool {
	return e.Define.Mutable()
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

func (e *Integer) Mutable() bool {
	return false
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

func (e *Float) Mutable() bool {
	return false
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

func (e *Unary) Mutable() bool {
	return false
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

	Assign    BinaryOp `enum:"="`
	AddAssign BinaryOp `enum:"+="`
	SubAssign BinaryOp `enum:"-="`
	MulAssign BinaryOp `enum:"*="`
	QuoAssign BinaryOp `enum:"/="`
	RemAssign BinaryOp `enum:"%="`
	AndAssign BinaryOp `enum:"&="`
	OrAssign  BinaryOp `enum:"|="`
	XorAssign BinaryOp `enum:"^="`
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
	switch e.Op {
	case BinaryOpEnum.Add, BinaryOpEnum.Sub, BinaryOpEnum.Mul, BinaryOpEnum.Quo, BinaryOpEnum.Rem, BinaryOpEnum.And, BinaryOpEnum.Or, BinaryOpEnum.Xor:
		return e.Left.GetType()
	case BinaryOpEnum.Assign, BinaryOpEnum.AddAssign, BinaryOpEnum.SubAssign, BinaryOpEnum.MulAssign, BinaryOpEnum.QuoAssign, BinaryOpEnum.RemAssign, BinaryOpEnum.AndAssign, BinaryOpEnum.OrAssign, BinaryOpEnum.XorAssign:
		return Unit
	default:
		panic("unreachable")
	}
}

func (e *Binary) Mutable() bool {
	return false
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

func (e *Func) Mutable() bool {
	return false
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

func (e *Call) Mutable() bool {
	return false
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

func (e *Tuple) Mutable() bool {
	return false
}

type TupleIndex struct {
	From  Expr
	Index *big.Int
}

func NewTupleIndex(from Expr, index *big.Int) *TupleIndex {
	return &TupleIndex{
		From:  from,
		Index: index,
	}
}

func (*TupleIndex) expr()  {}
func (*TupleIndex) local() {}

func (e *TupleIndex) print(p *printer) {
	p.WriteBy(e.From)
	p.WriteString("[")
	p.WriteString(e.Index.String())
	p.WriteString("]")
}

func (e *TupleIndex) GetType() Type {
	elems := e.From.GetType().(*TupleType).Elems
	return elems[e.Index.Int64()]
}

func (e *TupleIndex) Mutable() bool {
	return e.From.Mutable()
}

type Array struct {
	Type  Type
	Elems []Expr
}

func NewArray(t Type, elems ...Expr) *Array {
	return &Array{Type: t, Elems: elems}
}

func (*Array) expr()  {}
func (*Array) local() {}

func (e *Array) print(p *printer) {
	p.WriteString("[")
	for i, param := range e.Elems {
		p.WriteBy(param)
		if i < len(e.Elems)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString("]")
}

func (e *Array) GetType() Type {
	return e.Type
}

func (e *Array) Mutable() bool {
	return false
}

type ArrayIndex struct {
	From  Expr
	Index Expr
}

func NewArrayIndex(from, index Expr) *ArrayIndex {
	return &ArrayIndex{
		From:  from,
		Index: index,
	}
}

func (*ArrayIndex) expr()  {}
func (*ArrayIndex) local() {}

func (e *ArrayIndex) print(p *printer) {
	p.WriteBy(e.From)
	p.WriteString("[")
	p.WriteBy(e.Index)
	p.WriteString("]")
}

func (e *ArrayIndex) GetType() Type {
	return e.From.GetType().(*ArrayType).Elem
}

func (e *ArrayIndex) Mutable() bool {
	return e.From.Mutable()
}
