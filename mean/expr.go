package mean

import (
	"math/big"

	"github.com/samber/lo"
)

// Expr 表达式
type Expr interface {
	Stmt
	GetType() Type
}

// Ident 标识符
type Ident interface {
	Expr
	ident()
}

// Integer 整数
type Integer struct {
	Type  IntType
	Value big.Int
}

func (self *Integer) stmt() {}

func (self *Integer) GetType() Type {
	return self.Type
}

// Float 浮点数
type Float struct {
	Type  *FloatType
	Value big.Float
}

func (self *Float) stmt() {}

func (self *Float) GetType() Type {
	return self.Type
}

// Binary 二元运算
type Binary interface {
	Expr
	GetLeft() Expr
	GetRight() Expr
}

// IntAndInt 整数且整数
type IntAndInt struct {
	Left, Right Expr
}

func (self *IntAndInt) stmt() {}

func (self *IntAndInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntAndInt) GetLeft() Expr {
	return self.Left
}

func (self *IntAndInt) GetRight() Expr {
	return self.Right
}

// IntOrInt 整数或整数
type IntOrInt struct {
	Left, Right Expr
}

func (self *IntOrInt) stmt() {}

func (self *IntOrInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntOrInt) GetLeft() Expr {
	return self.Left
}

func (self *IntOrInt) GetRight() Expr {
	return self.Right
}

// IntXorInt 整数异或整数
type IntXorInt struct {
	Left, Right Expr
}

func (self *IntXorInt) stmt() {}

func (self *IntXorInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntXorInt) GetLeft() Expr {
	return self.Left
}

func (self *IntXorInt) GetRight() Expr {
	return self.Right
}

// IntShlInt 整数左移整数
type IntShlInt struct {
	Left, Right Expr
}

func (self *IntShlInt) stmt() {}

func (self *IntShlInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntShlInt) GetLeft() Expr {
	return self.Left
}

func (self *IntShlInt) GetRight() Expr {
	return self.Right
}

// IntShrInt 整数右移整数
type IntShrInt struct {
	Left, Right Expr
}

func (self *IntShrInt) stmt() {}

func (self *IntShrInt) GetType() Type {
	return self.Left.GetType()
}

func (self *IntShrInt) GetLeft() Expr {
	return self.Left
}

func (self *IntShrInt) GetRight() Expr {
	return self.Right
}

// NumAddNum 数字加数字
type NumAddNum struct {
	Left, Right Expr
}

func (self *NumAddNum) stmt() {}

func (self *NumAddNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumAddNum) GetLeft() Expr {
	return self.Left
}

func (self *NumAddNum) GetRight() Expr {
	return self.Right
}

// NumSubNum 数字减数字
type NumSubNum struct {
	Left, Right Expr
}

func (self *NumSubNum) stmt() {}

func (self *NumSubNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumSubNum) GetLeft() Expr {
	return self.Left
}

func (self *NumSubNum) GetRight() Expr {
	return self.Right
}

// NumMulNum 数字乘数字
type NumMulNum struct {
	Left, Right Expr
}

func (self *NumMulNum) stmt() {}

func (self *NumMulNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumMulNum) GetLeft() Expr {
	return self.Left
}

func (self *NumMulNum) GetRight() Expr {
	return self.Right
}

// NumDivNum 数字除数字
type NumDivNum struct {
	Left, Right Expr
}

func (self *NumDivNum) stmt() {}

func (self *NumDivNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumDivNum) GetLeft() Expr {
	return self.Left
}

func (self *NumDivNum) GetRight() Expr {
	return self.Right
}

// NumRemNum 数字取余数字
type NumRemNum struct {
	Left, Right Expr
}

func (self *NumRemNum) stmt() {}

func (self *NumRemNum) GetType() Type {
	return self.Left.GetType()
}

func (self *NumRemNum) GetLeft() Expr {
	return self.Left
}

func (self *NumRemNum) GetRight() Expr {
	return self.Right
}

// NumLtNum 数字小于数字
type NumLtNum struct {
	Left, Right Expr
}

func (self *NumLtNum) stmt() {}

func (self *NumLtNum) GetType() Type {
	return Bool
}

func (self *NumLtNum) GetLeft() Expr {
	return self.Left
}

func (self *NumLtNum) GetRight() Expr {
	return self.Right
}

// NumGtNum 数字大于数字
type NumGtNum struct {
	Left, Right Expr
}

func (self *NumGtNum) stmt() {}

func (self *NumGtNum) GetType() Type {
	return Bool
}

func (self *NumGtNum) GetLeft() Expr {
	return self.Left
}

func (self *NumGtNum) GetRight() Expr {
	return self.Right
}

