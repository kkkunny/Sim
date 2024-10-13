package local

import (
	"github.com/kkkunny/stl/list"

	"github.com/kkkunny/Sim/compiler/hir/values"
)

// MultiVarDef 多变量定义
type MultiVarDef struct {
	pos   *list.Element[Local]
	vars  []*values.VarDecl
	value values.Value
}

func NewMultiVarDef(decls []*values.VarDecl, v values.Value) *MultiVarDef {
	return &MultiVarDef{
		vars:  decls,
		value: v,
	}
}

func (self *MultiVarDef) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *MultiVarDef) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
}

func (self *MultiVarDef) Value() values.Value {
	return self.value
}

func (self *MultiVarDef) Vars() []*values.VarDecl {
	return self.vars
}
