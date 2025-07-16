package local

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/utils"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// CallableDef 可调用的定义
type CallableDef interface {
	values.Callable
	SetBody(b *Block)
	Body() (*Block, bool)
	Params() []*Param
	Parent() Scope
	GetName() (utils.Name, bool)
}

// Param 函数形参
type Param struct {
	values.VarDecl
}

func NewParam(mut bool, name utils.Name, t hir.Type) *Param {
	return &Param{
		VarDecl: values.NewVarDecl(mut, name, t),
	}
}
