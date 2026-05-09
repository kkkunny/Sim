package locals

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
	GetType() hir.Type
	Mutable() bool
	Temporary() bool
}

type Literal interface {
	Expr
	TryToType(t hir.Type)
}

type IdentExpr struct {
	Define hir.Ident
}

func NewIdentExpr(ident hir.Ident) *IdentExpr {
	return &IdentExpr{
		Define: ident,
	}
}

func (*IdentExpr) expr()  {}
func (*IdentExpr) local() {}

func (e *IdentExpr) Print(p *hir.Printer) {
	p.WriteString(e.Define.GetName())
}

func (e *IdentExpr) GetType() hir.Type {
	return e.Define.GetType()
}

func (e *IdentExpr) Mutable() bool {
	return e.Define.Mutable()
}

func (e *IdentExpr) Temporary() bool {
	return false
}

type Integer struct {
	Type  types.IntegerType
	Value *big.Int
}

func NewInteger(t types.IntegerType, v *big.Int) *Integer {
	return &Integer{
		Type:  t,
		Value: v,
	}
}

func (*Integer) expr()  {}
func (*Integer) local() {}

func (i *Integer) Print(p *hir.Printer) {
	p.WriteString(i.Value.String())
}

func (i *Integer) GetType() hir.Type {
	return i.Type
}

func (i *Integer) Mutable() bool {
	return false
}

func (i *Integer) Temporary() bool {
	return true
}

func (i *Integer) TryToType(t hir.Type) {
	it, ok := t.(types.IntegerType)
	if !ok {
		return
	}
	i.Type = it
}

type Float struct {
	Type  types.FloatType
	Value *big.Float
}

func NewFloat(t types.FloatType, v *big.Float) *Float {
	return &Float{
		Type:  t,
		Value: v,
	}
}

func (*Float) expr()  {}
func (*Float) local() {}

func (f *Float) Print(p *hir.Printer) {
	p.WriteString(f.Value.String())
}

func (f *Float) GetType() hir.Type {
	return f.Type
}

func (f *Float) Mutable() bool {
	return false
}

func (f *Float) Temporary() bool {
	return true
}

func (f *Float) TryToType(t hir.Type) {
	ft, ok := t.(types.FloatType)
	if !ok {
		return
	}
	f.Type = ft
}

type Boolean struct {
	Type  types.BooleanType
	Value bool
}

func NewBoolean(t types.BooleanType, v bool) *Boolean {
	return &Boolean{
		Type:  t,
		Value: v,
	}
}

func (*Boolean) expr()  {}
func (*Boolean) local() {}

func (b *Boolean) Print(p *hir.Printer) {
	if b.Value {
		p.WriteString("true")
	} else {
		p.WriteString("false")
	}
}

func (b *Boolean) GetType() hir.Type {
	return b.Type
}

func (b *Boolean) Mutable() bool {
	return false
}

func (b *Boolean) Temporary() bool {
	return true
}

func (b *Boolean) TryToType(t hir.Type) {
	bt, ok := t.(types.BooleanType)
	if !ok {
		return
	}
	b.Type = bt
}

type String struct {
	Type  types.StringType
	Value string
}

func NewString(t types.StringType, v string) *String {
	return &String{
		Type:  t,
		Value: v,
	}
}

func (*String) expr()  {}
func (*String) local() {}

func (s *String) Print(p *hir.Printer) {
	p.WriteString("\"")
	p.WriteString(s.Value)
	p.WriteString("\"")
}

func (s *String) GetType() hir.Type {
	return s.Type
}

func (s *String) Mutable() bool {
	return false
}

func (s *String) Temporary() bool {
	return true
}

func (s *String) TryToType(t hir.Type) {
	st, ok := t.(types.StringType)
	if !ok {
		return
	}
	s.Type = st
}

type Unary interface {
	Expr
	GetOpTarget() Expr
}

type BitsReverse struct {
	Target Expr
}

