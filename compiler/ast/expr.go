package ast

import (
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
)

// Expr 表达式
type Expr interface {
	Stmt
	expr()
}

// Integer 整数
type Integer struct {
	Value token.Token
}

func (self *Integer) Position() reader.Position {
	return self.Value.Position
}

func (self *Integer) stmt() {}

func (self *Integer) expr() {}

// Char 字符
type Char struct {
	Value token.Token
}

func (self *Char) Position() reader.Position {
	return self.Value.Position
}

func (self *Char) stmt() {}

func (self *Char) expr() {}

// Float 浮点数
type Float struct {
	Value token.Token
}

func (self *Float) Position() reader.Position {
	return self.Value.Position
}

func (self *Float) stmt() {}

func (self *Float) expr() {}

// Binary 二元运算
type Binary struct {
	Left  Expr
	Opera token.Token
	Right Expr
}

func (self *Binary) Position() reader.Position {
	return reader.MixPosition(self.Left.Position(), self.Right.Position())
}

func (self *Binary) stmt() {}

func (self *Binary) expr() {}

// Unary 一元运算
type Unary struct {
	Opera token.Token
	Value Expr
}

func (self *Unary) Position() reader.Position {
	return reader.MixPosition(self.Opera.Position, self.Value.Position())
}

func (self *Unary) stmt() {}

func (self *Unary) expr() {}

// IdentExpr 标识符表达式
type IdentExpr Ident

func (self *IdentExpr) Position() reader.Position {
	return (*Ident)(self).Position()
}

func (self *IdentExpr) stmt() {}

func (self *IdentExpr) expr() {}

// Call 调用
type Call struct {
	Func Expr
	Args []Expr
	End  reader.Position
}

func (self *Call) Position() reader.Position {
	return reader.MixPosition(self.Func.Position(), self.End)
}

func (self *Call) stmt() {}

func (self *Call) expr() {}

// Covert 类型转换
type Covert struct {
	Value Expr
	Type  Type
}

func (self *Covert) Position() reader.Position {
	return reader.MixPosition(self.Value.Position(), self.Type.Position())
}

func (self *Covert) stmt() {}

func (self *Covert) expr() {}

// Array 数组
type Array struct {
	Begin reader.Position
	Elems []Expr
	End   reader.Position
}

func (self *Array) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *Array) stmt() {}

func (self *Array) expr() {}

// Index 索引
type Index struct {
	From  Expr
	Index Expr
}

func (self *Index) Position() reader.Position {
	return reader.MixPosition(self.From.Position(), self.Index.Position())
}

func (self *Index) stmt() {}

func (self *Index) expr() {}

// Tuple 元组
type Tuple struct {
	Begin reader.Position
	Elems []Expr
	End   reader.Position
}

func (self *Tuple) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *Tuple) stmt() {}

func (self *Tuple) expr() {}

// Extract 提取
type Extract struct {
	From  Expr
	Index token.Token
}

func (self *Extract) Position() reader.Position {
	return reader.MixPosition(self.From.Position(), self.Index.Position)
}

func (self *Extract) stmt() {}

func (self *Extract) expr() {}

// Struct 结构体
type Struct struct {
	Type   Type
	Fields []pair.Pair[token.Token, Expr]
	End    reader.Position
}

func (self *Struct) Position() reader.Position {
	return reader.MixPosition(self.Type.Position(), self.End)
}

func (self *Struct) stmt() {}

func (self *Struct) expr() {}

// GetField 取字段
type GetField struct {
	From  Expr
	Index GenericName
}

func (self *GetField) Position() reader.Position {
	return reader.MixPosition(self.From.Position(), self.Index.Position())
}

func (self *GetField) stmt() {}

func (self *GetField) expr() {}

// String 字符串
type String struct {
	Value token.Token
}

func (self *String) Position() reader.Position {
	return self.Value.Position
}

func (self *String) stmt() {}

func (self *String) expr() {}

// Judgment 判断
type Judgment struct {
	Value Expr
	Type  Type
}

func (self *Judgment) Position() reader.Position {
	return reader.MixPosition(self.Value.Position(), self.Type.Position())
}

func (self *Judgment) stmt() {}

func (self *Judgment) expr() {}

// Null 空指针
type Null struct {
	Token token.Token
}

func (self *Null) Position() reader.Position {
	return self.Token.Position
}

func (self *Null) stmt() {}

func (self *Null) expr() {}

// CheckNull 空指针检查
type CheckNull struct {
	Value Expr
	End   reader.Position
}

func (self *CheckNull) Position() reader.Position {
	return reader.MixPosition(self.Value.Position(), self.End)
}

func (self *CheckNull) stmt() {}

func (self *CheckNull) expr() {}
