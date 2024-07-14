package hir

import (
	"math/big"

	stlbasic "github.com/kkkunny/stl/basic"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/util"
)

// ExprStmt 表达式语句
type ExprStmt interface {
	Stmt
	Expr
}

// Integer 整数
type Integer struct {
	Type  Type
	Value *big.Int
}

func (self *Integer) stmt() {}

func (self *Integer) GetType() Type {
	return self.Type
}

func (self *Integer) Mutable() bool {
	return false
}

func (*Integer) Temporary() bool {
	return true
}

// Float 浮点数
type Float struct {
	Type  Type
	Value *big.Float
}

func (self *Float) stmt() {}

func (self *Float) GetType() Type {
	return self.Type
}

func (self *Float) Mutable() bool {
	return false
}

func (*Float) Temporary() bool {
	return true
}

// Binary 二元运算
type Binary interface {
	ExprStmt
	GetLeft() Expr
	GetRight() Expr
}

// Assign 赋值
type Assign struct {
	Left, Right Expr
}

func (self *Assign) stmt() {}

func (self *Assign) GetType() Type {
	return NoThing
}

func (self *Assign) Mutable() bool {
	return false
}

func (self *Assign) GetLeft() Expr {
	return self.Left
}

func (self *Assign) GetRight() Expr {
	return self.Right
}

func (*Assign) Temporary() bool {
	return true
}

// IntAndInt 整数且整数
type IntAndInt struct {
	Left, Right Expr
}

func (self *IntAndInt) stmt() {}

func (self *IntAndInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntAndInt) Mutable() bool {
	return false
}

func (self *IntAndInt) GetLeft() Expr {
	return self.Left
}

func (self *IntAndInt) GetRight() Expr {
	return self.Right
}

func (*IntAndInt) Temporary() bool {
	return true
}

// IntOrInt 整数或整数
type IntOrInt struct {
	Left, Right Expr
}

func (self *IntOrInt) stmt() {}

func (self *IntOrInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntOrInt) Mutable() bool {
	return false
}

func (self *IntOrInt) GetLeft() Expr {
	return self.Left
}

func (self *IntOrInt) GetRight() Expr {
	return self.Right
}

func (*IntOrInt) Temporary() bool {
	return true
}

// IntXorInt 整数异或整数
type IntXorInt struct {
	Left, Right Expr
}

func (self *IntXorInt) stmt() {}

func (self *IntXorInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntXorInt) Mutable() bool {
	return false
}

func (self *IntXorInt) GetLeft() Expr {
	return self.Left
}

func (self *IntXorInt) GetRight() Expr {
	return self.Right
}

func (*IntXorInt) Temporary() bool {
	return true
}

// IntShlInt 整数左移整数
type IntShlInt struct {
	Left, Right Expr
}

func (self *IntShlInt) stmt() {}

func (self *IntShlInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntShlInt) Mutable() bool {
	return false
}

func (self *IntShlInt) GetLeft() Expr {
	return self.Left
}

func (self *IntShlInt) GetRight() Expr {
	return self.Right
}

func (*IntShlInt) Temporary() bool {
	return true
}

// IntShrInt 整数右移整数
type IntShrInt struct {
	Left, Right Expr
}

func (self *IntShrInt) stmt() {}

func (self *IntShrInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntShrInt) Mutable() bool {
	return false
}

func (self *IntShrInt) GetLeft() Expr {
	return self.Left
}

func (self *IntShrInt) GetRight() Expr {
	return self.Right
}

func (*IntShrInt) Temporary() bool {
	return true
}

// NumAddNum 数字加数字
type NumAddNum struct {
	Left, Right Expr
}

func (self *NumAddNum) stmt() {}

func (self *NumAddNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumAddNum) Mutable() bool {
	return false
}

func (self *NumAddNum) GetLeft() Expr {
	return self.Left
}

func (self *NumAddNum) GetRight() Expr {
	return self.Right
}

func (*NumAddNum) Temporary() bool {
	return true
}

// NumSubNum 数字减数字
type NumSubNum struct {
	Left, Right Expr
}

func (self *NumSubNum) stmt() {}

func (self *NumSubNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumSubNum) Mutable() bool {
	return false
}

