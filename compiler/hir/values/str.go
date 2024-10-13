package values

import (
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// String 字符串
type String struct {
	value string
}

func NewString(v string) *String {
	return &String{value: v}
}

func (self *String) Type() types.Type {
	return types.Str
}

func (self *String) Mutable() bool {
	return false
}

func (self *String) Storable() bool {
	return false
}
