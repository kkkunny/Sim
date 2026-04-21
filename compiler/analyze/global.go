package analyze

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

func (a *Analyzer) analyzeGlobal(global ast.Global) stmts.Global {
	switch global := global.(type) {
	case *ast.Let:
		return a.analyzeLet(global, true)
	default:
		panic("unreachable")
	}
}
