package ast

import (
	"io"

	"github.com/kkkunny/stl/container/optional"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/token"
)

// Ast 抽象语法树
type Ast interface {
	Position() reader.Position
	Output(io.Writer, uint) error
}

type Param struct {
	Mutable optional.Optional[token.Token]
	Name    optional.Optional[token.Token]
	Type    Type
}

func (self Param) Position() reader.Position {
	if self.Mutable.IsSome() {
		return reader.MixPosition(self.Mutable.MustValue().Position, self.Type.Position())
	} else if self.Name.IsSome() {
		return reader.MixPosition(self.Name.MustValue().Position, self.Type.Position())
	} else {
		return self.Type.Position()
	}
}

func (self Param) Output(w io.Writer, depth uint) (err error) {
	if mut, ok := self.Mutable.Value(); ok {
		if err = outputf(w, "%s ", mut.Source()); err != nil {
			return err
		}
	}
	if name, ok := self.Name.Value(); ok {
		if err = outputf(w, "%s: ", name.Source()); err != nil {
			return err
		}
	}
	return self.Type.Output(w, depth)
}

type Field struct {
	Public  bool
	Mutable bool
	Name    token.Token
	Type    Type
}

func (self Field) Output(w io.Writer, depth uint) (err error) {
	if self.Public {
		if err = outputf(w, "pub "); err != nil {
			return err
		}
	}
	if self.Mutable {
		if err = outputf(w, "mut "); err != nil {
			return err
		}
	}
	if err = outputf(w, "%s: ", self.Name.Source()); err != nil {
		return err
	}
	return self.Type.Output(w, depth)
}

// Ident 标识符
type Ident struct {
	Pkg         optional.Optional[token.Token]
	Name        token.Token
	GenericArgs optional.Optional[*GenericArgList]
}

func (self *Ident) Position() reader.Position {
	begin := stlval.TernaryAction(self.Pkg.IsNone(), func() reader.Position {
		return self.Name.Position
	}, func() reader.Position {
		return self.Pkg.MustValue().Position
	})
	end := stlval.TernaryAction(self.GenericArgs.IsNone(), func() reader.Position {
		return self.Name.Position
	}, func() reader.Position {
		return self.GenericArgs.MustValue().End
	})
	return reader.MixPosition(begin, end)
}

func (self *Ident) Output(w io.Writer, depth uint) (err error) {
	if pkg, ok := self.Pkg.Value(); ok {
		if err = outputf(w, "%s::", pkg.Source()); err != nil {
			return err
		}
	}
	if err = outputf(w, self.Name.Source()); err != nil {
		return err
	}
	if genericArgs, ok := self.GenericArgs.Value(); ok {
		if err = genericArgs.Output(w, depth); err != nil {
			return err
		}
	}
	return nil
}

type GenericParamList struct {
	Begin  reader.Position
	Params []token.Token
	End    reader.Position
}

func (self *GenericParamList) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *GenericParamList) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "<"); err != nil {
		return err
	}
	for i, param := range self.Params {
		if err = outputf(w, param.Source()); err != nil {
			return err
		}
		if i < len(self.Params)-1 {
			if err = outputf(w, ", "); err != nil {
				return err
			}
		}
	}
	return outputf(w, ">")
}

type GenericArgList struct {
	Begin reader.Position
	Args  []Type
	End   reader.Position
}

func (self *GenericArgList) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *GenericArgList) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "::<"); err != nil {
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
	return outputf(w, ">")
}

// FuncDecl 函数声明
type FuncDecl struct {
	Begin  reader.Position
	Name   token.Token
	Params []Param
	Ret    optional.Optional[Type]
	End    reader.Position
}

func (self *FuncDecl) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (self *FuncDecl) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "func %s(", self.Name.Source()); err != nil {
		return err
	}
	for i, param := range self.Params {
		if err = param.Output(w, depth); err != nil {
			return err
		}
		if i < len(self.Params)-1 {
			if err = outputf(w, ", "); err != nil {
				return err
			}
		}
	}
	if err = outputf(w, ")"); err != nil {
		return err
	}
	if ret, ok := self.Ret.Value(); ok {
		if err = outputf(w, " "); err != nil {
			return err
		}
		if err = ret.Output(w, depth); err != nil {
			return err
		}
		if err = outputf(w, " "); err != nil {
			return err
		}
	}
	return nil
}
