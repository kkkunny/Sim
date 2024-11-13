package local

import (
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// MultiVarDef 多变量定义
type MultiVarDef struct {
	vars  []*values.VarDecl
	value values.Value
}

func NewMultiVarDef(decls []*values.VarDecl, v values.Value) *MultiVarDef {
	return &MultiVarDef{
		vars:  decls,
		value: v,
	}
}

func (self *MultiVarDef) local() {
	return
}

func (self *MultiVarDef) Value() values.Value {
	return self.value
}

func (self *MultiVarDef) Vars() []*values.VarDecl {
	return self.vars
}
