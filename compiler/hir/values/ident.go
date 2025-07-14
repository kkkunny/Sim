package values

import "github.com/kkkunny/Sim/compiler/hir"

// Ident 标识符
type Ident interface {
	hir.Value
	GetName() (string, bool)
	Ident()
}
