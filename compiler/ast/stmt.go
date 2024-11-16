package ast

import (
	"io"

	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/token"
)

// Stmt 语句ast
type Stmt interface {
	Ast
	stmt()
}

// Block 代码块
type Block struct {
	Begin reader.Position
	Stmts linkedlist.LinkedList[Stmt]
	End   reader.Position
}

func (self *Block) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*Block) stmt() {}

func (self *Block) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "{\n"); err != nil {
		return err
	}

	for i, iter := 0, self.Stmts.Iterator(); iter.Next(); i++ {
		if err = outputDepth(w, depth+1); err != nil {
			return err
		}
		if err = iter.Value().Output(w, depth+1); err != nil {
			return err
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

// Return 函数返回
type Return struct {
	Begin reader.Position
	Value optional.Optional[Expr]
}

func (self *Return) Position() reader.Position {
	v, ok := self.Value.Value()
	if !ok {
		return self.Begin
	}
	return reader.MixPosition(self.Begin, v.Position())
}

func (*Return) stmt() {}

func (self *Return) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "return"); err != nil {
		return err
	}
	if value, ok := self.Value.Value(); ok {
		if err = outputf(w, " "); err != nil {
			return err
		}
		if err = value.Output(w, depth); err != nil {
			return err
		}
	}
	return nil
}

// IfElse if else
type IfElse struct {
	Begin reader.Position
	Cond  optional.Optional[Expr]
	Body  *Block
	Next  optional.Optional[*IfElse]
}

func (self *IfElse) IsIf() bool {
	return self.Cond.IsSome()
}

func (self *IfElse) IsElse() bool {
	return self.Cond.IsNone()
}

func (self *IfElse) Position() reader.Position {
	if next, ok := self.Next.Value(); !ok {
		return reader.MixPosition(self.Begin, self.Body.Position())
	} else {
		return reader.MixPosition(self.Begin, next.Position())
	}
}

func (*IfElse) stmt() {}

func (self *IfElse) Output(w io.Writer, depth uint) (err error) {
	if cond, ok := self.Cond.Value(); ok {
		if err = outputf(w, "if "); err != nil {
			return err
		}
		if err = cond.Output(w, depth); err != nil {
			return err
		}
		if err = outputf(w, " "); err != nil {
			return err
		}
	}
	if err = self.Body.Output(w, depth); err != nil {
		return err
	}
	if next, ok := self.Next.Value(); ok {
		if err = outputf(w, " else "); err != nil {
			return err
		}
		if err = next.Output(w, depth); err != nil {
			return err
		}
	}
	return nil
}

// While 循环
type While struct {
	Begin reader.Position
	Cond  Expr
	Body  *Block
}

func (self *While) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (*While) stmt() {}

func (self *While) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "for "); err != nil {
		return err
	}
	if err = self.Cond.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, " "); err != nil {
		return err
	}
	return self.Body.Output(w, depth)
}

// Break 跳出循环
type Break struct {
	Token token.Token
}

func (self *Break) Position() reader.Position {
	return self.Token.Position
}

func (*Break) stmt() {}

func (self *Break) Output(w io.Writer, depth uint) (err error) {
	return outputf(w, self.Token.Source())
}

// Continue 下一次循环
type Continue struct {
	Token token.Token
}

func (self *Continue) Position() reader.Position {
	return self.Token.Position
}

func (*Continue) stmt() {}

func (self *Continue) Output(w io.Writer, depth uint) (err error) {
	return outputf(w, self.Token.Source())
}

// For 遍历
type For struct {
	Begin     reader.Position
	CursorMut bool
	Cursor    token.Token
	Iterator  Expr
	Body      *Block
}

func (self *For) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Body.Position())
}

func (*For) stmt() {}

func (self *For) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "for "); err != nil {
		return err
	}
	if self.CursorMut {
		if err = outputf(w, "mut "); err != nil {
			return err
		}
	}
	if err = outputf(w, "%s in ", self.Cursor.Source()); err != nil {
		return err
	}
	if err = self.Iterator.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, " "); err != nil {
		return err
	}
	return self.Body.Output(w, depth)
}

type MatchCaseElem struct {
	Mutable bool
	Name    token.Token
}

func (self *MatchCaseElem) Output(w io.Writer, depth uint) (err error) {
	if self.Mutable {
		if err = outputf(w, "mut "); err != nil {
			return err
		}
	}
	return outputf(w, self.Name.Source())
}

type MatchCase struct {
	Name    token.Token
	Elem    optional.Optional[MatchCaseElem]
	ElemEnd reader.Position
	Body    *Block
}

func (self *MatchCase) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "case "); err != nil {
		return err
	}
	if err = outputf(w, self.Name.Source()); err != nil {
		return err
	}
	if elem, ok := self.Elem.Value(); ok {
		if err = outputf(w, "("); err != nil {
			return err
		}
		if err = elem.Output(w, depth); err != nil {
			return err
		}
		if err = outputf(w, ")"); err != nil {
			return err
		}
	}
	if err = outputf(w, ": \n"); err != nil {
		return err
	}
	for i, iter := 0, self.Body.Stmts.Iterator(); iter.Next(); i++ {
		if err = outputDepth(w, depth+1); err != nil {
			return err
		}
		if err = iter.Value().Output(w, depth+1); err != nil {
			return err
		}
		if i < int(self.Body.Stmts.Length())-1 {
			if err = outputf(w, "\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

// Match 匹配
type Match struct {
	Begin reader.Position
	Value Expr
	Cases []MatchCase
	Other optional.Optional[*Block]
	End   reader.Position
}

func (self *Match) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*Match) stmt() {}

func (self *Match) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "match "); err != nil {
		return err
	}
	if err = self.Value.Output(w, depth); err != nil {
		return err
	}
	if err = outputf(w, " {\n"); err != nil {
		return err
	}
	for _, matchCase := range self.Cases {
		if err = outputDepth(w, depth); err != nil {
			return err
		}
		if err = matchCase.Output(w, depth); err != nil {
			return err
		}
		if err = outputf(w, "\n"); err != nil {
			return err
		}
	}
	if other, ok := self.Other.Value(); ok {
		if err = outputDepth(w, depth); err != nil {
			return err
		}
		if err = outputf(w, "other:\n"); err != nil {
			return err
		}
		for i, iter := 0, other.Stmts.Iterator(); iter.Next(); i++ {
			if err = outputDepth(w, depth+1); err != nil {
				return err
			}
			if err = iter.Value().Output(w, depth+1); err != nil {
				return err
			}
			if err = outputf(w, "\n"); err != nil {
				return err
			}
		}
	}
	if err = outputDepth(w, depth); err != nil {
		return err
	}
	return outputf(w, "}")
}
