package local

import (
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// SingleVarDef 单变量定义
type SingleVarDef struct {
	values.VarDecl
	value values.Value
}

func NewSingleVarDef(mut bool, name string, t types.Type, v values.Value) *SingleVarDef {
	return &SingleVarDef{
		VarDecl: values.NewVarDecl(mut, name, t),
		value:   v,
	}
}

func (self *SingleVarDef) Value() values.Value {
	return self.value
}

func (self *SingleVarDef) local() {}
