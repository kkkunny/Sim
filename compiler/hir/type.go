package hir

import (
	"fmt"

	stlcmp "github.com/kkkunny/stl/cmp"
	stlhash "github.com/kkkunny/stl/hash"
)

// Type 类型
type Type interface {
	fmt.Stringer
	stlhash.Hashable
	stlcmp.Equalable[Type]
}
