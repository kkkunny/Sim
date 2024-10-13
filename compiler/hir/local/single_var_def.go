package local

import (
	"github.com/kkkunny/stl/list"

	"github.com/kkkunny/Sim/compiler/hir/values"
)

// SingleVarDef 单变量定义
type SingleVarDef struct {
	pos *list.Element[Local]
	values.VarDecl
	value   values.Value
	escaped bool
}

func NewSingleVarDef(decl *values.VarDecl, v values.Value) *SingleVarDef {
	return &SingleVarDef{
		VarDecl: *decl,
		value:   v,
	}
}

func (self *SingleVarDef) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *SingleVarDef) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
}

func (self *SingleVarDef) Value() values.Value {
	return self.value
}

func (self *SingleVarDef) Escaped() bool {
	return self.escaped
}
