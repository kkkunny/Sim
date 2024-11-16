package ast

import (
	"io"

	"github.com/kkkunny/stl/container/tuple"

	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/token"
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

func (self *Integer) Output(w io.Writer, depth uint) (err error) {
	return outputf(w, "%s", self.Value.Source())
}

// Char 字符
type Char struct {
	Value token.Token
}

func (self *Char) Position() reader.Position {
	return self.Value.Position
}

func (self *Char) stmt() {}

func (self *Char) expr() {}

func (self *Char) Output(w io.Writer, depth uint) (err error) {
	return outputf(w, "%s", self.Value.Source())
}

// Float 浮点数
type Float struct {
	Value token.Token
}

func (self *Float) Position() reader.Position {
	return self.Value.Position
}

func (self *Float) stmt() {}

func (self *Float) expr() {}

func (self *Float) Output(w io.Writer, depth uint) (err error) {
	return outputf(w, "%s", self.Value.Source())
}

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

func (self *Binary) Output(w io.Writer, depth uint) (err error) {
	if err = self.Left.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, " %s ", self.Opera.Source()); err != nil {
		return err
	}
	return self.Right.Output(w, depth)
}

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

func (self *Unary) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "%s", self.Opera.Source()); err != nil {
		return err
	}
	return self.Value.Output(w, depth)
}

// IdentExpr 标识符表达式
type IdentExpr Ident

func (self *IdentExpr) Position() reader.Position {
	return (*Ident)(self).Position()
}

func (self *IdentExpr) stmt() {}

func (self *IdentExpr) expr() {}

func (self *IdentExpr) Output(w io.Writer, depth uint) (err error) {
	return (*Ident)(self).Output(w, depth)
}

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

func (self *Call) Output(w io.Writer, depth uint) (err error) {
	if err = self.Func.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, "("); err != nil {
		return err
	}
	for i, arg := range self.Args {
		if err = arg.Output(w, depth); err != nil {
			return err
		}
		if i < len(self.Args)-1 {
			if err = outputf(w, ", "); err != nil {
				return err
			}
		}
	}
	return outputf(w, ")")
}

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

func (self *Covert) Output(w io.Writer, depth uint) (err error) {
	if err = self.Value.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, " as "); err != nil {
		return err
	}
	return self.Type.Output(w, depth)
}

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

func (self *Array) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "["); err != nil {
		return err
	}
	for i, elem := range self.Elems {
		if err = elem.Output(w, depth); err != nil {
			return err
		}
		if i < len(self.Elems)-1 {
			if err = outputf(w, ", "); err != nil {
				return err
			}
		}
	}
	return outputf(w, "]")
}

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

func (self *Index) Output(w io.Writer, depth uint) (err error) {
	if err = self.From.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, "["); err != nil {
		return err
	}
	if err = self.Index.Output(w, depth); err != nil {
		return err
	}
	return outputf(w, "]")
}

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

func (self *Tuple) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "("); err != nil {
		return err
	}
	for i, elem := range self.Elems {
		if err = elem.Output(w, depth); err != nil {
			return err
		}
		if i < len(self.Elems)-1 {
			if err = outputf(w, ", "); err != nil {
				return err
			}
		}
	}
	return outputf(w, ")")
}

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

func (self *Extract) Output(w io.Writer, depth uint) (err error) {
	if err = self.From.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, "["); err != nil {
		return err
	}
	if err = outputf(w, self.Index.Source()); err != nil {
		return err
	}
	return outputf(w, "]")
}

// Struct 结构体
type Struct struct {
	Type   Type
	Fields []tuple.Tuple2[token.Token, Expr]
	End    reader.Position
}

func (self *Struct) Position() reader.Position {
	return reader.MixPosition(self.Type.Position(), self.End)
}

func (self *Struct) stmt() {}

func (self *Struct) expr() {}

func (self *Struct) Output(w io.Writer, depth uint) (err error) {
	if err = self.Type.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, "{\n"); err != nil {
		return err
	}
	for i, field := range self.Fields {
		if err = outputDepth(w, depth+1); err != nil {
			return err
		}
		if err = outputf(w, "%s: ", field.E1().Source()); err != nil {
			return err
		}
		if err = field.E2().Output(w, depth+1); err != nil {
			return err
		}
		if i < len(self.Fields)-1 {
			if err = outputf(w, ","); err != nil {
				return err
			}
		}
		if err = outputf(w, "\n"); err != nil {
			return err
		}
	}
	if err = outputDepth(w, depth); err != nil {
		return err
	}
	return outputf(w, "}")
}

// Dot 点
type Dot struct {
	From  Expr
	Index token.Token
}

func (self *Dot) Position() reader.Position {
	return reader.MixPosition(self.From.Position(), self.Index.Position)
}

func (self *Dot) stmt() {}

func (self *Dot) expr() {}

func (self *Dot) Output(w io.Writer, depth uint) (err error) {
	if err = self.From.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, "."); err != nil {
		return err
	}
	return outputf(w, self.Index.Source())
}

// String 字符串
type String struct {
	Value token.Token
}

func (self *String) Position() reader.Position {
	return self.Value.Position
}

func (self *String) stmt() {}

func (self *String) expr() {}

func (self *String) Output(w io.Writer, depth uint) (err error) {
	return outputf(w, self.Value.Source())
}

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

func (self *Judgment) Output(w io.Writer, depth uint) (err error) {
	if err = self.Value.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, " is "); err != nil {
		return err
	}
	return self.Type.Output(w, depth)
}

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

func (self *CheckNull) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "!"); err != nil {
		return err
	}
	return self.Value.Output(w, depth)
}

// Lambda 匿名函数
type Lambda struct {
	Begin  reader.Position
	Params []Param
	Ret    Type
	Body   *Block
}

func (self *Lambda) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (self *Lambda) stmt() {}

func (self *Lambda) expr() {}

func (self *Lambda) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "("); err != nil {
		return err
	}
	for i, param := range self.Params {
		if err = param.Output(w, depth); err != nil {
			return err
		}
		if i <= len(self.Params)-1 {
			if err = outputf(w, ", "); err != nil {
				return err
			}
		}
	}
	if err = outputf(w, ") -> "); err != nil {
		return err
	}
	if err = self.Ret.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, " "); err != nil {
		return err
	}

	return self.Body.Output(w, depth)
}