func NewBitReverse(t Expr) *BitsReverse {
	return &BitsReverse{
		Target: t,
	}
}

func (*BitsReverse) expr()  {}
func (*BitsReverse) local() {}

func (e *BitsReverse) Print(p *hir.Printer) {
	p.WriteString("!")
	p.WriteBy(e.Target)
}

func (e *BitsReverse) GetType() hir.Type {
	return e.Target.GetType()
}

func (e *BitsReverse) Mutable() bool {
	return false
}

func (e *BitsReverse) Temporary() bool {
	return true
}

func (e *BitsReverse) GetOpTarget() Expr {
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

func (e *BooleanReverse) GetType() hir.Type {
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

func NewGetRef(mut bool, target Expr) *GetRef {
	return &GetRef{
		Mut:    mut,
		Target: target,
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

func (e *GetRef) GetType() hir.Type {
	return types.NewRefType(e.Mut, e.Target.GetType())
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

func (e *DeRef) GetType() hir.Type {
	return e.Target.GetType().(types.RefType).PtrTo()
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
	Shl BinaryOp `enum:"<<"`
	Shr BinaryOp `enum:">>"`

	Eq       BinaryOp `enum:"=="`
	Neq      BinaryOp `enum:"!="`
	Lt       BinaryOp `enum:"<"`
	Lte      BinaryOp `enum:"<="`
	Gt       BinaryOp `enum:">"`
	Gte      BinaryOp `enum:">="`
	LogicAnd BinaryOp `enum:"&&"`
	LogicOr  BinaryOp `enum:"||"`

	Assign    BinaryOp `enum:"="`
	AddAssign BinaryOp `enum:"+="`
	SubAssign BinaryOp `enum:"-="`
	MulAssign BinaryOp `enum:"*="`
	QuoAssign BinaryOp `enum:"/="`
	RemAssign BinaryOp `enum:"%="`
	AndAssign BinaryOp `enum:"&="`
	OrAssign  BinaryOp `enum:"|="`
	XorAssign BinaryOp `enum:"^="`
	ShlAssign BinaryOp `enum:"<<="`
	ShrAssign BinaryOp `enum:">>="`
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

func (e *Binary) GetType() hir.Type {
	switch e.Op {
	case BinaryOpEnum.Add, BinaryOpEnum.Sub, BinaryOpEnum.Mul, BinaryOpEnum.Quo, BinaryOpEnum.Rem, BinaryOpEnum.And,
		BinaryOpEnum.Or, BinaryOpEnum.Xor, BinaryOpEnum.Shl, BinaryOpEnum.Shr:
		return e.Left.GetType()
	case BinaryOpEnum.Assign, BinaryOpEnum.AddAssign, BinaryOpEnum.SubAssign, BinaryOpEnum.MulAssign,
		BinaryOpEnum.QuoAssign, BinaryOpEnum.RemAssign, BinaryOpEnum.AndAssign, BinaryOpEnum.OrAssign,
		BinaryOpEnum.XorAssign, BinaryOpEnum.ShlAssign, BinaryOpEnum.ShrAssign:
		return types.Unit
	case BinaryOpEnum.Eq, BinaryOpEnum.Neq, BinaryOpEnum.Lt, BinaryOpEnum.Lte, BinaryOpEnum.Gt, BinaryOpEnum.Gte,
		BinaryOpEnum.LogicAnd, BinaryOpEnum.LogicOr:
		return types.Bool
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
	Type   types.FuncType
	Params []*hir.Param
	Body   optional.Optional[*Block]

	CaptureVariables []hir.Ident
}

func NewFunc(t types.FuncType, params ...*hir.Param) *Func {
	return &Func{
		Type:   t,
		Params: params,
	}
}

func (*Func) expr()  {}
func (*Func) local() {}

func (f *Func) Print(p *hir.Printer) {
	p.WriteString("(")
	for i, param := range f.Params {
		p.WriteBy(param)
		if i < len(f.Params)-1 {
			p.WriteString(", ")
		}
	}
	p.WriteString(")")
	p.WriteString(" -> ")
	p.WriteBy(f.Type.GetReturn())
	if body, ok := f.Body.Value(); ok {
		p.WriteString(" ")
		p.WriteBy(body)
	}
}

func (f *Func) GetType() hir.Type {
	params := stlslices.Map(f.Params, func(_ int, param *hir.Param) hir.Type {
		return param.Type
	})
	return types.NewFuncType(f.Type.GetReturn(), params...)
}

func (f *Func) Mutable() bool {
	return false
}

func (f *Func) Temporary() bool {
	return true
}

func (f *Func) TryToType(t hir.Type) {
	ft, ok := t.(types.FuncType)
	if !ok ||
		len(ft.GetParams()) != len(f.Type.GetParams()) ||
		!ft.GetReturn().Equal(f.Type.GetReturn()) {
		return
	}
	for i, p := range ft.GetParams() {
		if !p.Equal(f.Type.GetParams()[i]) {
			return
		}
	}
	f.Type = ft
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

func (e *Call) GetType() hir.Type {
	return e.Func.GetType().(types.FuncType).GetReturn()
}

func (e *Call) Mutable() bool {
	return false
}

func (e *Call) Temporary() bool {
	return true
}

type Tuple struct {
	Type  types.TupleType
	Elems []Expr
}

func NewTuple(t types.TupleType, elems ...Expr) *Tuple {
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

func (e *Tuple) GetType() hir.Type {
	return e.Type
}

func (e *Tuple) Mutable() bool {
	return false
}

func (e *Tuple) Temporary() bool {
	return true
}

func (e *Tuple) TryToType(t hir.Type) {
	tt, ok := t.(types.TupleType)
	if !ok ||
		len(tt.GetElems()) != len(e.Type.GetElems()) {
		return
	}
	for i, et := range tt.GetElems() {
		if !et.Equal(e.Type.GetElems()[i]) {
			return
		}
	}
	e.Type = tt
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

func (e *TupleIndex) GetType() hir.Type {
	elems := e.From.GetType().(types.TupleType).GetElems()
	return elems[e.Index.Int64()]
}

func (e *TupleIndex) Mutable() bool {
	return e.From.Mutable()
}

func (e *TupleIndex) Temporary() bool {
	return e.From.Temporary()
}

type Array struct {
	Type  types.ArrayType
	Elems []Expr
}

func NewArray(t types.ArrayType, elems ...Expr) *Array {
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

func (e *Array) GetType() hir.Type {
	return e.Type
}

func (e *Array) Mutable() bool {
	return false
}

func (e *Array) Temporary() bool {
	return true
}

func (e *Array) TryToType(t hir.Type) {
	at, ok := t.(types.ArrayType)
	if !ok ||
		at.GetSize().String() != e.Type.GetSize().String() ||
		!at.GetElem().Equal(e.Type.GetElem()) {
		return
	}
	e.Type = at
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

func (e *ArrayIndex) GetType() hir.Type {
	return e.From.GetType().(types.ArrayType).GetElem()
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
}

type Union struct {
	From  Expr
	To    hir.Type
	Index uint8
}

func NewUnion(from Expr, to hir.Type, index uint8) *Union {
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

func (e *Union) GetType() hir.Type {
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
	To   hir.Type
}

func NewNumberCovert(from Expr, to hir.Type) *NumberCovert {
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

func (e *NumberCovert) GetType() hir.Type {
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

type TypedefCovert struct {
	From Expr
	To   hir.Type
}

func NewTypedefCovert(from Expr, to hir.Type) *TypedefCovert {
	return &TypedefCovert{
		From: from,
		To:   to,
	}
}

func (*TypedefCovert) expr()  {}
func (*TypedefCovert) local() {}

func (e *TypedefCovert) Print(p *hir.Printer) {
	p.WriteBy(e.From)
	p.WriteString(" as ")
	p.WriteBy(e.To)
}

func (e *TypedefCovert) GetType() hir.Type {
	return e.To
}

func (e *TypedefCovert) Mutable() bool {
	return e.From.Mutable()
}

func (e *TypedefCovert) Temporary() bool {
	return e.From.Temporary()
}

func (e *TypedefCovert) GetFrom() Expr {
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

func (e *Ternary) GetType() hir.Type {
	return e.TrueExpr.GetType()
}

func (e *Ternary) Mutable() bool {
	return e.TrueExpr.Mutable() && e.FalseExpr.Mutable()
}

func (e *Ternary) Temporary() bool {
	return e.TrueExpr.Temporary() || e.FalseExpr.Temporary()
}

type Struct struct {
	Type   types.StructType
	Fields map[string]Expr
}

func NewStruct(t types.StructType, fields map[string]Expr) *Struct {
	return &Struct{Type: t, Fields: fields}
}

func (*Struct) expr()  {}
func (*Struct) local() {}

func (e *Struct) Print(p *hir.Printer) {
	p.WriteBy(e.Type)
	p.WriteString("{")
	if len(e.Fields) > 0 {
		p.NextLine(+1)
		var i int
		for fn, fv := range e.Fields {
			p.WriteString(fn)
			p.WriteString(": ")
			p.WriteBy(fv)
			p.WriteString(",")
			if i < len(e.Fields)-1 {
				p.NextLine()
			}
			i++
		}
		p.NextLine(-1)
	}
	p.WriteString("}")
}

func (e *Struct) GetType() hir.Type {
	return e.Type
}

func (e *Struct) Mutable() bool {
	return false
}

func (e *Struct) Temporary() bool {
	return true
}

type GetField struct {
	From Expr
	Name string
}

func NewGetField(from Expr, name string) *GetField {
	return &GetField{From: from, Name: name}
}

func (*GetField) expr()  {}
func (*GetField) local() {}

func (e *GetField) Print(p *hir.Printer) {
	p.WriteBy(e.From)
	p.WriteString(".")
	p.WriteString(e.Name)
}

func (e *GetField) GetType() hir.Type {
	st := e.From.GetType().(types.StructType)
	for _, f := range st.GetFields() {
		if f.Name == e.Name {
			return f.Type
		}
	}
	panic("unreachable")
}

func (e *GetField) Mutable() bool {
	st := e.From.GetType().(types.StructType)
	for _, f := range st.GetFields() {
		if f.Name == e.Name {
			return f.Mut
		}
	}
	panic("unreachable")
}

func (e *GetField) Temporary() bool {
	return e.From.Temporary()
}

type GetBind struct {
	From Expr
	Bind *Let
}

func NewGetBind(from Expr, bind *Let) *GetBind {
	return &GetBind{From: from, Bind: bind}
}

func (*GetBind) expr()  {}
func (*GetBind) local() {}

func (e *GetBind) Print(p *hir.Printer) {
	p.WriteBy(e.From)
	p.WriteString(".")
	p.WriteString(e.Bind.Name)
}

func (e *GetBind) GetType() hir.Type {
	if e.IsStatic() {
		return e.Bind.GetType()
	}
	ft := e.Bind.GetType().(types.FuncType)
	return types.NewFuncType(ft.GetReturn(), ft.GetParams()[1:]...)
}

func (e *GetBind) Mutable() bool {
	return e.Bind.Mutable()
}

func (e *GetBind) Temporary() bool {
	return false
}

func (e *GetBind) IsStatic() bool {
	if e.Bind.Value.IsNone() || e.Bind.ExternalName.IsSome() || e.Bind.Mut {
		return true
	}
	f, ok := e.Bind.Value.MustValue().(*Func)
	if !ok {
		return true
	}
	return stlslices.All(f.Params, func(_ int, p *hir.Param) bool {
		return p.Name != "self"
	})
}
