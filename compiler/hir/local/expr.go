package local

import (
	"math/big"

	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/list"

	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// Expr 表达式
type Expr struct {
	pos   *list.Element[Local]
	value values.Value
}

func NewExpr(v values.Value) *Expr {
	return &Expr{
		value: v,
	}
}

func (self *Expr) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *Expr) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
}

func (self *Expr) Value() values.Value {
	return self.value
}

// ---------------------------------TypeExpr---------------------------------

// ArrayExpr 数组
type ArrayExpr struct {
	et    types.Type
	elems []values.Value
}

func NewArrayExpr(et types.Type, e ...values.Value) *ArrayExpr {
	return &ArrayExpr{
		et:    et,
		elems: e,
	}
}
func (self *ArrayExpr) Type() types.Type {
	return types.NewArrayType(self.et, uint(len(self.elems)))
}
func (self *ArrayExpr) Mutable() bool         { return false }
func (self *ArrayExpr) Storable() bool        { return false }
func (self *ArrayExpr) Elems() []values.Value { return self.elems }

// TupleExpr 元组
type TupleExpr struct {
	elems []values.Value
}

func NewTupleExpr(e ...values.Value) *TupleExpr {
	return &TupleExpr{
		elems: e,
	}
}
func (self *TupleExpr) Type() types.Type {
	return types.NewTupleType(stlslices.Map(self.elems, func(i int, e values.Value) types.Type {
		return e.Type()
	})...)
}
func (self *TupleExpr) Mutable() bool         { return false }
func (self *TupleExpr) Storable() bool        { return false }
func (self *TupleExpr) Elems() []values.Value { return self.elems }

// StructExpr 结构体
type StructExpr struct {
	st     types.StructType
	fields []values.Value
}

func NewStructExpr(st types.StructType, e ...values.Value) *StructExpr {
	return &StructExpr{
		st:     st,
		fields: e,
	}
}
func (self *StructExpr) Type() types.Type {
	return self.st
}
func (self *StructExpr) Mutable() bool          { return false }
func (self *StructExpr) Storable() bool         { return false }
func (self *StructExpr) Fields() []values.Value { return self.fields }

// LambdaExpr 匿名函数
type LambdaExpr struct {
	ret     types.Type
	params  []*Param
	body    *Block
	context []values.Ident
}

func NewLambdaExpr(p *Block, ret types.FuncType, params []*Param, ctx ...values.Ident) *LambdaExpr {
	return &LambdaExpr{
		ret:     ret,
		params:  params,
		body:    NewBlock(p),
		context: ctx,
	}
}
func NewLambdaExprWithNoParent(ret types.FuncType, params []*Param, ctx ...values.Ident) *LambdaExpr {
	expr := &LambdaExpr{
		ret:     ret,
		params:  params,
		context: ctx,
	}
	expr.body = NewFuncBody(expr)
	return expr
}
func (self *LambdaExpr) Type() types.Type { return self.CallableType() }
func (self *LambdaExpr) Mutable() bool    { return false }
func (self *LambdaExpr) Storable() bool   { return false }
func (self *LambdaExpr) CallableType() types.CallableType {
	return types.NewLambdaType(self.ret, stlslices.Map(self.params, func(i int, p *Param) types.Type {
		return p.Type()
	})...)
}
func (self *LambdaExpr) Params() []*Param        { return self.params }
func (self *LambdaExpr) Block() (*Block, bool)   { return self.body, true }
func (self *LambdaExpr) Context() []values.Ident { return self.context }

// EnumExpr 枚举值
type EnumExpr struct {
	enum  types.EnumType
	field string
	elem  values.Value
}

func NewEnumExpr(enum types.EnumType, field string, elem ...values.Value) *EnumExpr {
	return &EnumExpr{
		enum:  enum,
		field: field,
		elem:  stlslices.Last(elem),
	}
}
func (self *EnumExpr) Type() types.Type { return self.enum }
func (self *EnumExpr) Mutable() bool    { return false }
func (self *EnumExpr) Storable() bool   { return false }