func (self *NumSubNum) GetLeft() Expr {
	return self.Left
}

func (self *NumSubNum) GetRight() Expr {
	return self.Right
}

func (*NumSubNum) Temporary() bool {
	return true
}

// NumMulNum 数字乘数字
type NumMulNum struct {
	Left, Right Expr
}

func (self *NumMulNum) stmt() {}

func (self *NumMulNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumMulNum) Mutable() bool {
	return false
}

func (self *NumMulNum) GetLeft() Expr {
	return self.Left
}

func (self *NumMulNum) GetRight() Expr {
	return self.Right
}

func (*NumMulNum) Temporary() bool {
	return true
}

// NumDivNum 数字除数字
type NumDivNum struct {
	Left, Right Expr
}

func (self *NumDivNum) stmt() {}

func (self *NumDivNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumDivNum) Mutable() bool {
	return false
}

func (self *NumDivNum) GetLeft() Expr {
	return self.Left
}

func (self *NumDivNum) GetRight() Expr {
	return self.Right
}

func (*NumDivNum) Temporary() bool {
	return true
}

// NumRemNum 数字取余数字
type NumRemNum struct {
	Left, Right Expr
}

func (self *NumRemNum) stmt() {}

func (self *NumRemNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumRemNum) Mutable() bool {
	return false
}

func (self *NumRemNum) GetLeft() Expr {
	return self.Left
}

func (self *NumRemNum) GetRight() Expr {
	return self.Right
}

func (*NumRemNum) Temporary() bool {
	return true
}

// NumLtNum 数字小于数字
type NumLtNum struct {
	BoolType    Type
	Left, Right Expr
}

func (self *NumLtNum) stmt() {}

func (self *NumLtNum) GetType() Type {
	return self.BoolType
}

func (self *NumLtNum) Mutable() bool {
	return false
}

func (self *NumLtNum) GetLeft() Expr {
	return self.Left
}

func (self *NumLtNum) GetRight() Expr {
	return self.Right
}

func (*NumLtNum) Temporary() bool {
	return true
}

// NumGtNum 数字大于数字
type NumGtNum struct {
	BoolType    Type
	Left, Right Expr
}

func (self *NumGtNum) stmt() {}

func (self *NumGtNum) GetType() Type {
	return self.BoolType
}

func (self *NumGtNum) Mutable() bool {
	return false
}

func (self *NumGtNum) GetLeft() Expr {
	return self.Left
}

func (self *NumGtNum) GetRight() Expr {
	return self.Right
}

func (*NumGtNum) Temporary() bool {
	return true
}

// NumLeNum 数字小于等于数字
type NumLeNum struct {
	BoolType    Type
	Left, Right Expr
}

func (self *NumLeNum) stmt() {}

func (self *NumLeNum) GetType() Type {
	return self.BoolType
}

func (self *NumLeNum) Mutable() bool {
	return false
}

func (self *NumLeNum) GetLeft() Expr {
	return self.Left
}

func (self *NumLeNum) GetRight() Expr {
	return self.Right
}

func (*NumLeNum) Temporary() bool {
	return true
}

// NumGeNum 数字大于等于数字
type NumGeNum struct {
	BoolType    Type
	Left, Right Expr
}

func (self *NumGeNum) stmt() {}

func (self *NumGeNum) GetType() Type {
	return self.BoolType
}

func (self *NumGeNum) Mutable() bool {
	return false
}

func (self *NumGeNum) GetLeft() Expr {
	return self.Left
}

func (self *NumGeNum) GetRight() Expr {
	return self.Right
}

func (*NumGeNum) Temporary() bool {
	return true
}

// Equal 比较相等
type Equal struct {
	BoolType    Type
	Left, Right Expr
}

func (self *Equal) stmt() {}

func (self *Equal) GetType() Type {
	return self.BoolType
}

func (self *Equal) Mutable() bool {
	return false
}

func (self *Equal) GetLeft() Expr {
	return self.Left
}

func (self *Equal) GetRight() Expr {
	return self.Right
}

func (*Equal) Temporary() bool {
	return true
}

// NotEqual 比较不相等
type NotEqual struct {
	BoolType    Type
	Left, Right Expr
}

