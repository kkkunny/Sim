package types

import (
	"fmt"
)

var (
	I8    SintType = &_SintType_{kind: IntTypeKindByte}
	I16   SintType = &_SintType_{kind: IntTypeKindShort}
	I32   SintType = &_SintType_{kind: IntTypeKindInt}
	I64   SintType = &_SintType_{kind: IntTypeKindLong}
	Isize SintType = &_SintType_{kind: IntTypeKindSize}
)

// SintType 有符号整型
type SintType interface {
	IntType
	SignedType
}

type _SintType_ struct {
	kind IntTypeKind
}

func (self *_SintType_) String() string {
	if self.kind == IntTypeKindSize {
		return "isize"
	} else {
		return fmt.Sprintf("i%d", self.kind*8)
	}
}

func (self *_SintType_) Equal(dst Type) bool {
	t, ok := dst.(SintType)
	return ok && self.kind == t.Kind()
}

func (self *_SintType_) Kind() IntTypeKind {
	return self.kind
}

func (self *_SintType_) Number() {}
func (self *_SintType_) Signed() {}
