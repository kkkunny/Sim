package types

import (
	"fmt"
)

// UintType 无符号整型
type UintType struct {
	kind IntTypeKind
}

var (
	U8    = &UintType{kind: IntTypeKindByte}
	U16   = &UintType{kind: IntTypeKindShort}
	U32   = &UintType{kind: IntTypeKindInt}
	U64   = &UintType{kind: IntTypeKindLong}
	Usize = &UintType{kind: IntTypeKindSize}
)

func (self *UintType) String() string {
	if self.kind == IntTypeKindSize {
		return "usize"
	} else {
		return fmt.Sprintf("u%d", self.kind*8)
	}
}

func (self *UintType) Equal(dst Type) bool {
	t, ok := dst.(*UintType)
	return ok && self.kind == t.kind
}

func (self *UintType) Kind() IntTypeKind {
	return self.kind
}

func (self *UintType) num()      {}
func (self *UintType) unsigned() {}
