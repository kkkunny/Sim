package ast

import (
	"io"
	"strings"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/token"
)

// Global 全局ast
type Global interface {
	Ast
	global()
}

// Import 包导入
type Import struct {
	Begin reader.Position
	Paths []token.Token
	Alias optional.Optional[token.Token]
}

func (self *Import) Position() reader.Position {
	if alias, ok := self.Alias.Value(); ok {
		return reader.MixPosition(self.Begin, alias.Position)
	}
	return reader.MixPosition(self.Begin, stlslices.Last(self.Paths).Position)
}

func (*Import) global() {}

func (self *Import) Output(w io.Writer, depth uint) (err error) {
	if err = outputf(w, "import %s", strings.Join(stlslices.Map(self.Paths, func(_ int, path token.Token) string {
		return path.Source()
	}), "::")); err != nil {
		return err
	}
	if alias, ok := self.Alias.Value(); ok {
		if err = outputf(w, " as "); err != nil {
			return err
		}
		if err = outputf(w, alias.Source()); err != nil {
			return err
		}
	}
	return nil
}

type VariableDef interface {
	Global
	Stmt
	variable()
}

type VarDef struct {
	Mutable bool
	Name    token.Token
	Type    optional.Optional[Type]
}

func (self *VarDef) Output(w io.Writer, depth uint) (err error) {
	if self.Mutable {
		if err = outputf(w, "mut "); err != nil {
			return err
		}
	}
	if err = outputf(w, self.Name.Source()); err != nil {
		return err
	}
	if typ, ok := self.Type.Value(); ok {
		if err = outputf(w, ": "); err != nil {
			return err
		}
		if err = typ.Output(w, depth); err != nil {
			return err
		}
	}
	return nil
}

// SingleVariableDef 单变量定义
type SingleVariableDef struct {
	Attrs  []Attr
	Begin  reader.Position
	Public bool
	Var    VarDef
	Value  optional.Optional[Expr]
}

func (self *SingleVariableDef) Position() reader.Position {
	if v, ok := self.Value.Value(); ok {
		return reader.MixPosition(self.Begin, v.Position())
	} else {
		return reader.MixPosition(self.Begin, self.Var.Type.MustValue().Position())
	}
}

func (*SingleVariableDef) stmt() {}

func (*SingleVariableDef) global() {}

func (*SingleVariableDef) variable() {}

func (self *SingleVariableDef) Output(w io.Writer, depth uint) (err error) {
	for _, attr := range self.Attrs {
		if err = attr.Output(w, depth); err != nil {
			return err
		}
		if err = outputf(w, "\n"); err != nil {
			return err
		}
	}
	if self.Public {
		if err = outputf(w, "pub "); err != nil {
			return err
		}
	}
	if err = outputf(w, "let "); err != nil {
		return err
	}
	if err = self.Var.Output(w, depth); err != nil {
		return err
	}
	if value, ok := self.Value.Value(); ok {
		if err = outputf(w, " = "); err != nil {
			return err
		}
		if err = value.Output(w, depth); err != nil {
			return err
		}
	}
	return nil
}

// MultipleVariableDef 多变量定义
type MultipleVariableDef struct {
	Attrs  []Attr
	Begin  reader.Position
	Public bool
	Vars   []VarDef
	Value  optional.Optional[Expr]
	End    reader.Position
}

func (self *MultipleVariableDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*MultipleVariableDef) stmt() {}

func (*MultipleVariableDef) global() {}

func (*MultipleVariableDef) variable() {}

func (self *MultipleVariableDef) ToSingleList() []*SingleVariableDef {
	return lo.Map(self.Vars, func(item VarDef, _ int) *SingleVariableDef {
		return &SingleVariableDef{
			Attrs:  self.Attrs,
			Begin:  self.Begin,
			Public: self.Public,
			Var:    item,
			Value:  optional.None[Expr](),
		}
	})
}

