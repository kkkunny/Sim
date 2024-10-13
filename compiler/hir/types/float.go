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
	F32 FloatType = &_FloatType_{kind: FloatTypeKindFloat}
	F64 FloatType = &_FloatType_{kind: FloatTypeKindDouble}
)

// FloatType 浮点型
type FloatType interface {
	NumType
	SignedType
	Kind() FloatTypeKind
}

type _FloatType_ struct {
	kind FloatTypeKind
}

func (self *_FloatType_) String() string {
	return fmt.Sprintf("f%d", self.kind*8)
}

func (self *_FloatType_) Equal(dst Type) bool {
	t, ok := dst.(FloatType)
	return ok && self.kind == t.Kind()
}

func (self *_FloatType_) Kind() FloatTypeKind {
	return self.kind
}

func (self *_FloatType_) Number() {}
func (self *_FloatType_) Signed() {}
