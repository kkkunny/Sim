package types

import (
	"fmt"
)

// FloatTypeKind 浮点型类型
type FloatTypeKind uint8

const (
	FloatTypeKindFloat FloatTypeKind = 2 << (iota + 1)
	FloatTypeKindDouble
)

var (
	F32 = newFloatType(FloatTypeKindFloat)
	F64 = newFloatType(FloatTypeKindDouble)
)

// FloatType 浮点型
type FloatType struct {
	kind FloatTypeKind
}

func newFloatType(kind FloatTypeKind) *FloatType {
	return &FloatType{
		kind: kind,
	}
}

func (self *FloatType) String() string {
	return fmt.Sprintf("f%d", self.kind*8)
}

func (self *FloatType) Equal(dst Type) bool {
	t, ok := dst.(*FloatType)
	return ok && self.kind == t.kind
}

func (self *FloatType) Kind() FloatTypeKind {
	return self.kind
}

func (self *FloatType) num()    {}
func (self *FloatType) float()  {}
func (self *FloatType) signed() {}
