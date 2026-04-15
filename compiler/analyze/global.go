package analyze

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (a *Analyzer) analyzeGlobal(global ast.Global) hir.Global {
	switch global := global.(type) {
	case *ast.Let:
		return a.analyzeLet(global)
	default:
		panic("unreachable")
	}
}