func (self *NotEqual) stmt() {}

func (self *NotEqual) GetType() Type {
	return self.BoolType
}

func (self *NotEqual) Mutable() bool {
	return false
}

func (self *NotEqual) GetLeft() Expr {
	return self.Left
}

func (self *NotEqual) GetRight() Expr {
	return self.Right
}

func (*NotEqual) Temporary() bool {
	return true
}

// BoolAndBool 布尔并且布尔
type BoolAndBool struct {
	Left, Right Expr
}

func (self *BoolAndBool) stmt() {}

func (self *BoolAndBool) GetType() Type {
	return self.Left.GetType()
}

func (self *BoolAndBool) Mutable() bool {
	return false
}

func (self *BoolAndBool) GetLeft() Expr {
	return self.Left
}

func (self *BoolAndBool) GetRight() Expr {
	return self.Right
}

func (*BoolAndBool) Temporary() bool {
	return true
}

// BoolOrBool 布尔或者布尔
type BoolOrBool struct {
	Left, Right Expr
}

func (self *BoolOrBool) stmt() {}

func (self *BoolOrBool) GetType() Type {
	return self.Left.GetType()
}

func (self *BoolOrBool) Mutable() bool {
	return false
}

func (self *BoolOrBool) GetLeft() Expr {
	return self.Left
}

func (self *BoolOrBool) GetRight() Expr {
	return self.Right
}

func (*BoolOrBool) Temporary() bool {
	return true
}

// Unary 一元运算
type Unary interface {
	ExprStmt
	GetValue() Expr
}

// NumNegate 数字取相反数
type NumNegate struct {
	Value Expr
}

func (self *NumNegate) stmt() {}

func (self *NumNegate) GetType() Type {
	return self.Value.GetType()
}

func (self *NumNegate) Mutable() bool {
	return false
}

func (self *NumNegate) GetValue() Expr {
	return self.Value
}

func (*NumNegate) Temporary() bool {
	return true
}

// IntBitNegate 整数按位取反
type IntBitNegate struct {
	Value Expr
}

func (self *IntBitNegate) stmt() {}

func (self *IntBitNegate) GetType() Type {
	return self.Value.GetType()
}

func (self *IntBitNegate) Mutable() bool {
	return false
}

func (self *IntBitNegate) GetValue() Expr {
	return self.Value
}

func (*IntBitNegate) Temporary() bool {
	return true
}

// BoolNegate 布尔取反
type BoolNegate struct {
	Value Expr
}

func (self *BoolNegate) stmt() {}

func (self *BoolNegate) GetType() Type {
	return self.Value.GetType()
}

func (self *BoolNegate) Mutable() bool {
	return false
}

func (self *BoolNegate) GetValue() Expr {
	return self.Value
}

func (*BoolNegate) Temporary() bool {
	return true
}

// Call 调用
type Call struct {
	Func Expr
	Args []Expr
}

func (self *Call) stmt() {}

func (self *Call) GetType() Type {
	return AsType[CallableType](self.Func.GetType()).GetRet()
}

func (self *Call) Mutable() bool {
	return false
}

func (*Call) Temporary() bool {
	return true
}

// TypeCovert 类型转换
type TypeCovert interface {
	ExprStmt
	GetFrom() Expr
}

// Num2Num 数字类型转数字类型
type Num2Num struct {
	From Expr
	To   Type
}

func (self *Num2Num) stmt() {}

func (self *Num2Num) GetType() Type {
	return self.To
}

func (*Num2Num) Mutable() bool {
	return false
}

func (self *Num2Num) GetFrom() Expr {
	return self.From
}

func (*Num2Num) Temporary() bool {
	return true
}

// DoNothingCovert 什么事都不做的转换
type DoNothingCovert struct {
	From Expr
	To   Type
}

func (self *DoNothingCovert) stmt() {}

func (self *DoNothingCovert) GetType() Type {
	return self.To
}

func (*DoNothingCovert) Mutable() bool {
	return false
}

func (self *DoNothingCovert) GetFrom() Expr {
	return self.From
}

func (*DoNothingCovert) Temporary() bool {
	return true
}

