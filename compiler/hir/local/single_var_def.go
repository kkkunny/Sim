package local

import (
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// SingleVarDef 单变量定义
type SingleVarDef struct {
	VarDecl
	value values.Value
}

func NewSingleVarDef(decl *VarDecl, v values.Value) *SingleVarDef {
	return &SingleVarDef{
		VarDecl: *decl,
		value:   v,
	}
}

func (self *SingleVarDef) Value() values.Value {
	return self.value
}
