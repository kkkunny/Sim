package local

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// MultiVarDef 多变量定义
type MultiVarDef struct {
	vars  []values.VarDecl
	value hir.Value
}

func NewMultiVarDef(decls []values.VarDecl, v hir.Value) *MultiVarDef {
	return &MultiVarDef{
		vars:  decls,
		value: v,
	}
}

func (self *MultiVarDef) Local() {
	return
}

func (self *MultiVarDef) Value() hir.Value {
	return self.value
}

func (self *MultiVarDef) Vars() []values.VarDecl {
	return self.vars
}