func (self *MultipleVariableDef) Output(w io.Writer, depth uint) (err error) {
	for _, attr := range self.Attrs {
		if err = attr.Output(w, depth); err != nil {
			return err
		}
		if err = outputf(w, "\n"); err != nil {
			return err
		}
	}
	if self.Public {
		if err = outputf(w, "pub "); err != nil {
			return err
		}
	}
	if err = outputf(w, "let ("); err != nil {
		return err
	}
	for i, v := range self.Vars {
		if err = v.Output(w, depth); err != nil {
			return err
		}
		if i < len(self.Vars)-1 {
			if err = outputf(w, ", "); err != nil {
				return err
			}
		}
	}
	if err = outputf(w, ")"); err != nil {
		return err
	}
	if value, ok := self.Value.Value(); ok {
		if err = outputf(w, " = "); err != nil {
			return err
		}
		if err = value.Output(w, depth); err != nil {
			return err
		}
	}
	return nil
}

// FuncDef 函数定义
type FuncDef struct {
	Attrs    []Attr
	Begin    reader.Position
	Public   bool
	SelfType optional.Optional[token.Token]
	FuncDecl
	Body optional.Optional[*Block]
}

func (self *FuncDef) Position() reader.Position {
	if b, ok := self.Body.Value(); ok {
		return reader.MixPosition(self.Begin, b.Position())
	} else {
		return reader.MixPosition(self.Begin, self.FuncDecl.End)
	}
}

func (*FuncDef) global() {}

func (self *FuncDef) Output(w io.Writer, depth uint) (err error) {
	for _, attr := range self.Attrs {
		if err = attr.Output(w, depth); err != nil {
			return err
		}
		if err = outputf(w, "\n"); err != nil {
			return err
		}
	}
	if self.Public {
		if err = outputf(w, "pub "); err != nil {
			return err
		}
	}
	if selfType, ok := self.SelfType.Value(); ok {
		if err = outputf(w, "("); err != nil {
			return err
		}
		if err = outputf(w, selfType.Source()); err != nil {
			return err
		}
		if err = outputf(w, ") "); err != nil {
			return err
		}
	}
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
		if err = ret.Output(w, depth); err != nil {
			return err
		}
	}
	if body, ok := self.Body.Value(); ok {
		if err = outputf(w, " "); err != nil {
			return err
		}
		if err = body.Output(w, depth); err != nil {
			return err
		}
	}
	return nil
}

// TypeDef 类型定义
type TypeDef struct {
	Begin  reader.Position
	Public bool
	Name   token.Token
	Target Type
}

func (self *TypeDef) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Target.Position())
}

func (*TypeDef) global() {}

func (self *TypeDef) Output(w io.Writer, depth uint) (err error) {
	if self.Public {
		if err = outputf(w, "pub "); err != nil {
			return err
		}
	}
	if err = outputf(w, "type %s ", self.Name.Source()); err != nil {
		return err
	}
	return self.Target.Output(w, depth)
}

// TypeAlias 类型别名
type TypeAlias struct {
	Begin  reader.Position
	Public bool
	Name   token.Token
	Target Type
}

func (self *TypeAlias) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.Target.Position())
}

func (*TypeAlias) global() {}

func (self *TypeAlias) Output(w io.Writer, depth uint) (err error) {
	if self.Public {
		if err = outputf(w, "pub "); err != nil {
			return err
		}
	}
	if err = outputf(w, "type %s = ", self.Name.Source()); err != nil {
		return err
	}
	return self.Target.Output(w, depth)
}

// Trait 特征
type Trait struct {
	Begin   reader.Position
	Public  bool
	Name    token.Token
	Methods []*FuncDecl
	End     reader.Position
}

func (self *Trait) Position() reader.Position {
	return reader.MixPosition(self.Begin, self.End)
}

func (*Trait) global() {}

func (self *Trait) Output(w io.Writer, depth uint) (err error) {
	if self.Public {
		if err = outputf(w, "pub "); err != nil {
			return err
		}
	}
	if err = outputf(w, "trait %s {", self.Name.Source()); err != nil {
		return err
	}
	for _, method := range self.Methods {
		if err = outputDepth(w, depth+1); err != nil {
			return err
		}
		if err = method.Output(w, depth+1); err != nil {
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