// DefaultExpr 默认值
type DefaultExpr struct {
	t types.Type
}

func NewDefaultExpr(t types.Type) *DefaultExpr { return &DefaultExpr{t: t} }
func (self *DefaultExpr) Type() types.Type     { return self.t }
func (self *DefaultExpr) Mutable() bool        { return false }
func (self *DefaultExpr) Storable() bool       { return false }

// ---------------------------------BinaryExpr---------------------------------

// BinaryExpr 双元表达式
type BinaryExpr interface {
	values.Value
	GetLeft() values.Value
	GetRight() values.Value
}

// AssignExpr 赋值
type AssignExpr struct {
	left, right values.Value
}

func NewAssignExpr(l, r values.Value) *AssignExpr { return &AssignExpr{left: l, right: r} }
func (self *AssignExpr) Type() types.Type         { return types.NoThing }
func (self *AssignExpr) Mutable() bool            { return false }
func (self *AssignExpr) Storable() bool           { return false }
func (self *AssignExpr) GetLeft() values.Value    { return self.left }
func (self *AssignExpr) GetRight() values.Value   { return self.right }

// AndExpr 与
type AndExpr struct {
	left, right values.Value
}

func NewAndExpr(l, r values.Value) *AndExpr  { return &AndExpr{left: l, right: r} }
func (self *AndExpr) Type() types.Type       { return self.left.Type() }
func (self *AndExpr) Mutable() bool          { return false }
func (self *AndExpr) Storable() bool         { return false }
func (self *AndExpr) GetLeft() values.Value  { return self.left }
func (self *AndExpr) GetRight() values.Value { return self.right }

// OrExpr 或
type OrExpr struct {
	left, right values.Value
}

func NewOrExpr(l, r values.Value) *OrExpr   { return &OrExpr{left: l, right: r} }
func (self *OrExpr) Type() types.Type       { return self.left.Type() }
func (self *OrExpr) Mutable() bool          { return false }
func (self *OrExpr) Storable() bool         { return false }
func (self *OrExpr) GetLeft() values.Value  { return self.left }
func (self *OrExpr) GetRight() values.Value { return self.right }

// XorExpr 异或
type XorExpr struct {
	left, right values.Value
}

func NewXorExpr(l, r values.Value) *XorExpr  { return &XorExpr{left: l, right: r} }
func (self *XorExpr) Type() types.Type       { return self.left.Type() }
func (self *XorExpr) Mutable() bool          { return false }
func (self *XorExpr) Storable() bool         { return false }
func (self *XorExpr) GetLeft() values.Value  { return self.left }
func (self *XorExpr) GetRight() values.Value { return self.right }

// ShlExpr 左移
type ShlExpr struct {
	left, right values.Value
}

func NewShlExpr(l, r values.Value) *ShlExpr  { return &ShlExpr{left: l, right: r} }
func (self *ShlExpr) Type() types.Type       { return self.left.Type() }
func (self *ShlExpr) Mutable() bool          { return false }
func (self *ShlExpr) Storable() bool         { return false }
func (self *ShlExpr) GetLeft() values.Value  { return self.left }
func (self *ShlExpr) GetRight() values.Value { return self.right }

// ShrExpr 右移
type ShrExpr struct {
	left, right values.Value
}

func NewShrExpr(l, r values.Value) *ShrExpr  { return &ShrExpr{left: l, right: r} }
func (self *ShrExpr) Type() types.Type       { return self.left.Type() }
func (self *ShrExpr) Mutable() bool          { return false }
func (self *ShrExpr) Storable() bool         { return false }
func (self *ShrExpr) GetLeft() values.Value  { return self.left }
func (self *ShrExpr) GetRight() values.Value { return self.right }

// AddExpr 加法
type AddExpr struct {
	left, right values.Value
}

