package types

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/hir"
)

type Type interface {
	hir.PrintWriter
	fmt.Stringer
	Equal(Type) bool
}

type NumberType interface {
	Type
	GetBits() uint8
}

type IntegerType interface {
	NumberType
	GetBits() uint8
	integer()
}
