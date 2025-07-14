package values

import (
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

// String 字符串
type String struct {
	typ   types.StrType
	value string
}

func NewString(t types.StrType, v string) *String {
	return &String{
		typ:   t,
		value: v,
	}
}

func (self *String) Type() hir.Type {
	return self.typ
}

func (self *String) Mutable() bool {
	return false
}

func (self *String) Storable() bool {
	return false
}
func (self *String) Value() string {
	return self.value
}
