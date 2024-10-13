package types

import (
	"fmt"
)

// ArrayType 数组类型
type ArrayType interface {
	Type
	Elem() Type
	Size() uint
}

type _ArrayType_ struct {
	elem Type
	size uint
}

func NewArrayType(e Type, s uint) ArrayType {
	return &_ArrayType_{
		elem: e,
		size: s,
	}
}

func (self *_ArrayType_) String() string {
	return fmt.Sprintf("[%d]%s", self.size, self.elem.String())
}

func (self *_ArrayType_) Equal(dst Type) bool {
	t, ok := dst.(ArrayType)
	return ok && self.size == t.Size() && self.elem.Equal(t.Elem())
}

func (self *_ArrayType_) Elem() Type {
	return self.elem
}

func (self *_ArrayType_) Size() uint {
	return self.size
}
