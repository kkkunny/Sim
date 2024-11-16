package ast

import (
	"io"

	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/token"
)

// Type 类型
type Type interface {
	Ast
	typ()
}

// IdentType 标识符类型
type IdentType Ident

func (self *IdentType) Position() reader.Position {
	return (*Ident)(self).Position()
}

func (self *IdentType) typ() {}

func (self *IdentType) Output(w io.Writer, depth uint) (err error) {
	return (*Ident)(self).Output(w, depth)
}

// FuncType 函数类型
type FuncType struct {
	Begin  reader.Position
	Params []Type
	Ret    optional.Optional[Type]
	End    reader.Position
}

func (self *FuncType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *FuncType) typ() {}

func (self *FuncType) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "func ("); err != nil {
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
	if err = outputf(w, ")"); err != nil {
		return err
	}
	if ret, ok := self.Ret.Value(); ok {
		if err = ret.Output(w, depth); err != nil {
			return err
		}
	}
	return nil
}

// ArrayType 数组类型
type ArrayType struct {
	Begin reader.Position
	Size  token.Token
	Elem  Type
}

func (self *ArrayType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Elem.Position())
}

func (self *ArrayType) typ() {}

func (self *ArrayType) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "["); err != nil {
		return err
	}
	if err = outputf(w, self.Size.Source()); err != nil {
		return err
	}
	if err = outputf(w, "]"); err != nil {
		return err
	}
	return self.Elem.Output(w, depth)
}

// TupleType 元组类型
type TupleType struct {
	Begin reader.Position
	Elems []Type
	End   reader.Position
}

func (self *TupleType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *TupleType) typ() {}

func (self *TupleType) Output(w io.Writer, depth uint) (err error) {
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

// RefType 引用类型
type RefType struct {
	Begin reader.Position
	Mut   bool
	Elem  Type
}

func (self *RefType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Elem.Position())
}

func (self *RefType) typ() {}

func (self *RefType) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "&"); err != nil {
		return err
	}
	if self.Mut {
		if err = outputf(w, "mut "); err != nil {
			return err
		}
	}
	return self.Elem.Output(w, depth)
}

type StructType struct {
	Begin  reader.Position
	Fields []Field
	End    reader.Position
}

func (self *StructType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *StructType) typ() {}

func (self *StructType) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "struct {\n"); err != nil {
		return err
	}
	for i, field := range self.Fields {
		if err = outputDepth(w, depth+1); err != nil {
			return err
		}
		if err = field.Output(w, depth+1); err != nil {
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

// LambdaType 匿名函数类型
type LambdaType struct {
	Begin  reader.Position
	Params []Type
	Ret    Type
}

func (self *LambdaType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Ret.Position())
}

func (self *LambdaType) typ() {}

func (self *LambdaType) Output(w io.Writer, depth uint) (err error) {
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
	return self.Ret.Output(w, depth)
}

type EnumField struct {
	Name token.Token
	Elem optional.Optional[Type]
}

func (self *EnumField) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "%s", self.Name.Source()); err != nil {
		return err
	}
	if elem, ok := self.Elem.Value(); ok {
		if err = outputf(w, ": "); err != nil {
			return err
		}
		if err = elem.Output(w, depth); err != nil {
			return err
		}
	}
	return nil
}

// EnumType 枚举类型
type EnumType struct {
	Begin  reader.Position
	Fields []EnumField
	End    reader.Position
}

func (self *EnumType) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *EnumType) typ() {}

func (self *EnumType) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "enum {\n"); err != nil {
		return err
	}
	for i, field := range self.Fields {
		if err = outputDepth(w, depth+1); err != nil {
			return err
		}
		if err = field.Output(w, depth+1); err != nil {
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
