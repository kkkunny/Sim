package local

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// SingleVarDef 单变量定义
type SingleVarDef struct {
	values.VarDecl
	value hir.Value
}

func NewSingleVarDef(mut bool, name string, t hir.Type, v hir.Value) *SingleVarDef {
	return &SingleVarDef{
		VarDecl: values.NewVarDecl(mut, name, t),
		value:   v,
	}
}

func (self *SingleVarDef) Value() hir.Value {
	return self.value
}

func (self *SingleVarDef) Local() {}
