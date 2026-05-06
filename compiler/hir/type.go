package hir

import (
	"fmt"
)

type Type interface {
	PrintWriter
	fmt.Stringer
	Equal(Type) bool
}
