package cir

import (
	"math/big"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/enum"
)

type Expr interface {
	Local
	expr()
}

type IdentExpr struct {
	Name string
}

func NewIdentExpr(name string) *IdentExpr {
	return &IdentExpr{name}
}

func (*IdentExpr) local() {}
func (*IdentExpr) expr()  {}

func (e *IdentExpr) print(p *printer) {
	p.WriteString(e.Name)
}

type IntegerExpr struct {
	Value *big.Int
}

func (*IntegerExpr) local() {}
func (*IntegerExpr) expr()  {}

func (e *IntegerExpr) print(p *printer) {
	p.WriteString(e.Value.String())
}

type FloatExpr struct {
	Value *big.Float
}

func (*FloatExpr) local() {}
func (*FloatExpr) expr()  {}

func (e *FloatExpr) print(p *printer) {
	p.WriteString(e.Value.String())
}

type UnaryOp string

var UnaryOpEnum = enum.New[struct {
	Not UnaryOp `enum:"NOT"`
	Mul UnaryOp `enum:"*"`
	AND UnaryOp `enum:"&"`
}]()

type UnaryExpr struct {
	Op   UnaryOp
	Expr Expr
}

func NewUnaryExpr(op UnaryOp, expr Expr) *UnaryExpr {
	return &UnaryExpr{
		Op:   op,
		Expr: expr,
	}
}

func (*UnaryExpr) local() {}
func (*UnaryExpr) expr()  {}

func (e *UnaryExpr) print(p *printer) {
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
}]()

type BinaryExpr struct {
	Op    BinaryOp
	Left  Expr
	Right Expr
}

func (*BinaryExpr) local() {}
func (*BinaryExpr) expr()  {}

func (e *BinaryExpr) print(p *printer) {
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

func (*FuncExpr) local() {}
func (*FuncExpr) expr()  {}

func (e *FuncExpr) print(p *printer) {
	p.WriteString(e.Decl.Name)
}

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

func (*MacroExpr) local() {}
func (*MacroExpr) expr()  {}

func (e *MacroExpr) print(p *printer) {
	p.WriteString(e.Name)
	p.WriteString("(")
	for i, arg := range e.Args {
		p.WriteBy(arg)
		if i < len(e.Args)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
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

func (*Call) local() {}
func (*Call) expr()  {}

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

func (*Covert) local() {}
func (*Covert) expr()  {}

func (e *Covert) print(p *printer) {
	p.WriteString("(")
	p.WriteBy(e.Type)
	p.WriteString(")")
	p.WriteString("(")
	p.WriteBy(e.Value)
	p.WriteString(")")
}

type Member struct {
	From Expr
	Name string
}

func NewMember(f Expr, name string) *Member {
	return &Member{
		From: f,
		Name: name,
	}
}

func (*Member) local() {}
func (*Member) expr()  {}

func (e *Member) print(p *printer) {
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

func (*Struct) local() {}
func (*Struct) expr()  {}

func (e *Struct) print(p *printer) {
	if len(e.Fields) == 0 {
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
	Elems []Expr
}

func NewArray(elems ...Expr) *Array {
	return &Array{Elems: elems}
}

func (*Array) local() {}
func (*Array) expr()  {}

func (e *Array) print(p *printer) {
	if len(e.Elems) == 0 {
		p.WriteString("ZERO_TYPE_VALUE")
		return
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

func (*Offset) local() {}
func (*Offset) expr()  {}

func (e *Offset) print(p *printer) {
	p.WriteBy(e.From)
	p.WriteString("[")
	p.WriteBy(e.Offset)
	p.WriteString("]")
}