// NoReturn2Any 无返回转任意类型
type NoReturn2Any struct {
	From Expr
	To   Type
}

func (self *NoReturn2Any) stmt() {}

func (self *NoReturn2Any) GetType() Type {
	return self.To
}

func (*NoReturn2Any) Mutable() bool {
	return false
}

func (self *NoReturn2Any) GetFrom() Expr {
	return self.From
}

func (*NoReturn2Any) Temporary() bool {
	return true
}

// Array 数组
type Array struct {
	Type  Type
	Elems []Expr
}

func (self *Array) stmt() {}

func (self *Array) GetType() Type {
	return self.Type
}

func (self *Array) Mutable() bool {
	return false
}

func (*Array) Temporary() bool {
	return true
}

// Index 索引
type Index struct {
	From  Expr
	Index Expr
}

func (self *Index) stmt() {}

func (self *Index) GetType() Type {
	return AsType[*ArrayType](self.From.GetType()).Elem
}

func (self *Index) Mutable() bool {
	return self.From.Mutable() && !self.Temporary()
}

func (self *Index) Temporary() bool {
	return self.From.Temporary()
}

// Tuple 元组
type Tuple struct {
	Elems []Expr
}

func (self *Tuple) stmt() {}

func (self *Tuple) GetType() Type {
	ets := stlslices.Map(self.Elems, func(_ int, e Expr) Type {
		return e.GetType()
	})
	return NewTupleType(ets...)
}

func (self *Tuple) Mutable() bool {
	for _, elem := range self.Elems {
		if !elem.Mutable() {
			return false
		}
	}
	return true
}

func (*Tuple) Temporary() bool {
	return true
}

// Extract 提取
type Extract struct {
	From  Expr
	Index uint
}

func (self *Extract) stmt() {}

func (self *Extract) GetType() Type {
	return AsType[*TupleType](self.From.GetType()).Elems[self.Index]
}

func (self *Extract) Mutable() bool {
	return self.From.Mutable() && !self.Temporary()
}

func (self *Extract) Temporary() bool {
	return self.From.Temporary()
}

// Struct 结构体
type Struct struct {
	Type   Type
	Fields []Expr
}

func (*Struct) stmt() {}

func (self *Struct) GetType() Type {
	return self.Type
}

func (self *Struct) Mutable() bool {
	return false
}

func (*Struct) Temporary() bool {
	return true
}

// Default 默认值
type Default struct {
	Type Type
}

func (*Default) stmt() {}

func (self *Default) GetType() Type {
	return self.Type
}

func (self *Default) Mutable() bool {
	return false
}

func (*Default) Temporary() bool {
	return true
}

// GetField 取字段
type GetField struct {
	Internal bool // 是否在包内部，不受字段可变性影响
	From     Expr
	Index    uint
}

func (self *GetField) stmt() {}

func (self *GetField) GetType() Type {
	return AsType[*StructType](self.From.GetType()).Fields.Values().Get(self.Index).Type
}

func (self *GetField) Mutable() bool {
	if self.Temporary() {
		return false
	} else if self.Internal {
		return true
	}
	return self.From.Mutable() && AsType[*StructType](self.From.GetType()).Fields.Values().Get(self.Index).Mutable
}

func (self *GetField) Temporary() bool {
	return self.From.Temporary()
}

// String 字符串
type String struct {
	Type  Type
	Value string
}

func (self *String) stmt() {}

func (self *String) GetType() Type {
	return self.Type
}

func (self *String) Mutable() bool {
	return false
}

func (*String) Temporary() bool {
	return true
}

// TypeJudgment 类型判断
type TypeJudgment struct {
	BoolType Type
	Value    Expr
	Type     Type
}

func (self *TypeJudgment) stmt() {}

func (self *TypeJudgment) GetType() Type {
	return self.BoolType
}

func (self *TypeJudgment) Mutable() bool {
	return false
}

func (*TypeJudgment) Temporary() bool {
	return true
}

// GetRef 取引用
type GetRef struct {
	Mut   bool
	Value Expr
}

func (self *GetRef) stmt() {}

func (self *GetRef) GetType() Type {
	return NewRefType(self.Mut, self.Value.GetType())
}

func (self *GetRef) Mutable() bool {
	return false
}

