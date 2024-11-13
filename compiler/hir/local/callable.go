package local

import (
	"github.com/kkkunny/Sim/compiler/hir/values"
)

// CallableDef 可调用的定义
type CallableDef interface {
	values.Callable
	SetBody(b *Block)
	Body() (*Block, bool)
	Params() []*Param
	Parent() Scope
	GetName() (string, bool)
}