func NewAddExpr(l, r values.Value) *AddExpr  { return &AddExpr{left: l, right: r} }
func (self *AddExpr) Type() types.Type       { return self.left.Type() }
func (self *AddExpr) Mutable() bool          { return false }
func (self *AddExpr) Storable() bool         { return false }
func (self *AddExpr) GetLeft() values.Value  { return self.left }
func (self *AddExpr) GetRight() values.Value { return self.right }

// SubExpr 减法
type SubExpr struct {
	left, right values.Value
}

func NewSubExpr(l, r values.Value) *SubExpr  { return &SubExpr{left: l, right: r} }
func (self *SubExpr) Type() types.Type       { return self.left.Type() }
func (self *SubExpr) Mutable() bool          { return false }
func (self *SubExpr) Storable() bool         { return false }
func (self *SubExpr) GetLeft() values.Value  { return self.left }
func (self *SubExpr) GetRight() values.Value { return self.right }

// MulExpr 乘法
type MulExpr struct {
	left, right values.Value
}

func NewMulExpr(l, r values.Value) *MulExpr  { return &MulExpr{left: l, right: r} }
func (self *MulExpr) Type() types.Type       { return self.left.Type() }
func (self *MulExpr) Mutable() bool          { return false }
func (self *MulExpr) Storable() bool         { return false }
func (self *MulExpr) GetLeft() values.Value  { return self.left }
func (self *MulExpr) GetRight() values.Value { return self.right }

// DivExpr 除法
type DivExpr struct {
	left, right values.Value
}

func NewDivExpr(l, r values.Value) *DivExpr  { return &DivExpr{left: l, right: r} }
func (self *DivExpr) Type() types.Type       { return self.left.Type() }
func (self *DivExpr) Mutable() bool          { return false }
func (self *DivExpr) Storable() bool         { return false }
func (self *DivExpr) GetLeft() values.Value  { return self.left }
func (self *DivExpr) GetRight() values.Value { return self.right }

// RemExpr 取余
type RemExpr struct {
	left, right values.Value
}

func NewRemExpr(l, r values.Value) *RemExpr  { return &RemExpr{left: l, right: r} }
func (self *RemExpr) Type() types.Type       { return self.left.Type() }
func (self *RemExpr) Mutable() bool          { return false }
func (self *RemExpr) Storable() bool         { return false }
func (self *RemExpr) GetLeft() values.Value  { return self.left }
func (self *RemExpr) GetRight() values.Value { return self.right }

// LtExpr 小于
type LtExpr struct {
	left, right values.Value
}

func NewLtExpr(l, r values.Value) *LtExpr   { return &LtExpr{left: l, right: r} }
func (self *LtExpr) Type() types.Type       { return types.Bool }
func (self *LtExpr) Mutable() bool          { return false }
func (self *LtExpr) Storable() bool         { return false }
func (self *LtExpr) GetLeft() values.Value  { return self.left }
func (self *LtExpr) GetRight() values.Value { return self.right }

// GtExpr 大于
type GtExpr struct {
	left, right values.Value
}

func NewGtExpr(l, r values.Value) *GtExpr   { return &GtExpr{left: l, right: r} }
func (self *GtExpr) Type() types.Type       { return types.Bool }
func (self *GtExpr) Mutable() bool          { return false }
func (self *GtExpr) Storable() bool         { return false }
func (self *GtExpr) GetLeft() values.Value  { return self.left }
func (self *GtExpr) GetRight() values.Value { return self.right }

// LeExpr 小于等于
type LeExpr struct {
	left, right values.Value
}

func NewLeExpr(l, r values.Value) *LeExpr   { return &LeExpr{left: l, right: r} }
func (self *LeExpr) Type() types.Type       { return types.Bool }
func (self *LeExpr) Mutable() bool          { return false }
func (self *LeExpr) Storable() bool         { return false }
func (self *LeExpr) GetLeft() values.Value  { return self.left }
func (self *LeExpr) GetRight() values.Value { return self.right }

// GeExpr 大于等于
type GeExpr struct {
	left, right values.Value
}

