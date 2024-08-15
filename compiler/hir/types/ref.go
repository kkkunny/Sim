package types

import (
	"fmt"
)

// RefType 引用类型
type RefType struct {
	elem Type
}

func NewRefType(e Type) *RefType {
	return &RefType{
		elem: e,
	}
}

func (self *RefType) String() string {
	return fmt.Sprintf("&%s", self.elem.String())
}

func (self *RefType) Equal(dst Type) bool {
	t, ok := dst.(*RefType)
	return ok && self.elem.Equal(t.Elem())
}

func (self *RefType) Elem() Type {
	return self.elem
}