// NumLeNum 数字小于等于数字
type NumLeNum struct {
	Left, Right Expr
}

func (self *NumLeNum) stmt() {}

func (self *NumLeNum) GetType() Type {
	return Bool
}

func (self *NumLeNum) GetLeft() Expr {
	return self.Left
}

func (self *NumLeNum) GetRight() Expr {
	return self.Right
}

// NumGeNum 数字大于等于数字
type NumGeNum struct {
	Left, Right Expr
}

func (self *NumGeNum) stmt() {}

func (self *NumGeNum) GetType() Type {
	return Bool
}

func (self *NumGeNum) GetLeft() Expr {
	return self.Left
}

func (self *NumGeNum) GetRight() Expr {
	return self.Right
}

// NumEqNum 数字等于数字
type NumEqNum struct {
	Left, Right Expr
}

func (self *NumEqNum) stmt() {}

func (self *NumEqNum) GetType() Type {
	return Bool
}

func (self *NumEqNum) GetLeft() Expr {
	return self.Left
}

func (self *NumEqNum) GetRight() Expr {
	return self.Right
}

// BoolEqBool 布尔等于布尔
type BoolEqBool struct {
	Left, Right Expr
}

func (self *BoolEqBool) stmt() {}

func (self *BoolEqBool) GetType() Type {
	return Bool
}

func (self *BoolEqBool) GetLeft() Expr {
	return self.Left
}

func (self *BoolEqBool) GetRight() Expr {
	return self.Right
}

// FuncEqFunc 函数等于函数
type FuncEqFunc struct {
	Left, Right Expr
}

func (self *FuncEqFunc) stmt() {}

func (self *FuncEqFunc) GetType() Type {
	return Bool
}

func (self *FuncEqFunc) GetLeft() Expr {
	return self.Left
}

func (self *FuncEqFunc) GetRight() Expr {
	return self.Right
}

// ArrayEqArray 数组等于数组
type ArrayEqArray struct {
	Left, Right Expr
}

func (self *ArrayEqArray) stmt() {}

func (self *ArrayEqArray) GetType() Type {
	return Bool
}

func (self *ArrayEqArray) GetLeft() Expr {
	return self.Left
}

func (self *ArrayEqArray) GetRight() Expr {
	return self.Right
}

// TupleEqTuple 元组等于元组
type TupleEqTuple struct {
	Left, Right Expr
}

func (self *TupleEqTuple) stmt() {}

func (self *TupleEqTuple) GetType() Type {
	return Bool
}

func (self *TupleEqTuple) GetLeft() Expr {
	return self.Left
}

func (self *TupleEqTuple) GetRight() Expr {
	return self.Right
}

// StructEqStruct 结构体等于结构体
type StructEqStruct struct {
	Left, Right Expr
}

func (self *StructEqStruct) stmt() {}

func (self *StructEqStruct) GetType() Type {
	return Bool
}

func (self *StructEqStruct) GetLeft() Expr {
	return self.Left
}

func (self *StructEqStruct) GetRight() Expr {
	return self.Right
}

// NumNeNum 数字不等数字
type NumNeNum struct {
	Left, Right Expr
}

func (self *NumNeNum) stmt() {}

func (self *NumNeNum) GetType() Type {
	return Bool
}

func (self *NumNeNum) GetLeft() Expr {
	return self.Left
}

func (self *NumNeNum) GetRight() Expr {
	return self.Right
}

// BoolNeBool 布尔不等布尔
type BoolNeBool struct {
	Left, Right Expr
}

func (self *BoolNeBool) stmt() {}

func (self *BoolNeBool) GetType() Type {
	return Bool
}

func (self *BoolNeBool) GetLeft() Expr {
	return self.Left
}

func (self *BoolNeBool) GetRight() Expr {
	return self.Right
}

// FuncNeFunc 函数不等函数
type FuncNeFunc struct {
	Left, Right Expr
}

func (self *FuncNeFunc) stmt() {}

func (self *FuncNeFunc) GetType() Type {
	return Bool
}

func (self *FuncNeFunc) GetLeft() Expr {
	return self.Left
}

func (self *FuncNeFunc) GetRight() Expr {
	return self.Right
}

// ArrayNeArray 数组不等数组
type ArrayNeArray struct {
	Left, Right Expr
}

func (self *ArrayNeArray) stmt() {}

func (self *ArrayNeArray) GetType() Type {
	return Bool
}

func (self *ArrayNeArray) GetLeft() Expr {
	return self.Left
}

func (self *ArrayNeArray) GetRight() Expr {
	return self.Right
}

// TupleNeTuple 元组不等元组
type TupleNeTuple struct {
	Left, Right Expr
}