func NewGeExpr(l, r values.Value) *GeExpr   { return &GeExpr{left: l, right: r} }
func (self *GeExpr) Type() types.Type       { return types.Bool }
func (self *GeExpr) Mutable() bool          { return false }
func (self *GeExpr) Storable() bool         { return false }
func (self *GeExpr) GetLeft() values.Value  { return self.left }
func (self *GeExpr) GetRight() values.Value { return self.right }

// EqExpr 等于
type EqExpr struct {
	left, right values.Value
}

func NewEqExpr(l, r values.Value) *EqExpr   { return &EqExpr{left: l, right: r} }
func (self *EqExpr) Type() types.Type       { return types.Bool }
func (self *EqExpr) Mutable() bool          { return false }
func (self *EqExpr) Storable() bool         { return false }
func (self *EqExpr) GetLeft() values.Value  { return self.left }
func (self *EqExpr) GetRight() values.Value { return self.right }

// NeExpr 不等于
type NeExpr struct {
	left, right values.Value
}

func NewNeExpr(l, r values.Value) *NeExpr   { return &NeExpr{left: l, right: r} }
func (self *NeExpr) Type() types.Type       { return types.Bool }
func (self *NeExpr) Mutable() bool          { return false }
func (self *NeExpr) Storable() bool         { return false }
func (self *NeExpr) GetLeft() values.Value  { return self.left }
func (self *NeExpr) GetRight() values.Value { return self.right }

// LogicAndExpr 并且
type LogicAndExpr struct {
	left, right values.Value
}

func NewLogicAndExpr(l, r values.Value) *LogicAndExpr { return &LogicAndExpr{left: l, right: r} }
func (self *LogicAndExpr) Type() types.Type           { return self.left.Type() }
func (self *LogicAndExpr) Mutable() bool              { return false }
func (self *LogicAndExpr) Storable() bool             { return false }
func (self *LogicAndExpr) GetLeft() values.Value      { return self.left }
func (self *LogicAndExpr) GetRight() values.Value     { return self.right }

// LogicOrExpr 或者
type LogicOrExpr struct {
	left, right values.Value
}

func NewLogicOrExpr(l, r values.Value) *LogicOrExpr { return &LogicOrExpr{left: l, right: r} }
func (self *LogicOrExpr) Type() types.Type          { return self.left.Type() }
func (self *LogicOrExpr) Mutable() bool             { return false }
func (self *LogicOrExpr) Storable() bool            { return false }
func (self *LogicOrExpr) GetLeft() values.Value     { return self.left }
func (self *LogicOrExpr) GetRight() values.Value    { return self.right }

// IndexExpr 索引
type IndexExpr struct {
	from, index values.Value
}

func NewIndexExpr(f, i values.Value) *IndexExpr { return &IndexExpr{from: f, index: i} }
func (self *IndexExpr) Type() types.Type {
	return self.from.Type().(types.ArrayType).Elem()
}
func (self *IndexExpr) Mutable() bool          { return false }
func (self *IndexExpr) Storable() bool         { return false }
func (self *IndexExpr) GetLeft() values.Value  { return self.from }
func (self *IndexExpr) GetRight() values.Value { return self.index }

// ExtractExpr 提取
type ExtractExpr struct {
	from  values.Value
	index uint
}

func NewExtractExpr(f values.Value, i uint) *ExtractExpr { return &ExtractExpr{from: f, index: i} }
func (self *ExtractExpr) Type() types.Type {
	return self.from.Type().(types.TupleType).Fields()[self.index]
}
func (self *ExtractExpr) Mutable() bool         { return false }
func (self *ExtractExpr) Storable() bool        { return false }
func (self *ExtractExpr) GetLeft() values.Value { return self.from }
func (self *ExtractExpr) GetRight() values.Value {
	return values.NewInteger(types.Usize, big.NewInt(int64(self.index)))
}
func (self *ExtractExpr) Index() uint { return self.index }

