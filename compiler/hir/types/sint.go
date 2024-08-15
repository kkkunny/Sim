package types

import (
	"fmt"
)

// SintType 有符号整型
type SintType struct {
	kind IntTypeKind
}

var (
	I8    = &SintType{kind: IntTypeKindByte}
	I16   = &SintType{kind: IntTypeKindShort}
	I32   = &SintType{kind: IntTypeKindInt}
	I64   = &SintType{kind: IntTypeKindLong}
	Isize = &SintType{kind: IntTypeKindSize}
)

func (self *SintType) String() string {
	if self.kind == IntTypeKindSize {
		return "isize"
	} else {
		return fmt.Sprintf("i%d", self.kind*8)
	}
}

func (self *SintType) Equal(dst Type) bool {
	t, ok := dst.(*SintType)
	return ok && self.kind == t.kind
}

func (self *SintType) Kind() IntTypeKind {
	return self.kind
}

func (self *SintType) num()    {}
func (self *SintType) signed() {}
