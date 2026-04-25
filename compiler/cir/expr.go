package cir

import (
	"math/big"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/enum"
)

type Expr interface {
	printWriter
	expr()
}

type IdentExpr struct {
	Name string
}

func NewIdentExpr(name string) *IdentExpr {
	return &IdentExpr{name}
}

func (*IdentExpr) expr() {}

func (e *IdentExpr) print(p *printer) {
	p.WriteString(e.Name)
}

type Integer struct {
	Value *big.Int
}

func NewInteger(v *big.Int) *Integer {
	return &Integer{Value: v}
}

func (*Integer) expr() {}

func (e *Integer) print(p *printer) {
	p.WriteString(e.Value.String())
}

type FloatExpr struct {
	Value *big.Float
}

func (*FloatExpr) expr() {}

func (e *FloatExpr) print(p *printer) {
	p.WriteString(e.Value.String())
}

type UnaryOp string

var UnaryOpEnum = enum.New[struct {
	Not     UnaryOp `enum:"NOT"`
	Mul     UnaryOp `enum:"*"`
	AND     UnaryOp `enum:"&"`
	SelfAdd UnaryOp `enum:"SELFADD"`
}]()

type Unary struct {
	Op   UnaryOp
	Expr Expr
}

func NewUnary(op UnaryOp, expr Expr) *Unary {
	return &Unary{
		Op:   op,
		Expr: expr,
	}
}

func (*Unary) expr() {}

func (e *Unary) print(p *printer) {
	p.WriteString(string(e.Op))
	p.WriteString("(")
	p.WriteBy(e.Expr)
	p.WriteString(")")
}

type BinaryOp string

var BinaryOpEnum = enum.New[struct {
	Add BinaryOp `enum:"ADD"`
	Sub BinaryOp `enum:"SUB"`
	Mul BinaryOp `enum:"MUL"`
	Quo BinaryOp `enum:"QUO"`
	Rem BinaryOp `enum:"REM"`
	And BinaryOp `enum:"AND"`
	Or  BinaryOp `enum:"OR"`
	Xor BinaryOp `enum:"XOR"`
	Eq  BinaryOp `enum:"EQ"`
	Neq BinaryOp `enum:"NEQ"`
	Lt  BinaryOp `enum:"LT"`
	Lte BinaryOp `enum:"LTE"`
	Gt  BinaryOp `enum:"GT"`
	Gte BinaryOp `enum:"GTE"`
}]()

type Binary struct {
	Op    BinaryOp
	Left  Expr
	Right Expr
}