// FieldExpr 字段
type FieldExpr struct {
	from  values.Value
	field string
}

func NewFieldExpr(f values.Value, i string) *FieldExpr { return &FieldExpr{from: f, field: i} }
func (self *FieldExpr) Type() types.Type {
	field, _ := stlslices.FindFirst(self.from.Type().(types.StructType).Fields(), func(i int, f *types.Field) bool {
		return f.Name() == self.field
	})
	return field.Type()
}
func (self *FieldExpr) Mutable() bool          { return false }
func (self *FieldExpr) Storable() bool         { return false }
func (self *FieldExpr) GetLeft() values.Value  { return self.from }
func (self *FieldExpr) GetRight() values.Value { return values.NewString(self.field) }
func (self *FieldExpr) Field() string          { return self.field }

// MethodExpr 方法
type MethodExpr struct {
	self       values.Value
	methodName string
	methodDef  CallableDef
}

func NewMethodExpr(self values.Value, methodName string, methodDef CallableDef) *MethodExpr {
	return &MethodExpr{self: self, methodName: methodName, methodDef: methodDef}
}
func (self *MethodExpr) Type() types.Type {
	ft := self.methodDef.CallableType()
	return types.NewFuncType(ft.Ret(), ft.Params()[1:]...)
}
func (self *MethodExpr) Mutable() bool             { return false }
func (self *MethodExpr) Storable() bool            { return false }
func (self *MethodExpr) GetLeft() values.Value     { return self.self }
func (self *MethodExpr) GetRight() values.Value    { return values.NewString(self.methodName) }
func (self *MethodExpr) MethodName() string        { return self.methodName }
func (self *MethodExpr) MethodDefine() CallableDef { return self.methodDef }

// ---------------------------------UnaryExpr---------------------------------

// UnaryExpr 一元表达式
type UnaryExpr interface {
	values.Value
	GetValue() values.Value
}

// OppositeExpr 取相反数
type OppositeExpr struct {
	value values.Value
}

func NewOppositeExpr(v values.Value) *OppositeExpr { return &OppositeExpr{value: v} }
func (self *OppositeExpr) Type() types.Type        { return self.value.Type() }
func (self *OppositeExpr) Mutable() bool           { return false }
func (self *OppositeExpr) Storable() bool          { return false }
func (self *OppositeExpr) GetValue() values.Value  { return self.value }

// NegExpr 按位取反
type NegExpr struct {
	value values.Value
}

func NewNegExpr(v values.Value) *NegExpr     { return &NegExpr{value: v} }
func (self *NegExpr) Type() types.Type       { return self.value.Type() }
func (self *NegExpr) Mutable() bool          { return false }
func (self *NegExpr) Storable() bool         { return false }
func (self *NegExpr) GetValue() values.Value { return self.value }

// NotExpr 否定
type NotExpr struct {
	value values.Value
}

func NewNotExpr(v values.Value) *NotExpr     { return &NotExpr{value: v} }
func (self *NotExpr) Type() types.Type       { return self.value.Type() }
func (self *NotExpr) Mutable() bool          { return false }
func (self *NotExpr) Storable() bool         { return false }
func (self *NotExpr) GetValue() values.Value { return self.value }

// GetRefExpr 取引用
type GetRefExpr struct {
	mut   bool
	value values.Value
}

func NewGetRefExpr(mut bool, v values.Value) *GetRefExpr {
	return &GetRefExpr{
		mut:   mut,
		value: v,
	}
}
func (self *GetRefExpr) Type() types.Type       { return types.NewRefType(self.mut, self.value.Type()) }
func (self *GetRefExpr) Mutable() bool          { return false }
func (self *GetRefExpr) Storable() bool         { return false }
func (self *GetRefExpr) GetValue() values.Value { return self.value }
func (self *GetRefExpr) RefMutable() bool {
	return self.mut
}

// DeRefExpr 解引用
type DeRefExpr struct {
	value values.Value
}

