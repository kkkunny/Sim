package local

import (
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// SingleVarDef 单变量定义
type SingleVarDef struct {
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

func (self *SingleVarDef) local() {
	return
}

func (self *SingleVarDef) Value() values.Value {
	return self.value
}

func (self *SingleVarDef) SetEscaped(v bool) {
	self.escaped = v
}

func (self *SingleVarDef) Escaped() bool {
	return self.escaped
}