func (self *TupleNeTuple) stmt() {}

func (self *TupleNeTuple) GetType() Type {
	return Bool
}

func (self *TupleNeTuple) GetLeft() Expr {
	return self.Left
}

func (self *TupleNeTuple) GetRight() Expr {
	return self.Right
}

// StructNeStruct 结构体不等结构体
type StructNeStruct struct {
	Left, Right Expr
}

func (self *StructNeStruct) stmt() {}

func (self *StructNeStruct) GetType() Type {
	return Bool
}

func (self *StructNeStruct) GetLeft() Expr {
	return self.Left
}

func (self *StructNeStruct) GetRight() Expr {
	return self.Right
}

// BoolAndBool 布尔并且布尔
type BoolAndBool struct {
	Left, Right Expr
}

func (self *BoolAndBool) stmt() {}

func (self *BoolAndBool) GetType() Type {
	return Bool
}

func (self *BoolAndBool) GetLeft() Expr {
	return self.Left
}

func (self *BoolAndBool) GetRight() Expr {
	return self.Right
}

// BoolOrBool 布尔或者布尔
type BoolOrBool struct {
	Left, Right Expr
}

func (self *BoolOrBool) stmt() {}

func (self *BoolOrBool) GetType() Type {
	return Bool
}

func (self *BoolOrBool) GetLeft() Expr {
	return self.Left
}

func (self *BoolOrBool) GetRight() Expr {
	return self.Right
}

// Unary 一元运算
type Unary interface {
	Expr
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

func (self *NumNegate) GetValue() Expr {
	return self.Value
}

// IntBitNegate 整数按位取反
type IntBitNegate struct {
	Value Expr
}

func (self *IntBitNegate) stmt() {}

func (self *IntBitNegate) GetType() Type {
	return self.Value.GetType()
}

func (self *IntBitNegate) GetValue() Expr {
	return self.Value
}

// BoolNegate 布尔取反
type BoolNegate struct {
	Value Expr
}

func (self *BoolNegate) stmt() {}

func (self *BoolNegate) GetType() Type {
	return Bool
}

func (self *BoolNegate) GetValue() Expr {
	return self.Value
}

// Boolean 布尔值
type Boolean struct {
	Value bool
}

func (self *Boolean) stmt() {}

func (self *Boolean) GetType() Type {
	return Bool
}

// Call 调用
type Call struct {
	Func Expr
	Args []Expr
}

func (self *Call) stmt() {}

func (self *Call) GetType() Type {
	return self.Func.GetType().(*FuncType).Ret
}

// Covert 类型转换
type Covert interface {
	Expr
	GetFrom() Expr
}

// Num2Num 数字类型转数字类型
type Num2Num struct {
	From Expr
	To   NumberType
}

func (self *Num2Num) stmt() {}

func (self *Num2Num) GetType() Type {
	return self.To
}

func (self *Num2Num) GetFrom() Expr {
	return self.From
}

// Array 数组
type Array struct {
	Type  *ArrayType
	Elems []Expr
}

func (self *Array) stmt() {}

func (self *Array) GetType() Type {
	return self.Type
}

// Index 索引
type Index struct {
	From  Expr
	Index Expr
}

func (self *Index) stmt() {}

func (self *Index) GetType() Type {
	return self.From.GetType().(*ArrayType).Elem
}

// Tuple 元组
type Tuple struct {
	Elems []Expr
}

func (self *Tuple) stmt() {}

func (self *Tuple) GetType() Type {
	ets := lo.Map(self.Elems, func(item Expr, index int) Type {
		return item.GetType()
	})
	return &TupleType{Elems: ets}
}

// Extract 提取
type Extract struct {
	From  Expr
	Index uint
}

func (self *Extract) stmt() {}

func (self *Extract) GetType() Type {
	return self.From.GetType().(*TupleType).Elems[self.Index]
}

// Param 参数
type Param struct {
	Type Type
	Name string
}

func (*Param) stmt() {}

func (self *Param) GetType() Type {
	return self.Type
}

func (*Param) ident() {}

// Struct 结构体
type Struct struct {
	Type   *StructType
	Fields []Expr
}

func (*Struct) stmt() {}

func (self *Struct) GetType() Type {
	return self.Type
}

// Zero 零值
type Zero struct {
	Type Type
}

func (*Zero) stmt() {}

func (self *Zero) GetType() Type {
	return self.Type
}

// Field 字段
type Field struct {
	From  Expr
	Index uint
}

func (self *Field) stmt() {}

func (self *Field) GetType() Type {
	return self.From.GetType().(*StructType).Fields.Values().Get(self.Index)
}