func NewDeRefExpr(v values.Value) *DeRefExpr { return &DeRefExpr{value: v} }
func (self *DeRefExpr) Type() types.Type     { return self.value.Type().(types.RefType).Pointer() }
func (self *DeRefExpr) Mutable() bool {
	return self.value.Mutable() && self.value.Type().(types.RefType).Mutable()
}
func (self *DeRefExpr) Storable() bool         { return true }
func (self *DeRefExpr) GetValue() values.Value { return self.value }

// ---------------------------------CovertExpr--------------------------------

// CovertExpr 转换表达式
type CovertExpr interface {
	values.Value
	GetFrom() values.Value
	GetToType() types.Type
}

// Int2IntExpr int -> int
type Int2IntExpr struct {
	from values.Value
	to   types.IntType
}

func NewInt2IntExpr(f values.Value, t types.IntType) *Int2IntExpr {
	return &Int2IntExpr{from: f, to: t}
}
func (self *Int2IntExpr) Type() types.Type      { return self.to }
func (self *Int2IntExpr) Mutable() bool         { return false }
func (self *Int2IntExpr) Storable() bool        { return false }
func (self *Int2IntExpr) GetFrom() values.Value { return self.from }
func (self *Int2IntExpr) GetToType() types.Type { return self.to }

// Int2FloatExpr int -> float
type Int2FloatExpr struct {
	from values.Value
	to   types.FloatType
}

func NewInt2FloatExpr(f values.Value, t types.FloatType) *Int2FloatExpr {
	return &Int2FloatExpr{from: f, to: t}
}
func (self *Int2FloatExpr) Type() types.Type      { return self.to }
func (self *Int2FloatExpr) Mutable() bool         { return false }
func (self *Int2FloatExpr) Storable() bool        { return false }
func (self *Int2FloatExpr) GetFrom() values.Value { return self.from }
func (self *Int2FloatExpr) GetToType() types.Type { return self.to }

// Float2IntExpr float -> int
type Float2IntExpr struct {
	from values.Value
	to   types.IntType
}

func NewFloat2IntExpr(f values.Value, t types.IntType) *Float2IntExpr {
	return &Float2IntExpr{from: f, to: t}
}
func (self *Float2IntExpr) Type() types.Type      { return self.to }
func (self *Float2IntExpr) Mutable() bool         { return false }
func (self *Float2IntExpr) Storable() bool        { return false }
func (self *Float2IntExpr) GetFrom() values.Value { return self.from }
func (self *Float2IntExpr) GetToType() types.Type { return self.to }

// Float2FloatExpr float -> float
type Float2FloatExpr struct {
	from values.Value
	to   types.FloatType
}

func NewFloat2FloatExpr(f values.Value, t types.FloatType) *Float2FloatExpr {
	return &Float2FloatExpr{from: f, to: t}
}
func (self *Float2FloatExpr) Type() types.Type      { return self.to }
func (self *Float2FloatExpr) Mutable() bool         { return false }
func (self *Float2FloatExpr) Storable() bool        { return false }
func (self *Float2FloatExpr) GetFrom() values.Value { return self.from }
func (self *Float2FloatExpr) GetToType() types.Type { return self.to }

// WrapTypeExpr custom -> custom
type WrapTypeExpr struct {
	from values.Value
	to   types.Type
}

func NewWrapTypeExpr(f values.Value, t types.Type) *WrapTypeExpr {
	return &WrapTypeExpr{from: f, to: t}
}
func (self *WrapTypeExpr) Type() types.Type      { return self.to }
func (self *WrapTypeExpr) Mutable() bool         { return self.from.Mutable() }
func (self *WrapTypeExpr) Storable() bool        { return self.from.Storable() }
func (self *WrapTypeExpr) GetFrom() values.Value { return self.from }
func (self *WrapTypeExpr) GetToType() types.Type { return self.to }

// NoReturn2AnyExpr X -> any
type NoReturn2AnyExpr struct {
	from values.Value
	to   types.Type
}