func NewBinary(op BinaryOp, left, right Expr) *Binary {
	return &Binary{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

func (*Binary) expr() {}

func (e *Binary) print(p *printer) {
	p.WriteString(string(e.Op))
	p.WriteString("(")
	p.WriteBy(e.Left)
	p.WriteString(", ")
	p.WriteBy(e.Right)
	p.WriteString(")")
}

type FuncExpr struct {
	Decl *FuncDecl
}

func (*FuncExpr) expr() {}

func (e *FuncExpr) print(p *printer) {
	p.WriteString(e.Decl.Name)
}

var (
	True  = NewMacroExpr("true")
	False = NewMacroExpr("false")
)

type MacroExpr struct {
	Name string
	Args []Expr
}

func NewMacroExpr(name string, args ...Expr) *MacroExpr {
	return &MacroExpr{
		Name: name,
		Args: args,
	}
}

func (*MacroExpr) expr() {}

func (e *MacroExpr) print(p *printer) {
	p.WriteString(e.Name)
	if len(e.Args) > 0 {
		p.WriteString("(")
		for i, arg := range e.Args {
			p.WriteBy(arg)
			if i < len(e.Args)-1 {
				p.WriteString(", ")
			}
		}
		p.WriteString(")")
	}
}

type Call struct {
	Func Expr
	Args []Expr
}

func NewCall(f Expr, args ...Expr) *Call {
	return &Call{
		Func: f,
		Args: args,
	}
}

func (*Call) expr() {}

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

type Covert struct {
	Type  Type
	Value Expr
}

func NewCovert(t Type, v Expr) *Covert {
	return &Covert{
		Type:  t,
		Value: v,
	}
}

func (*Covert) expr() {}

func (e *Covert) print(p *printer) {
	p.WriteString("(")
	p.WriteBy(e.Type)
	p.WriteString(")")
	p.WriteString("(")
	p.WriteBy(e.Value)
	p.WriteString(")")
}

type GetMember struct {
	From Expr
	Name string
}

func NewGetMember(f Expr, name string) *GetMember {
	return &GetMember{
		From: f,
		Name: name,
	}
}

func (*GetMember) expr() {}

func (e *GetMember) print(p *printer) {
	p.WriteBy(e.From)
	p.WriteString(".")
	p.WriteString(e.Name)
}

type Struct struct {
	Type   optional.Optional[Type]
	Fields map[string]Expr
}

func NewStruct(fields map[string]Expr, t ...Type) *Struct {
	return &Struct{
		Type:   optional.UnEmpty(stlslices.Last(t)),
		Fields: fields,
	}
}

func (*Struct) expr() {}

func (e *Struct) print(p *printer) {
	if t, ok := e.Type.Value(); ok && t.zeroSize() {
		p.WriteString("ZERO_TYPE_VALUE")
		return
	}

	if t, ok := e.Type.Value(); ok {
		p.WriteString("(")
		p.WriteBy(t)
		p.WriteString(")")
	}

	p.WriteString("{")
	var i int
	for fn, fv := range e.Fields {
		p.WriteString(".")
		p.WriteString(fn)
		p.WriteString("=")
		p.WriteBy(fv)
		if i < len(e.Fields)-1 {
			p.WriteString(", ")
		}
		i++
	}
	p.WriteString("}")
}

type Array struct {
	Type  optional.Optional[Type]
	Elems []Expr
}

func NewArray(elems []Expr, t ...Type) *Array {
	return &Array{
		Type:  optional.UnEmpty(stlslices.Last(t)),
		Elems: elems,
	}
}

func (*Array) expr() {}

func (e *Array) print(p *printer) {
	if t, ok := e.Type.Value(); ok && t.zeroSize() {
		p.WriteString("ZERO_TYPE_VALUE")
		return
	}

	if t, ok := e.Type.Value(); ok {
		p.WriteString("(")
		p.WriteBy(t)
		p.WriteString(")")
	}

	p.WriteString("{")
	for i, elem := range e.Elems {
		p.WriteBy(elem)
		if i < len(e.Elems)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString("}")
}

type Offset struct {
	From   Expr
	Offset Expr
}

func NewOffset(from, offset Expr) *Offset {
	return &Offset{
		From:   from,
		Offset: offset,
	}
}

func (*Offset) expr() {}

func (e *Offset) print(p *printer) {
	p.WriteBy(e.From)
	p.WriteString("[")
	p.WriteBy(e.Offset)
	p.WriteString("]")
}

type AssignOp string

var AssignOpEnum = enum.New[struct {
	Assign    AssignOp `enum:"="`
	AddAssign AssignOp `enum:"+="`
	SubAssign AssignOp `enum:"-="`
	MulAssign AssignOp `enum:"*="`
	QuoAssign AssignOp `enum:"/="`
	RemAssign AssignOp `enum:"%="`
	AndAssign AssignOp `enum:"&="`
	OrAssign  AssignOp `enum:"|="`
	XorAssign AssignOp `enum:"^="`
}]()

type Assign struct {
	Op    AssignOp
	Left  Expr
	Right Expr
}

func NewAssign(op AssignOp, left Expr, right Expr) *Assign {
	return &Assign{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

func (*Assign) expr() {}

func (e *Assign) print(p *printer) {
	p.WriteBy(e.Left)
	p.WriteString(" ")
	p.WriteString(string(e.Op))
	p.WriteString(" ")
	p.WriteBy(e.Right)
}

type Ternary struct {
	Condition Expr
	TrueExpr  Expr
	FalseExpr Expr
}

func NewTernaryExpr(cond, trueExpr, falseExpr Expr) *Ternary {
	return &Ternary{
		Condition: cond,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}
}

func (*Ternary) expr() {}

func (e *Ternary) print(p *printer) {
	p.WriteString("(")
	p.WriteBy(e.Condition)
	p.WriteString(" ? ")
	p.WriteBy(e.TrueExpr)
	p.WriteString(" : ")
	p.WriteBy(e.FalseExpr)
	p.WriteString(")")
}
