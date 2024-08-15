package types

import (
	"fmt"
)

// ArrayType 数组类型
type ArrayType struct {
	elem Type
	size uint
}

func NewArrayType(e Type, s uint) *ArrayType {
	return &ArrayType{
		elem: e,
		size: s,
	}
}

func (self *ArrayType) String() string {
	return fmt.Sprintf("[%d]%s", self.size, self.elem.String())
}

func (self *ArrayType) Equal(dst Type) bool {
	t, ok := dst.(*ArrayType)
	return ok && self.size == t.size && self.elem.Equal(t.elem)
}

func (self *ArrayType) Elem() Type {
	return self.elem
}

func (self *ArrayType) Size() uint {
	return self.size
}