func NewNoReturn2AnyExpr(f values.Value, t types.Type) *NoReturn2AnyExpr {
	return &NoReturn2AnyExpr{from: f, to: t}
}
func (self *NoReturn2AnyExpr) Type() types.Type      { return self.to }
func (self *NoReturn2AnyExpr) Mutable() bool         { return false }
func (self *NoReturn2AnyExpr) Storable() bool        { return false }
func (self *NoReturn2AnyExpr) GetFrom() values.Value { return self.from }
func (self *NoReturn2AnyExpr) GetToType() types.Type { return self.to }

// Func2LambdaExpr func -> lambda
type Func2LambdaExpr struct {
	from values.Value
	to   types.LambdaType
}

func NewFunc2LambdaExpr(f values.Value, t types.LambdaType) *Func2LambdaExpr {
	return &Func2LambdaExpr{from: f, to: t}
}
func (self *Func2LambdaExpr) Type() types.Type      { return self.to }
func (self *Func2LambdaExpr) Mutable() bool         { return false }
func (self *Func2LambdaExpr) Storable() bool        { return false }
func (self *Func2LambdaExpr) GetFrom() values.Value { return self.from }
func (self *Func2LambdaExpr) GetToType() types.Type { return self.to }

// Enum2NumberExpr enum -> num
type Enum2NumberExpr struct {
	from values.Value
	to   types.NumType
}

func NewEnum2NumberExpr(f values.Value, t types.NumType) *Enum2NumberExpr {
	return &Enum2NumberExpr{from: f, to: t}
}
func (self *Enum2NumberExpr) Type() types.Type      { return self.to }
func (self *Enum2NumberExpr) Mutable() bool         { return false }
func (self *Enum2NumberExpr) Storable() bool        { return false }
func (self *Enum2NumberExpr) GetFrom() values.Value { return self.from }
func (self *Enum2NumberExpr) GetToType() types.Type { return self.to }

// Number2EnumExpr num -> enum
type Number2EnumExpr struct {
	from values.Value
	to   types.EnumType
}

func NewNumber2EnumExpr(f values.Value, t types.EnumType) *Number2EnumExpr {
	return &Number2EnumExpr{from: f, to: t}
}
func (self *Number2EnumExpr) Type() types.Type      { return self.to }
func (self *Number2EnumExpr) Mutable() bool         { return false }
func (self *Number2EnumExpr) Storable() bool        { return false }
func (self *Number2EnumExpr) GetFrom() values.Value { return self.from }
func (self *Number2EnumExpr) GetToType() types.Type { return self.to }

// ---------------------------------OtherExpr---------------------------------

// CallExpr 调用函数
type CallExpr struct {
	fn   values.Value
	args []values.Value
}

func NewCallExpr(fn values.Value, args ...values.Value) *CallExpr {
	return &CallExpr{
		fn:   fn,
		args: args,
	}
}
func (self *CallExpr) Type() types.Type {
	return self.fn.Type().(types.CallableType).Ret()
}
func (self *CallExpr) Mutable() bool           { return false }
func (self *CallExpr) Storable() bool          { return false }
func (self *CallExpr) GetFunc() values.Value   { return self.fn }
func (self *CallExpr) GetArgs() []values.Value { return self.args }

// TypeJudgmentExpr 类型判断
type TypeJudgmentExpr struct {
	value  values.Value
	target types.Type
}

func NewTypeJudgmentExpr(v values.Value, t types.Type) *TypeJudgmentExpr {
	return &TypeJudgmentExpr{
		value:  v,
		target: t,
	}
}
func (self *TypeJudgmentExpr) Type() types.Type {
	return types.Bool
}
func (self *TypeJudgmentExpr) Mutable() bool       { return false }
func (self *TypeJudgmentExpr) Storable() bool      { return false }
func (self *TypeJudgmentExpr) Value() values.Value { return self.value }
func (self *TypeJudgmentExpr) Target() types.Type  { return self.target }
