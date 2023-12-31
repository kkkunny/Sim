package ast

import (
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
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

// Boolean 布尔值
type Boolean struct {
	Value token.Token
}

func (self *Boolean) Position() reader.Position {
	return self.Value.Position
}

func (self *Boolean) stmt() {}

func (self *Boolean) expr() {}

// Ident 标识符
type Ident struct {
	Pkg  util.Option[token.Token]
	Name token.Token
	GenericArgs []Type
	End reader.Position
}

func (self *Ident) Position() reader.Position {
	if pkg, ok := self.Pkg.Value(); ok {
		return reader.MixPosition(pkg.Position, self.End)
	}else{
		return reader.MixPosition(self.Name.Position, self.End)
	}
}

func (self *Ident) stmt() {}

func (self *Ident) expr() {}

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
	Type   *IdentType
	Fields []pair.Pair[token.Token, Expr]
	End    reader.Position
}

func (self *Struct) Position() reader.Position {
	return reader.MixPosition(self.Type.Position(), self.End)
}

func (self *Struct) stmt() {}

func (self *Struct) expr() {}

// Field 取字段
type Field struct {
	From  Expr
	Index GenericName
}

func (self *Field) Position() reader.Position {
	return reader.MixPosition(self.From.Position(), self.Index.Position())
}

func (self *Field) stmt() {}

func (self *Field) expr() {}

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

type SelfValue struct {
	Token token.Token
}

func (self *SelfValue) Position() reader.Position {
	return self.Token.Position
}

func (self *SelfValue) stmt() {}

func (self *SelfValue) expr() {}
