package values

import (
	"math/big"

	"github.com/kkkunny/Sim/compiler/hir/types"
)

// Float 浮点数
type Float struct {
	typ   *types._FloatType_
	value *big.Float
}

func NewFloat(t *types._FloatType_, v *big.Float) *Float {
	return &Float{
		typ:   t,
		value: v,
	}
}

func (self *Float) Type() types.Type {
	return self.typ
}

func (self *Float) Mutable() bool {
	return false
}

func (self *Float) Storable() bool {
	return false
}
