package values

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// Boolean 布尔值
type Boolean struct {
	value bool
}

func NewBoolean(v bool) *Boolean {
	return &Boolean{
		value: v,
	}
}

func (self *Boolean) Type() hir.Type {
	return types.Bool
}

func (self *Boolean) Mutable() bool {
	return false
}

func (self *Boolean) Storable() bool {
	return false
}

func (self *Boolean) Value() bool {
	return self.value
}
