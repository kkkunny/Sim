package values

import (
	"math/big"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// Float 浮点数
type Float struct {
	typ   types.FloatType
	value *big.Float
}

func NewFloat(t types.FloatType, v *big.Float) *Float {
	return &Float{
		typ:   t,
		value: v,
	}
}

func (self *Float) Type() hir.Type {
	return self.typ
}

func (self *Float) Mutable() bool {
	return false
}

func (self *Float) Storable() bool {
	return false
}
func (self *Float) Value() *big.Float {
	return self.value
}