func (self *GetRef) GetValue() Expr {
	return self.Value
}

func (*GetRef) Temporary() bool {
	return true
}

// DeRef 解引用
type DeRef struct {
	Value Expr
}

func (self *DeRef) stmt() {}

func (self *DeRef) GetType() Type {
	return AsType[*RefType](self.Value.GetType()).Elem
}

func (self *DeRef) Mutable() bool {
	return AsType[*RefType](self.Value.GetType()).Mut && self.Value.Mutable()
}

func (self *DeRef) GetValue() Expr {
	return self.Value
}

func (*DeRef) Temporary() bool {
	return false
}

// Method 方法
// TODO: 逃逸分析
type Method struct {
	Self   Expr
	Define *MethodDef
}

// LoopFindMethodWithNoCheck 循环寻找方法
func LoopFindMethodWithNoCheck(ct *CustomType, selfVal util.Option[Expr], name string) util.Option[Expr] {
	method := ct.Methods.Get(name)
	if method != nil {
		if method.IsStatic() {
			return util.Some[Expr](method)
		} else if selfVal.IsNone() {
			return util.None[Expr]()
		} else {
			return util.Some[Expr](&Method{
				Self:   selfVal.MustValue(),
				Define: method,
			})
		}
	}
	tct, ok := TryCustomType(ct.Target)
	if !ok {
		return util.None[Expr]()
	}
	return LoopFindMethodWithNoCheck(tct, stlbasic.TernaryAction(selfVal.IsSome(), func() util.Option[Expr] {
		return util.Some[Expr](&DoNothingCovert{From: selfVal.MustValue(), To: tct})
	}, func() util.Option[Expr] {
		return util.None[Expr]()
	}), name)
}

func (self *Method) stmt() {}

func (self *Method) GetType() Type {
	ft := self.Define.GetMethodType()
	return NewLambdaType(ft.GetRet(), ft.GetParams()...)
}

func (self *Method) Mutable() bool {
	return false
}

func (*Method) Temporary() bool {
	return true
}

// Lambda 匿名函数
type Lambda struct {
	Type    Type
	Params  []*Param
	Ret     Type
	Body    *Block
	Context []Ident
}

func (self *Lambda) stmt() {}

func (self *Lambda) GetType() Type {
	return self.Type
}

func (self *Lambda) Mutable() bool {
	return false
}

func (*Lambda) Temporary() bool {
	return true
}

func (self *Lambda) GetFuncType() *FuncType {
	return NewFuncType(self.Ret, stlslices.Map(self.Params, func(_ int, e *Param) Type {
		return e.GetType()
	})...)
}

// Func2Lambda 函数转匿名函数
type Func2Lambda struct {
	From Expr
	To   Type
}

func (self *Func2Lambda) stmt() {}

func (self *Func2Lambda) GetType() Type {
	return self.To
}

func (*Func2Lambda) Mutable() bool {
	return false
}

func (self *Func2Lambda) GetFrom() Expr {
	return self.From
}

func (*Func2Lambda) Temporary() bool {
	return true
}

// Enum 枚举
type Enum struct {
	From  Type
	Field string
	Elems []Expr
}

func (self *Enum) stmt() {}

func (self *Enum) GetType() Type {
	return self.From
}

func (*Enum) Mutable() bool {
	return false
}

func (*Enum) Temporary() bool {
	return true
}

// Enum2Number 枚举转数字
type Enum2Number struct {
	From Expr
	To   Type
}

func (self *Enum2Number) stmt() {}

func (self *Enum2Number) GetType() Type {
	return self.To
}

func (*Enum2Number) Mutable() bool {
	return false
}

func (self *Enum2Number) GetFrom() Expr {
	return self.From
}

func (*Enum2Number) Temporary() bool {
	return true
}

// Number2Enum 数字转枚举
type Number2Enum struct {
	From Expr
	To   Type
}

func (self *Number2Enum) stmt() {}

func (self *Number2Enum) GetType() Type {
	return self.To
}

func (*Number2Enum) Mutable() bool {
	return false
}

func (self *Number2Enum) GetFrom() Expr {
	return self.From
}

func (*Number2Enum) Temporary() bool {
	return true
}
