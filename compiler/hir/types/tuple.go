package types

import (
	"fmt"
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"
)

// TupleType 元组类型
type TupleType struct {
	elems []Type
}

func NewTupleType(es ...Type) *TupleType {
	return &TupleType{
		elems: es,
	}
}

func (self *TupleType) String() string {
	elems := stlslices.Map(self.elems, func(_ int, e Type) string { return e.String() })
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *TupleType) Equal(dst Type) bool {
	t, ok := dst.(*TupleType)
	if !ok || len(self.elems) != len(t.elems) {
		return false
	}
	return stlslices.All(self.elems, func(i int, e Type) bool {
		return e.Equal(t.elems[i])
	})
}

func (self *TupleType) Elems() []Type {
	return self.elems
}
