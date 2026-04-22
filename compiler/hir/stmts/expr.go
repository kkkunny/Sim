package stmts

import (
	"math/big"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/enum"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

type Expr interface {
	Local
	expr()
	GetType() types.Type
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

func (e *IdentExpr) Print(p *hir.Printer) {
	p.WriteString(e.Define.GetName())
}

func (e *IdentExpr) GetType() types.Type {
	return e.Define.GetType()
}

func (e *IdentExpr) Mutable() bool {
	return e.Define.Mutable()
}

func (e *IdentExpr) Temporary() bool {
	return false
}

type Integer struct {
	Type  types.Type
	Value *big.Int
}

func NewInteger(t types.Type, v *big.Int) *Integer {
	return &Integer{
		Type:  t,
		Value: v,
	}
}

func (*Integer) expr()  {}
func (*Integer) local() {}

func (e *Integer) Print(p *hir.Printer) {
	p.WriteString(e.Value.String())
}

func (e *Integer) GetType() types.Type {
	return e.Type
}

func (e *Integer) Mutable() bool {
	return false
}

func (e *Integer) Temporary() bool {
	return true
}

type Float struct {
	Type  types.Type
	Value *big.Float
}

func NewFloat(t types.Type, v *big.Float) *Float {
	return &Float{
		Type:  t,
		Value: v,
	}
}

func (*Float) expr()  {}
func (*Float) local() {}

func (e *Float) Print(p *hir.Printer) {
	p.WriteString(e.Value.String())
}

func (e *Float) GetType() types.Type {
	return e.Type
}

func (e *Float) Mutable() bool {
	return false
}

func (e *Float) Temporary() bool {
	return true
}

type Unary interface {
	Expr
	GetOpTarget() Expr
}

type BitReverse struct {
	Target Expr
}

func NewBitReverse(t Expr) *BitReverse {
	return &BitReverse{
		Target: t,
	}
}

func (*BitReverse) expr()  {}
func (*BitReverse) local() {}

func (e *BitReverse) Print(p *hir.Printer) {
	p.WriteString("!")
	p.WriteBy(e.Target)
}

func (e *BitReverse) GetType() types.Type {
	return e.Target.GetType()
}

func (e *BitReverse) Mutable() bool {
	return false
}

func (e *BitReverse) Temporary() bool {
	return true
}

func (e *BitReverse) GetOpTarget() Expr {
	return e.Target
}

type BooleanReverse struct {
	Target Expr
}

func NewBooleanReverse(t Expr) *BooleanReverse {
	return &BooleanReverse{
		Target: t,
	}
}

func (*BooleanReverse) expr()  {}
func (*BooleanReverse) local() {}

func (e *BooleanReverse) Print(p *hir.Printer) {
	p.WriteString("!")
	p.WriteBy(e.Target)
}

func (e *BooleanReverse) GetType() types.Type {
	return e.Target.GetType()
}

func (e *BooleanReverse) Mutable() bool {
	return false
}

func (e *BooleanReverse) Temporary() bool {
	return true
}

func (e *BooleanReverse) GetOpTarget() Expr {
	return e.Target
}

type GetRef struct {
	Mut    bool
	Target Expr
}

func NewGetRef(mut bool, t Expr) *GetRef {
	return &GetRef{
		Mut:    mut,
		Target: t,
	}
}

func (*GetRef) expr()  {}
func (*GetRef) local() {}

func (e *GetRef) Print(p *hir.Printer) {
	p.WriteString("&")
	if e.Mut {
		p.WriteString("mut ")
	}
	p.WriteBy(e.Target)
}

func (e *GetRef) GetType() types.Type {
	return types.NewPointerType(e.Mut, e.Target.GetType())
}

func (e *GetRef) Mutable() bool {
	return e.Mut
}

func (e *GetRef) Temporary() bool {
	return true
}

func (e *GetRef) GetOpTarget() Expr {
	return e.Target
}

type DeRef struct {
	Target Expr
}

func NewDeRef(t Expr) *DeRef {
	return &DeRef{
		Target: t,
	}
}

func (*DeRef) expr()  {}
func (*DeRef) local() {}

func (e *DeRef) Print(p *hir.Printer) {
	p.WriteString("*")
	p.WriteBy(e.Target)
}

func (e *DeRef) GetType() types.Type {
	return e.Target.GetType().(*types.PointerType).Elem
}

func (e *DeRef) Mutable() bool {
	return e.Target.Mutable()
}

func (e *DeRef) Temporary() bool {
	return e.Target.Temporary()
}

func (e *DeRef) GetOpTarget() Expr {
	return e.Target
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

func (e *Binary) Print(p *hir.Printer) {
	p.WriteBy(e.Left)
	p.WriteString(" ")
	p.WriteString(string(e.Op))
	p.WriteString(" ")
	p.WriteBy(e.Right)
}

func (e *Binary) GetType() types.Type {
	switch e.Op {
	case BinaryOpEnum.Add, BinaryOpEnum.Sub, BinaryOpEnum.Mul, BinaryOpEnum.Quo, BinaryOpEnum.Rem, BinaryOpEnum.And, BinaryOpEnum.Or, BinaryOpEnum.Xor:
		return e.Left.GetType()
	case BinaryOpEnum.Assign, BinaryOpEnum.AddAssign, BinaryOpEnum.SubAssign, BinaryOpEnum.MulAssign, BinaryOpEnum.QuoAssign, BinaryOpEnum.RemAssign, BinaryOpEnum.AndAssign, BinaryOpEnum.OrAssign, BinaryOpEnum.XorAssign:
		return types.Unit
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
	Return types.Type
	Params []*Param
	Body   optional.Optional[*Block]

	UsedExternalVariables []Ident
}

func NewFunc(rt types.Type, params ...*Param) *Func {
	return &Func{
		Return: rt,
		Params: params,
	}
}

func (*Func) expr()  {}
func (*Func) local() {}

func (e *Func) Print(p *hir.Printer) {
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

func (e *Func) GetType() types.Type {
	params := stlslices.Map(e.Params, func(_ int, param *Param) types.Type {
		return param.Type
	})
	return types.NewFuncType(e.Return, params...)
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

func (e *Call) Print(p *hir.Printer) {
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

func (e *Call) GetType() types.Type {
	return e.Func.GetType().(*types.FuncType).Return
}

func (e *Call) Mutable() bool {
	return false
}

func (e *Call) Temporary() bool {
	return true
}

type Tuple struct {
	Type  types.Type
	Elems []Expr
}

func NewTuple(t types.Type, elems ...Expr) *Tuple {
	return &Tuple{Type: t, Elems: elems}
}

func (*Tuple) expr()  {}
func (*Tuple) local() {}

func (e *Tuple) Print(p *hir.Printer) {
	p.WriteString("(")
	for i, param := range e.Elems {
		p.WriteBy(param)
		if i < len(e.Elems)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
}

func (e *Tuple) GetType() types.Type {
	return e.Type
}

func (e *Tuple) Mutable() bool {
	return false
}

func (e *Tuple) Temporary() bool {
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

func (e *TupleIndex) Print(p *hir.Printer) {
	p.WriteBy(e.From)
	p.WriteString("[")
	p.WriteString(e.Index.String())
	p.WriteString("]")
}

func (e *TupleIndex) GetType() types.Type {
	elems := e.From.GetType().(*types.TupleType).Elems
	return elems[e.Index.Int64()]
}

func (e *TupleIndex) Mutable() bool {
	return e.From.Mutable()
}

func (e *TupleIndex) Temporary() bool {
	return e.From.Temporary()
}

type Array struct {
	Type  types.Type
	Elems []Expr
}

func NewArray(t types.Type, elems ...Expr) *Array {
	return &Array{Type: t, Elems: elems}
}

func (*Array) expr()  {}
func (*Array) local() {}

func (e *Array) Print(p *hir.Printer) {
	p.WriteString("[")
	for i, param := range e.Elems {
		p.WriteBy(param)
		if i < len(e.Elems)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString("]")
}

func (e *Array) GetType() types.Type {
	return e.Type
}

func (e *Array) Mutable() bool {
	return false
}

func (e *Array) Temporary() bool {
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

func (e *ArrayIndex) Print(p *hir.Printer) {
	p.WriteBy(e.From)
	p.WriteString("[")
	p.WriteBy(e.Index)
	p.WriteString("]")
}

func (e *ArrayIndex) GetType() types.Type {
	return e.From.GetType().(*types.ArrayType).Elem
}

func (e *ArrayIndex) Mutable() bool {
	return e.From.Mutable()
}

func (e *ArrayIndex) Temporary() bool {
	return e.From.Temporary()
}

type Boolean struct {
	Value bool
}

func NewBoolean(v bool) *Boolean {
	return &Boolean{
		Value: v,
	}
}

func (*Boolean) expr()  {}
func (*Boolean) local() {}

func (e *Boolean) Print(p *hir.Printer) {
	if e.Value {
		p.WriteString("true")
	} else {
		p.WriteString("false")
	}
}

func (e *Boolean) GetType() types.Type {
	return types.Bool
}

func (e *Boolean) Mutable() bool {
	return false
}

func (e *Boolean) Temporary() bool {
	return true
}

type Covert interface {
	Expr
	GetFrom() Expr
}

type Union struct {
	From  Expr
	To    types.Type
	Index uint8
}

func NewUnion(from Expr, to types.Type, index uint8) *Union {
	return &Union{
		From:  from,
		To:    to,
		Index: index,
	}
}

func (*Union) expr()  {}
func (*Union) local() {}

func (e *Union) Print(p *hir.Printer) {
	p.WriteBy(e.From)
	p.WriteString(" as ")
	p.WriteBy(e.To)
}

func (e *Union) GetType() types.Type {
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

type NumberCovert struct {
	From Expr
	To   types.Type
}

func NewNumberCovert(from Expr, to types.Type) *NumberCovert {
	return &NumberCovert{
		From: from,
		To:   to,
	}
}

func (*NumberCovert) expr()  {}
func (*NumberCovert) local() {}

func (e *NumberCovert) Print(p *hir.Printer) {
	p.WriteBy(e.From)
	p.WriteString(" as ")
	p.WriteBy(e.To)
}

func (e *NumberCovert) GetType() types.Type {
	return e.To
}

func (e *NumberCovert) Mutable() bool {
	return false
}

func (e *NumberCovert) Temporary() bool {
	return true
}

func (e *NumberCovert) GetFrom() Expr {
	return e.From
}

type Ternary struct {
	Condition Expr
	TrueExpr  Expr
	FalseExpr Expr
}

func NewTernary(cond, trueExpr, falseExpr Expr) *Ternary {
	return &Ternary{
		Condition: cond,
		TrueExpr:  trueExpr,
		FalseExpr: falseExpr,
	}
}

func (*Ternary) expr()  {}
func (*Ternary) local() {}

func (e *Ternary) Print(p *hir.Printer) {
	p.WriteBy(e.Condition)
	p.WriteString(" ? ")
	p.WriteBy(e.TrueExpr)
	p.WriteString(" : ")
	p.WriteBy(e.FalseExpr)
}

func (e *Ternary) GetType() types.Type {
	return e.TrueExpr.GetType()
}

func (e *Ternary) Mutable() bool {
	return e.TrueExpr.Mutable() && e.FalseExpr.Mutable()
}

func (e *Ternary) Temporary() bool {
	return e.TrueExpr.Temporary() || e.FalseExpr.Temporary()
}
