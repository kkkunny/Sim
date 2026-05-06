package types

import "github.com/kkkunny/Sim/compiler/hir"

type NumberType interface {
	hir.Type
	GetBits() uint8
}

type IntegerType interface {
	NumberType
	GetBits() uint8
	integer()
}
