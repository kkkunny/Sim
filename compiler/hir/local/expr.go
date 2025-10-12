package local

import (
	"math/big"

	"github.com/kkkunny/stl/container/hashmap"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

type Ident interface {
	hir.Local
	values.Ident
}

// Expr 表达式
type Expr struct {
	value hir.Value
}

func NewExpr(v hir.Value) *Expr {
	return &Expr{
		value: v,
	}
}

func (self *Expr) Local() {
	return
}

func (self *Expr) Value() hir.Value {
	return self.value
}

// ---------------------------------TypeExpr---------------------------------

// ArrayExpr 数组
type ArrayExpr struct {
	typ   types.ArrayType
	elems []hir.Value
}

func NewArrayExpr(t types.ArrayType, e ...hir.Value) *ArrayExpr {
	return &ArrayExpr{
		typ:   t,
		elems: e,
	}
}
func (self *ArrayExpr) Type() hir.Type {
	return self.typ
}
func (self *ArrayExpr) Mutable() bool      { return false }
func (self *ArrayExpr) Storable() bool     { return false }
func (self *ArrayExpr) Elems() []hir.Value { return self.elems }

// TupleExpr 元组
type TupleExpr struct {
	elems []hir.Value
}

func NewTupleExpr(e ...hir.Value) *TupleExpr {
	return &TupleExpr{
		elems: e,
	}
}
func (self *TupleExpr) Type() hir.Type {
	return types.NewTupleType(stlslices.Map(self.elems, func(i int, e hir.Value) hir.Type {
		return e.Type()
	})...)
}
func (self *TupleExpr) Mutable() bool      { return false }
func (self *TupleExpr) Storable() bool     { return false }
func (self *TupleExpr) Elems() []hir.Value { return self.elems }

// StructExpr 结构体
type StructExpr struct {
	st     types.StructType
	fields []hir.Value
}

func NewStructExpr(st types.StructType, e ...hir.Value) *StructExpr {
	return &StructExpr{
		st:     st,
		fields: e,
	}
}
func (self *StructExpr) Type() hir.Type {
	return self.st
}
func (self *StructExpr) Mutable() bool       { return false }
func (self *StructExpr) Storable() bool      { return false }
func (self *StructExpr) Fields() []hir.Value { return self.fields }

// LambdaExpr 匿名函数
type LambdaExpr struct {
	parent  Scope
	typ     types.LambdaType
	params  []*Param
	body    *Block
	context []values.Ident

	onCapture func(ident any)
}

func NewLambdaExpr(env Scope, typ types.LambdaType, params []*Param, onCapture ...func(ident any)) *LambdaExpr {
	return &LambdaExpr{
		parent:    env,
		typ:       typ,
		params:    params,
		onCapture: stlslices.Last(onCapture),
	}
}
func (self *LambdaExpr) Type() hir.Type { return self.typ }
func (self *LambdaExpr) Mutable() bool  { return false }
func (self *LambdaExpr) Storable() bool { return false }
func (self *LambdaExpr) CallableType() types.CallableType {
	return self.typ
}
func (self *LambdaExpr) Params() []*Param     { return self.params }
func (self *LambdaExpr) SetBody(b *Block)     { self.body = b }
func (self *LambdaExpr) Body() (*Block, bool) { return self.body, true }
func (self *LambdaExpr) Parent() Scope {
	return self.parent
}
func (self *LambdaExpr) GetName() (string, bool)        { return "", false }
func (self *LambdaExpr) SetContext(ctx ...values.Ident) { self.context = ctx }
func (self *LambdaExpr) Context() []values.Ident        { return self.context }

// EnumExpr 枚举值
type EnumExpr struct {
	enum  types.EnumType
	field string
	elem  hir.Value
}

func NewEnumExpr(enum types.EnumType, field string, elem ...hir.Value) *EnumExpr {
	return &EnumExpr{
		enum:  enum,
		field: field,
		elem:  stlslices.Last(elem),
	}
}
func (self *EnumExpr) Type() hir.Type { return self.enum }
func (self *EnumExpr) Mutable() bool  { return false }
func (self *EnumExpr) Storable() bool { return false }
func (self *EnumExpr) Field() string {
	return self.field
}
func (self *EnumExpr) Elem() (hir.Value, bool) {
	return self.elem, self.elem != nil
}

// DefaultExpr 默认值
type DefaultExpr struct {
	t hir.Type
}

func NewDefaultExpr(t hir.Type) *DefaultExpr { return &DefaultExpr{t: t} }
func (self *DefaultExpr) Type() hir.Type     { return self.t }
func (self *DefaultExpr) Mutable() bool      { return false }
func (self *DefaultExpr) Storable() bool     { return false }

// ---------------------------------BinaryExpr---------------------------------

// BinaryExpr 双元表达式
type BinaryExpr interface {
	hir.Value
	GetLeft() hir.Value
	GetRight() hir.Value
}

// AssignExpr 赋值
type AssignExpr struct {
	left, right hir.Value
}

func NewAssignExpr(l, r hir.Value) *AssignExpr { return &AssignExpr{left: l, right: r} }
func (self *AssignExpr) Type() hir.Type        { return types.NoThing }
func (self *AssignExpr) Mutable() bool         { return false }
func (self *AssignExpr) Storable() bool        { return false }
func (self *AssignExpr) GetLeft() hir.Value    { return self.left }
func (self *AssignExpr) GetRight() hir.Value   { return self.right }

// AndExpr 与
type AndExpr struct {
	left, right hir.Value
}

func NewAndExpr(l, r hir.Value) *AndExpr  { return &AndExpr{left: l, right: r} }
func (self *AndExpr) Type() hir.Type      { return self.left.Type() }
func (self *AndExpr) Mutable() bool       { return false }
func (self *AndExpr) Storable() bool      { return false }
func (self *AndExpr) GetLeft() hir.Value  { return self.left }
func (self *AndExpr) GetRight() hir.Value { return self.right }

// OrExpr 或
type OrExpr struct {
	left, right hir.Value
}

func NewOrExpr(l, r hir.Value) *OrExpr   { return &OrExpr{left: l, right: r} }
func (self *OrExpr) Type() hir.Type      { return self.left.Type() }
func (self *OrExpr) Mutable() bool       { return false }
func (self *OrExpr) Storable() bool      { return false }
func (self *OrExpr) GetLeft() hir.Value  { return self.left }
func (self *OrExpr) GetRight() hir.Value { return self.right }

// XorExpr 异或
type XorExpr struct {
	left, right hir.Value
}

func NewXorExpr(l, r hir.Value) *XorExpr  { return &XorExpr{left: l, right: r} }
func (self *XorExpr) Type() hir.Type      { return self.left.Type() }
func (self *XorExpr) Mutable() bool       { return false }
func (self *XorExpr) Storable() bool      { return false }
func (self *XorExpr) GetLeft() hir.Value  { return self.left }
func (self *XorExpr) GetRight() hir.Value { return self.right }

// ShlExpr 左移
type ShlExpr struct {
	left, right hir.Value
}

func NewShlExpr(l, r hir.Value) *ShlExpr  { return &ShlExpr{left: l, right: r} }
func (self *ShlExpr) Type() hir.Type      { return self.left.Type() }
func (self *ShlExpr) Mutable() bool       { return false }
func (self *ShlExpr) Storable() bool      { return false }
func (self *ShlExpr) GetLeft() hir.Value  { return self.left }
func (self *ShlExpr) GetRight() hir.Value { return self.right }

// ShrExpr 右移
type ShrExpr struct {
	left, right hir.Value
}

func NewShrExpr(l, r hir.Value) *ShrExpr  { return &ShrExpr{left: l, right: r} }
func (self *ShrExpr) Type() hir.Type      { return self.left.Type() }
func (self *ShrExpr) Mutable() bool       { return false }
func (self *ShrExpr) Storable() bool      { return false }
func (self *ShrExpr) GetLeft() hir.Value  { return self.left }
func (self *ShrExpr) GetRight() hir.Value { return self.right }

// AddExpr 加法
type AddExpr struct {
	left, right hir.Value
}

func NewAddExpr(l, r hir.Value) *AddExpr  { return &AddExpr{left: l, right: r} }
func (self *AddExpr) Type() hir.Type      { return self.left.Type() }
func (self *AddExpr) Mutable() bool       { return false }
func (self *AddExpr) Storable() bool      { return false }
func (self *AddExpr) GetLeft() hir.Value  { return self.left }
func (self *AddExpr) GetRight() hir.Value { return self.right }

// SubExpr 减法
type SubExpr struct {
	left, right hir.Value
}

func NewSubExpr(l, r hir.Value) *SubExpr  { return &SubExpr{left: l, right: r} }
func (self *SubExpr) Type() hir.Type      { return self.left.Type() }
func (self *SubExpr) Mutable() bool       { return false }
func (self *SubExpr) Storable() bool      { return false }
func (self *SubExpr) GetLeft() hir.Value  { return self.left }
func (self *SubExpr) GetRight() hir.Value { return self.right }

// MulExpr 乘法
type MulExpr struct {
	left, right hir.Value
}

func NewMulExpr(l, r hir.Value) *MulExpr  { return &MulExpr{left: l, right: r} }
func (self *MulExpr) Type() hir.Type      { return self.left.Type() }
func (self *MulExpr) Mutable() bool       { return false }
func (self *MulExpr) Storable() bool      { return false }
func (self *MulExpr) GetLeft() hir.Value  { return self.left }
func (self *MulExpr) GetRight() hir.Value { return self.right }

// DivExpr 除法
type DivExpr struct {
	left, right hir.Value
}

func NewDivExpr(l, r hir.Value) *DivExpr  { return &DivExpr{left: l, right: r} }
func (self *DivExpr) Type() hir.Type      { return self.left.Type() }
func (self *DivExpr) Mutable() bool       { return false }
func (self *DivExpr) Storable() bool      { return false }
func (self *DivExpr) GetLeft() hir.Value  { return self.left }
func (self *DivExpr) GetRight() hir.Value { return self.right }

// RemExpr 取余
type RemExpr struct {
	left, right hir.Value
}

func NewRemExpr(l, r hir.Value) *RemExpr  { return &RemExpr{left: l, right: r} }
func (self *RemExpr) Type() hir.Type      { return self.left.Type() }
func (self *RemExpr) Mutable() bool       { return false }
func (self *RemExpr) Storable() bool      { return false }
func (self *RemExpr) GetLeft() hir.Value  { return self.left }
func (self *RemExpr) GetRight() hir.Value { return self.right }

// LtExpr 小于
type LtExpr struct {
	left, right hir.Value
}

func NewLtExpr(l, r hir.Value) *LtExpr   { return &LtExpr{left: l, right: r} }
func (self *LtExpr) Type() hir.Type      { return types.Bool }
func (self *LtExpr) Mutable() bool       { return false }
func (self *LtExpr) Storable() bool      { return false }
func (self *LtExpr) GetLeft() hir.Value  { return self.left }
func (self *LtExpr) GetRight() hir.Value { return self.right }

// GtExpr 大于
type GtExpr struct {
	left, right hir.Value
}

func NewGtExpr(l, r hir.Value) *GtExpr   { return &GtExpr{left: l, right: r} }
func (self *GtExpr) Type() hir.Type      { return types.Bool }
func (self *GtExpr) Mutable() bool       { return false }
func (self *GtExpr) Storable() bool      { return false }
func (self *GtExpr) GetLeft() hir.Value  { return self.left }
func (self *GtExpr) GetRight() hir.Value { return self.right }

// LeExpr 小于等于
type LeExpr struct {
	left, right hir.Value
}

func NewLeExpr(l, r hir.Value) *LeExpr   { return &LeExpr{left: l, right: r} }
func (self *LeExpr) Type() hir.Type      { return types.Bool }
func (self *LeExpr) Mutable() bool       { return false }
func (self *LeExpr) Storable() bool      { return false }
func (self *LeExpr) GetLeft() hir.Value  { return self.left }
func (self *LeExpr) GetRight() hir.Value { return self.right }

// GeExpr 大于等于
type GeExpr struct {
	left, right hir.Value
}

func NewGeExpr(l, r hir.Value) *GeExpr   { return &GeExpr{left: l, right: r} }
func (self *GeExpr) Type() hir.Type      { return types.Bool }
func (self *GeExpr) Mutable() bool       { return false }
func (self *GeExpr) Storable() bool      { return false }
func (self *GeExpr) GetLeft() hir.Value  { return self.left }
func (self *GeExpr) GetRight() hir.Value { return self.right }

// EqExpr 等于
type EqExpr struct {
	left, right hir.Value
}

func NewEqExpr(l, r hir.Value) *EqExpr   { return &EqExpr{left: l, right: r} }
func (self *EqExpr) Type() hir.Type      { return types.Bool }
func (self *EqExpr) Mutable() bool       { return false }
func (self *EqExpr) Storable() bool      { return false }
func (self *EqExpr) GetLeft() hir.Value  { return self.left }
func (self *EqExpr) GetRight() hir.Value { return self.right }

// NeExpr 不等于
type NeExpr struct {
	left, right hir.Value
}

func NewNeExpr(l, r hir.Value) *NeExpr   { return &NeExpr{left: l, right: r} }
func (self *NeExpr) Type() hir.Type      { return types.Bool }
func (self *NeExpr) Mutable() bool       { return false }
func (self *NeExpr) Storable() bool      { return false }
func (self *NeExpr) GetLeft() hir.Value  { return self.left }
func (self *NeExpr) GetRight() hir.Value { return self.right }

// LogicAndExpr 并且
type LogicAndExpr struct {
	left, right hir.Value
}

func NewLogicAndExpr(l, r hir.Value) *LogicAndExpr { return &LogicAndExpr{left: l, right: r} }
func (self *LogicAndExpr) Type() hir.Type          { return self.left.Type() }
func (self *LogicAndExpr) Mutable() bool           { return false }
func (self *LogicAndExpr) Storable() bool          { return false }
func (self *LogicAndExpr) GetLeft() hir.Value      { return self.left }
func (self *LogicAndExpr) GetRight() hir.Value     { return self.right }

// LogicOrExpr 或者
type LogicOrExpr struct {
	left, right hir.Value
}

func NewLogicOrExpr(l, r hir.Value) *LogicOrExpr { return &LogicOrExpr{left: l, right: r} }
func (self *LogicOrExpr) Type() hir.Type         { return self.left.Type() }
func (self *LogicOrExpr) Mutable() bool          { return false }
func (self *LogicOrExpr) Storable() bool         { return false }
func (self *LogicOrExpr) GetLeft() hir.Value     { return self.left }
func (self *LogicOrExpr) GetRight() hir.Value    { return self.right }

// IndexExpr 索引
type IndexExpr struct {
	from, index hir.Value
}

func NewIndexExpr(f, i hir.Value) *IndexExpr { return &IndexExpr{from: f, index: i} }
func (self *IndexExpr) Type() hir.Type {
	return self.from.Type().(types.ArrayType).Elem()
}
func (self *IndexExpr) Mutable() bool       { return false }
func (self *IndexExpr) Storable() bool      { return false }
func (self *IndexExpr) GetLeft() hir.Value  { return self.from }
func (self *IndexExpr) GetRight() hir.Value { return self.index }

// ExtractExpr 提取
type ExtractExpr struct {
	from  hir.Value
	index uint
}

func NewExtractExpr(f hir.Value, i uint) *ExtractExpr { return &ExtractExpr{from: f, index: i} }
func (self *ExtractExpr) Type() hir.Type {
	return self.from.Type().(types.TupleType).Elems()[self.index]
}
func (self *ExtractExpr) Mutable() bool      { return false }
func (self *ExtractExpr) Storable() bool     { return false }
func (self *ExtractExpr) GetLeft() hir.Value { return self.from }
func (self *ExtractExpr) GetRight() hir.Value {
	return values.NewInteger(types.Usize, big.NewInt(int64(self.index)))
}
func (self *ExtractExpr) Index() uint { return self.index }

// FieldExpr 字段
type FieldExpr struct {
	from  hir.Value
	field string
}

func NewFieldExpr(f hir.Value, i string) *FieldExpr { return &FieldExpr{from: f, field: i} }
func (self *FieldExpr) Type() hir.Type {
	field, _ := stlslices.FindFirst(stlval.IgnoreWith(types.As[types.StructType](self.from.Type())).Fields().Values(), func(i int, f *types.Field) bool {
		return f.Name() == self.field
	})
	return field.Type()
}
func (self *FieldExpr) Mutable() bool       { return self.from.Mutable() }
func (self *FieldExpr) Storable() bool      { return self.from.Storable() }
func (self *FieldExpr) GetLeft() hir.Value  { return self.from }
func (self *FieldExpr) GetRight() hir.Value { return values.NewString(types.Str, self.field) }
func (self *FieldExpr) Field() string       { return self.field }

// MethodExpr 方法
type MethodExpr struct {
	self        hir.Value
	method      CallableDef
	genericArgs []hir.Type
}

func NewMethodExpr(self hir.Value, method CallableDef, genericArgs []hir.Type) *MethodExpr {
	return &MethodExpr{self: self, method: method, genericArgs: genericArgs}
}
func (self *MethodExpr) Type() hir.Type     { return self.CallableType() }
func (self *MethodExpr) Mutable() bool      { return false }
func (self *MethodExpr) Storable() bool     { return false }
func (self *MethodExpr) GetLeft() hir.Value { return self.self }
func (self *MethodExpr) GetRight() hir.Value {
	return values.NewString(types.Str, stlval.IgnoreWith(self.method.GetName()))
}
func (self *MethodExpr) Method() CallableDef { return self.method }
func (self *MethodExpr) GenericArgs() []hir.Type {
	return self.genericArgs
}
func (self *MethodExpr) GenericParamMap() hashmap.HashMap[types.VirtualType, hir.Type] {
	type getCompiler interface {
		GenericParams() []types.GenericParamType
	}
	var i int
	genericParamMap := hashmap.AnyWith[types.VirtualType, hir.Type](stlslices.FlatMap(self.method.(getCompiler).GenericParams(), func(_ int, compileParam types.GenericParamType) []any {
		i++
		return []any{types.VirtualType(compileParam), self.genericArgs[i-1]}
	})...)

	if t, ok := types.As[types.RefType](self.self.Type(), true); ok {
		if t, ok := types.As[types.GenericCustomType](t.Pointer(), true); ok {
			for iter := t.GenericParamMap().Iterator(); iter.Next(); {
				genericParamMap.Set(iter.Value().Unpack())
			}
		}
	} else if t, ok := types.As[types.GenericCustomType](self.self.Type(), true); ok {
		for iter := t.GenericParamMap().Iterator(); iter.Next(); {
			genericParamMap.Set(iter.Value().Unpack())
		}
	}
	return genericParamMap
}
func (self *MethodExpr) CallableType() types.CallableType {
	ft := self.method.CallableType()
	ft = types.NewLambdaType(ft.Ret(), ft.Params()[1:]...)

	genericParamMap := self.GenericParamMap()
	if genericParamMap.Empty() {
		return ft
	} else {
		return types.ReplaceVirtualType(genericParamMap, ft).(types.CallableType)
	}
}

// ---------------------------------UnaryExpr---------------------------------

// UnaryExpr 一元表达式
type UnaryExpr interface {
	hir.Value
	GetValue() hir.Value
}

// NegExpr -
type NegExpr struct {
	value hir.Value
}

func NewNegExpr(v hir.Value) *NegExpr     { return &NegExpr{value: v} }
func (self *NegExpr) Type() hir.Type      { return self.value.Type() }
func (self *NegExpr) Mutable() bool       { return false }
func (self *NegExpr) Storable() bool      { return false }
func (self *NegExpr) GetValue() hir.Value { return self.value }

// NotExpr !
type NotExpr struct {
	value hir.Value
}

func NewNotExpr(v hir.Value) *NotExpr     { return &NotExpr{value: v} }
func (self *NotExpr) Type() hir.Type      { return self.value.Type() }
func (self *NotExpr) Mutable() bool       { return false }
func (self *NotExpr) Storable() bool      { return false }
func (self *NotExpr) GetValue() hir.Value { return self.value }

// GetRefExpr 取引用
type GetRefExpr struct {
	mut   bool
	value hir.Value
}

func NewGetRefExpr(mut bool, v hir.Value) *GetRefExpr {
	return &GetRefExpr{
		mut:   mut,
		value: v,
	}
}
func (self *GetRefExpr) Type() hir.Type      { return types.NewRefType(self.mut, self.value.Type()) }
func (self *GetRefExpr) Mutable() bool       { return false }
func (self *GetRefExpr) Storable() bool      { return false }
func (self *GetRefExpr) GetValue() hir.Value { return self.value }
func (self *GetRefExpr) RefMutable() bool {
	return self.mut
}

// DeRefExpr 解引用
type DeRefExpr struct {
	value hir.Value
}

func NewDeRefExpr(v hir.Value) *DeRefExpr { return &DeRefExpr{value: v} }
func (self *DeRefExpr) Type() hir.Type    { return self.value.Type().(types.RefType).Pointer() }
func (self *DeRefExpr) Mutable() bool {
	return self.value.Mutable() && self.value.Type().(types.RefType).Mutable()
}
func (self *DeRefExpr) Storable() bool      { return true }
func (self *DeRefExpr) GetValue() hir.Value { return self.value }

// ---------------------------------CovertExpr--------------------------------

// CovertExpr 转换表达式
type CovertExpr interface {
	hir.Value
	GetFrom() hir.Value
	GetToType() hir.Type
}

// Int2IntExpr int -> int
type Int2IntExpr struct {
	from hir.Value
	to   types.IntType
}

func NewInt2IntExpr(f hir.Value, t types.IntType) *Int2IntExpr {
	return &Int2IntExpr{from: f, to: t}
}
func (self *Int2IntExpr) Type() hir.Type      { return self.to }
func (self *Int2IntExpr) Mutable() bool       { return false }
func (self *Int2IntExpr) Storable() bool      { return false }
func (self *Int2IntExpr) GetFrom() hir.Value  { return self.from }
func (self *Int2IntExpr) GetToType() hir.Type { return self.to }

// Int2FloatExpr int -> float
type Int2FloatExpr struct {
	from hir.Value
	to   types.FloatType
}

func NewInt2FloatExpr(f hir.Value, t types.FloatType) *Int2FloatExpr {
	return &Int2FloatExpr{from: f, to: t}
}
func (self *Int2FloatExpr) Type() hir.Type      { return self.to }
func (self *Int2FloatExpr) Mutable() bool       { return false }
func (self *Int2FloatExpr) Storable() bool      { return false }
func (self *Int2FloatExpr) GetFrom() hir.Value  { return self.from }
func (self *Int2FloatExpr) GetToType() hir.Type { return self.to }

// Float2IntExpr float -> int
type Float2IntExpr struct {
	from hir.Value
	to   types.IntType
}

func NewFloat2IntExpr(f hir.Value, t types.IntType) *Float2IntExpr {
	return &Float2IntExpr{from: f, to: t}
}
func (self *Float2IntExpr) Type() hir.Type      { return self.to }
func (self *Float2IntExpr) Mutable() bool       { return false }
func (self *Float2IntExpr) Storable() bool      { return false }
func (self *Float2IntExpr) GetFrom() hir.Value  { return self.from }
func (self *Float2IntExpr) GetToType() hir.Type { return self.to }

// Float2FloatExpr float -> float
type Float2FloatExpr struct {
	from hir.Value
	to   types.FloatType
}

func NewFloat2FloatExpr(f hir.Value, t types.FloatType) *Float2FloatExpr {
	return &Float2FloatExpr{from: f, to: t}
}
func (self *Float2FloatExpr) Type() hir.Type      { return self.to }
func (self *Float2FloatExpr) Mutable() bool       { return false }
func (self *Float2FloatExpr) Storable() bool      { return false }
func (self *Float2FloatExpr) GetFrom() hir.Value  { return self.from }
func (self *Float2FloatExpr) GetToType() hir.Type { return self.to }

// Ref2UsizeExpr ref -> usize
type Ref2UsizeExpr struct {
	from hir.Value
}

func NewRef2UsizeExpr(f hir.Value) *Ref2UsizeExpr {
	return &Ref2UsizeExpr{from: f}
}
func (self *Ref2UsizeExpr) Type() hir.Type      { return self.GetToType() }
func (self *Ref2UsizeExpr) Mutable() bool       { return false }
func (self *Ref2UsizeExpr) Storable() bool      { return false }
func (self *Ref2UsizeExpr) GetFrom() hir.Value  { return self.from }
func (self *Ref2UsizeExpr) GetToType() hir.Type { return types.Usize }

// Usize2RefExpr usize -> ref
type Usize2RefExpr struct {
	from hir.Value
	to   types.RefType
}

func NewUsize2RefExpr(f hir.Value, t types.RefType) *Usize2RefExpr {
	return &Usize2RefExpr{from: f, to: t}
}
func (self *Usize2RefExpr) Type() hir.Type      { return self.to }
func (self *Usize2RefExpr) Mutable() bool       { return false }
func (self *Usize2RefExpr) Storable() bool      { return false }
func (self *Usize2RefExpr) GetFrom() hir.Value  { return self.from }
func (self *Usize2RefExpr) GetToType() hir.Type { return self.to }

// WrapTypeExpr custom -> custom
type WrapTypeExpr struct {
	from hir.Value
	to   hir.Type
}

func NewWrapTypeExpr(f hir.Value, t hir.Type) *WrapTypeExpr {
	return &WrapTypeExpr{from: f, to: t}
}
func (self *WrapTypeExpr) Type() hir.Type      { return self.to }
func (self *WrapTypeExpr) Mutable() bool       { return self.from.Mutable() }
func (self *WrapTypeExpr) Storable() bool      { return self.from.Storable() }
func (self *WrapTypeExpr) GetFrom() hir.Value  { return self.from }
func (self *WrapTypeExpr) GetToType() hir.Type { return self.to }

// NoReturn2AnyExpr X -> any
type NoReturn2AnyExpr struct {
	from hir.Value
	to   hir.Type
}

func NewNoReturn2AnyExpr(f hir.Value, t hir.Type) *NoReturn2AnyExpr {
	return &NoReturn2AnyExpr{from: f, to: t}
}
func (self *NoReturn2AnyExpr) Type() hir.Type      { return self.to }
func (self *NoReturn2AnyExpr) Mutable() bool       { return false }
func (self *NoReturn2AnyExpr) Storable() bool      { return false }
func (self *NoReturn2AnyExpr) GetFrom() hir.Value  { return self.from }
func (self *NoReturn2AnyExpr) GetToType() hir.Type { return self.to }

// Func2LambdaExpr func -> lambda
type Func2LambdaExpr struct {
	from hir.Value
	to   types.LambdaType
}

func NewFunc2LambdaExpr(f hir.Value, t types.LambdaType) *Func2LambdaExpr {
	return &Func2LambdaExpr{from: f, to: t}
}
func (self *Func2LambdaExpr) Type() hir.Type      { return self.to }
func (self *Func2LambdaExpr) Mutable() bool       { return false }
func (self *Func2LambdaExpr) Storable() bool      { return false }
func (self *Func2LambdaExpr) GetFrom() hir.Value  { return self.from }
func (self *Func2LambdaExpr) GetToType() hir.Type { return self.to }

// Enum2NumberExpr enum -> num
type Enum2NumberExpr struct {
	from hir.Value
	to   types.NumType
}

func NewEnum2NumberExpr(f hir.Value, t types.NumType) *Enum2NumberExpr {
	return &Enum2NumberExpr{from: f, to: t}
}
func (self *Enum2NumberExpr) Type() hir.Type      { return self.to }
func (self *Enum2NumberExpr) Mutable() bool       { return false }
func (self *Enum2NumberExpr) Storable() bool      { return false }
func (self *Enum2NumberExpr) GetFrom() hir.Value  { return self.from }
func (self *Enum2NumberExpr) GetToType() hir.Type { return self.to }

// Number2EnumExpr num -> enum
type Number2EnumExpr struct {
	from hir.Value
	to   types.EnumType
}

func NewNumber2EnumExpr(f hir.Value, t types.EnumType) *Number2EnumExpr {
	return &Number2EnumExpr{from: f, to: t}
}
func (self *Number2EnumExpr) Type() hir.Type      { return self.to }
func (self *Number2EnumExpr) Mutable() bool       { return false }
func (self *Number2EnumExpr) Storable() bool      { return false }
func (self *Number2EnumExpr) GetFrom() hir.Value  { return self.from }
func (self *Number2EnumExpr) GetToType() hir.Type { return self.to }

// ---------------------------------OtherExpr---------------------------------

// CallExpr 调用函数
type CallExpr struct {
	fn   hir.Value
	args []hir.Value
}

func NewCallExpr(fn hir.Value, args ...hir.Value) *CallExpr {
	return &CallExpr{
		fn:   fn,
		args: args,
	}
}
func (self *CallExpr) Type() hir.Type {
	return self.fn.Type().(types.CallableType).Ret()
}
func (self *CallExpr) Mutable() bool        { return false }
func (self *CallExpr) Storable() bool       { return false }
func (self *CallExpr) GetFunc() hir.Value   { return self.fn }
func (self *CallExpr) GetArgs() []hir.Value { return self.args }

// TypeJudgmentExpr 类型判断
type TypeJudgmentExpr struct {
	value  hir.Value
	target hir.Type
}

func NewTypeJudgmentExpr(v hir.Value, t hir.Type) *TypeJudgmentExpr {
	return &TypeJudgmentExpr{
		value:  v,
		target: t,
	}
}
func (self *TypeJudgmentExpr) Type() hir.Type {
	return types.Bool
}
func (self *TypeJudgmentExpr) Mutable() bool    { return false }
func (self *TypeJudgmentExpr) Storable() bool   { return false }
func (self *TypeJudgmentExpr) Value() hir.Value { return self.value }
func (self *TypeJudgmentExpr) Target() hir.Type { return self.target }

// GenericFuncInstExpr 泛型函数实例
type GenericFuncInstExpr struct {
	fn   values.Callable
	args []hir.Type
}

func NewGenericFuncInstExpr(fn values.Callable, args ...hir.Type) *GenericFuncInstExpr {
	return &GenericFuncInstExpr{
		fn:   fn,
		args: args,
	}
}
func (self *GenericFuncInstExpr) Type() hir.Type {
	return self.CallableType()
}
func (self *GenericFuncInstExpr) Mutable() bool            { return false }
func (self *GenericFuncInstExpr) Storable() bool           { return false }
func (self *GenericFuncInstExpr) GetFunc() values.Callable { return self.fn }
func (self *GenericFuncInstExpr) GenericArgs() []hir.Type  { return self.args }
func (self *GenericFuncInstExpr) GenericParamMap() hashmap.HashMap[types.VirtualType, hir.Type] {
	type getCompiler interface {
		GenericParams() []types.GenericParamType
	}
	var i int
	return hashmap.AnyWith[types.VirtualType, hir.Type](stlslices.FlatMap(self.fn.(getCompiler).GenericParams(), func(_ int, compileParam types.GenericParamType) []any {
		i++
		return []any{types.VirtualType(compileParam), self.args[i-1]}
	})...)
}
func (self *GenericFuncInstExpr) CallableType() types.CallableType {
	return types.ReplaceVirtualType(self.GenericParamMap(), self.fn.CallableType()).(types.CallableType)
}

// StaticMethodExpr 静态方法
type StaticMethodExpr struct {
	self        types.CustomType
	method      CallableDef
	genericArgs []hir.Type
}

func NewStaticMethodExpr(self types.CustomType, method CallableDef, genericArgs []hir.Type) *StaticMethodExpr {
	return &StaticMethodExpr{self: self, method: method, genericArgs: genericArgs}
}
func (self *StaticMethodExpr) Type() hir.Type          { return self.CallableType() }
func (self *StaticMethodExpr) Mutable() bool           { return false }
func (self *StaticMethodExpr) Storable() bool          { return false }
func (self *StaticMethodExpr) Method() CallableDef     { return self.method }
func (self *StaticMethodExpr) GenericArgs() []hir.Type { return self.genericArgs }
func (self *StaticMethodExpr) GenericParamMap() hashmap.HashMap[types.VirtualType, hir.Type] {
	type getCompiler interface {
		GenericParams() []types.GenericParamType
	}
	var i int
	genericParamMap := hashmap.AnyWith[types.VirtualType, hir.Type](stlslices.FlatMap(self.method.(getCompiler).GenericParams(), func(_ int, compileParam types.GenericParamType) []any {
		i++
		return []any{types.VirtualType(compileParam), self.genericArgs[i-1]}
	})...)

	if t, ok := types.As[types.GenericCustomType](self.self, true); ok {
		for iter := t.GenericParamMap().Iterator(); iter.Next(); {
			genericParamMap.Set(iter.Value().Unpack())
		}
	}
	return genericParamMap
}
func (self *StaticMethodExpr) CallableType() types.CallableType {
	ft := self.method.CallableType()

	genericParamMap := self.GenericParamMap()
	if genericParamMap.Empty() {
		return ft
	} else {
		return types.ReplaceVirtualType(genericParamMap, ft).(types.CallableType)
	}
}

// TraitStaticMethodExpr Trait静态方法
type TraitStaticMethodExpr struct {
	self   types.GenericParamType
	method string
}

func NewTraitStaticMethodExpr(self types.GenericParamType, method string) *TraitStaticMethodExpr {
	return &TraitStaticMethodExpr{self: self, method: method}
}
func (self *TraitStaticMethodExpr) Type() hir.Type               { return self.CallableType() }
func (self *TraitStaticMethodExpr) Mutable() bool                { return false }
func (self *TraitStaticMethodExpr) Storable() bool               { return false }
func (self *TraitStaticMethodExpr) Self() types.GenericParamType { return self.self }
func (self *TraitStaticMethodExpr) Method() string               { return self.method }
func (self *TraitStaticMethodExpr) CallableType() types.CallableType {
	trait := stlval.IgnoreWith(self.self.Restraint())
	ft := stlval.IgnoreWith(trait.GetMethodType(self.method))
	return types.ReplaceVirtualType(hashmap.AnyWith[types.VirtualType, hir.Type](types.Self, self.self), ft).(types.CallableType)
}

// TraitMethodExpr Trait方法
type TraitMethodExpr struct {
	self      hir.Value
	method    string
	onCapture func(hir.Type)
}

func NewTraitMethodExpr(self hir.Value, method string) *TraitMethodExpr {
	return &TraitMethodExpr{self: self, method: method}
}
func (self *TraitMethodExpr) Type() hir.Type  { return self.CallableType() }
func (self *TraitMethodExpr) Mutable() bool   { return false }
func (self *TraitMethodExpr) Storable() bool  { return false }
func (self *TraitMethodExpr) Self() hir.Value { return self.self }
func (self *TraitMethodExpr) Method() string  { return self.method }
func (self *TraitMethodExpr) CallableType() types.CallableType {
	gpt := stlval.IgnoreWith(types.As[types.GenericParamType](self.self.Type(), true))
	trait := stlval.IgnoreWith(gpt.Restraint())
	ft := stlval.IgnoreWith(trait.GetMethodType(self.method)).(types.CallableType)
	ft = types.NewLambdaType(ft.Ret(), ft.Params()[1:]...)
	return types.ReplaceVirtualType(hashmap.AnyWith[types.VirtualType, hir.Type](types.Self, self.self.Type()), ft).(types.CallableType)
}

// DropExpr 释放内存
type DropExpr struct {
	value hir.Value
}

func NewDropExpr(value hir.Value) *DropExpr {
	return &DropExpr{value: value}
}
func (self *DropExpr) Type() hir.Type { return types.NoReturn }
func (self *DropExpr) Mutable() bool  { return false }
func (self *DropExpr) Storable() bool { return false }
func (self *DropExpr) Value() hir.Value {
	return self.value
}
