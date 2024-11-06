package local

import (
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// CallableDef 可调用的定义
type CallableDef interface {
	values.Callable
	SetBody(b *Block)
	Block() (*Block, bool)
	Params() []*Param
	Parent() Scope
	GetName() (string, bool)
}

// Param 函数形参
type Param struct {
	values.VarDecl
}

func NewParam(mut bool, name string, t types.Type) *Param {
	return &Param{
		VarDecl: *values.NewVarDecl(mut, name, t),
	}
}
