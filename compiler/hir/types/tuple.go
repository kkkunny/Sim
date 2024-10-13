package types

import (
	"fmt"
	"strings"

	stlslices "github.com/kkkunny/stl/container/slices"
)

// TupleType 元组类型
type TupleType interface {
	Type
	Fields() []Type
}

func NewTupleType(es ...Type) TupleType {
	return &_TupleType_{
		elems: es,
	}
}

type _TupleType_ struct {
	elems []Type
}

func (self *_TupleType_) String() string {
	elems := stlslices.Map(self.elems, func(_ int, e Type) string { return e.String() })
	return fmt.Sprintf("(%s)", strings.Join(elems, ", "))
}

func (self *_TupleType_) Equal(dst Type) bool {
	t, ok := dst.(TupleType)
	if !ok || len(self.elems) != len(t.Fields()) {
		return false
	}
	return stlslices.All(self.elems, func(i int, e Type) bool {
		return e.Equal(t.Fields()[i])
	})
}

func (self *_TupleType_) Fields() []Type {
	return self.elems
}
