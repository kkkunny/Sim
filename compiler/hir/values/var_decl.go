package values

import (
	"github.com/kkkunny/Sim/compiler/hir"
)

// VarDecl 变量声明
type VarDecl interface {
	Ident
	SetEscaped(v bool)
	Escaped() bool
}

type __VarDecl__ struct {
	mut     bool
	name    string
	typ     hir.Type
	escaped bool
}

func NewVarDecl(mut bool, name string, t hir.Type) VarDecl {
	return &__VarDecl__{
		mut:  mut,
		name: name,
		typ:  t,
	}
}

func (self *__VarDecl__) Type() hir.Type {
	return self.typ
}

func (self *__VarDecl__) Mutable() bool {
	return self.mut
}

func (self *__VarDecl__) GetName() (string, bool) {
	return self.name, self.name != "_"
}

func (self *__VarDecl__) Storable() bool {
	return true
}

func (self *__VarDecl__) Ident() {}

func (self *__VarDecl__) SetEscaped(v bool) {
	self.escaped = v
}

func (self *__VarDecl__) Escaped() bool {
	return self.escaped
}
