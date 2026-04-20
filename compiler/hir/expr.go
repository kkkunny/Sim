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
	Temporary() bool
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

func (e *IdentExpr) Temporary() bool {
	return false
}

type Integer struct {
	Type  Type
	Value *big.Int
}

func NewInteger(t Type, v *big.Int) *Integer {
	return &Integer{
		Type:  t,
		Value: v,
	}
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

func (e *Integer) Temporary() bool {
	return true
}

type Float struct {
	Type  Type
	Value *big.Float
}

func NewFloat(t Type, v *big.Float) *Float {
	return &Float{
		Type:  t,
		Value: v,
	}
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

func (e *Float) Temporary() bool {
	return true
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

func (e *Unary) Temporary() bool {
	return true
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

func (e *Binary) Temporary() bool {
	return true
}

type Func struct {
	Return Type
	Params []*Param
	Body   optional.Optional[*Block]

	UsedExternalVariables []Ident
}

func NewFunc(rt Type, params ...*Param) *Func {
	return &Func{
		Return: rt,
		Params: params,
	}
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
	p.WriteBy(e.Return)
	if body, ok := e.Body.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(body)
	}
}

func (e *Func) GetType() Type {
	params := stlslices.Map(e.Params, func(_ int, param *Param) Type {
		return param.Type
	})
	return NewFuncType(e.Return, params...)
}

func (e *Func) Mutable() bool {
	return false
}

func (e *Func) Temporary() bool {
	return true
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

func (e *Call) Temporary() bool {
	return true
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

func (e *Tuple) Temporary() bool {
	return true
}

type EmptyTuple struct {
	Type Type
}

func NewEmptyTuple(t Type) *EmptyTuple {
	return &EmptyTuple{Type: t}
}

func (*EmptyTuple) expr()  {}
func (*EmptyTuple) local() {}

func (e *EmptyTuple) print(p *printer) {
	p.WriteString("(...)")
}

func (e *EmptyTuple) GetType() Type {
	return e.Type
}

func (e *EmptyTuple) Mutable() bool {
	return false
}

func (e *EmptyTuple) Temporary() bool {
	return true
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

func (e *TupleIndex) Temporary() bool {
	return e.From.Temporary()
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

func (e *Array) Temporary() bool {
	return true
}

type EmptyArray struct {
	Type Type
}

func NewEmptyArray(t Type) *EmptyArray {
	return &EmptyArray{Type: t}
}

func (*EmptyArray) expr()  {}
func (*EmptyArray) local() {}

func (e *EmptyArray) print(p *printer) {
	p.WriteString("[...]")
}

func (e *EmptyArray) GetType() Type {
	return e.Type
}

func (e *EmptyArray) Mutable() bool {
	return false
}

func (e *EmptyArray) Temporary() bool {
	return true
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

func (e *ArrayIndex) Temporary() bool {
	return e.From.Temporary()
}

type Covert interface {
	Expr
	GetFrom() Expr
	GetTo() Type
}

type Union struct {
	From  Expr
	To    Type
	Index uint8
}

func NewUnion(from Expr, to Type, index uint8) *Union {
	return &Union{
		From:  from,
		To:    to,
		Index: index,
	}
}

func (*Union) expr()  {}
func (*Union) local() {}

func (e *Union) print(p *printer) {
	p.WriteBy(e.From)
	p.WriteString(" as ")
	p.WriteBy(e.To)
}

func (e *Union) GetType() Type {
	return e.To
}

func (e *Union) Mutable() bool {
	return false
}

func (e *Union) Temporary() bool {
	return true
}

func (e *Union) GetFrom() Expr {
	return e.From
}

func (e *Union) GetTo() Type {
	return e.To
}
