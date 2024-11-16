package values

import (
	"math/big"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// Integer 整数
type Integer struct {
	typ   types.IntType
	value *big.Int
}

func NewInteger(t types.IntType, v *big.Int) *Integer {
	return &Integer{
		typ:   t,
		value: v,
	}
}

func (self *Integer) Type() hir.Type {
	return self.typ
}

func (self *Integer) Mutable() bool {
	return false
}

func (self *Integer) Storable() bool {
	return false
}

func (self *Integer) Value() *big.Int {
	return self.value
}
