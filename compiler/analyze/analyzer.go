package analyze

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
)

type Analyzer struct {
	scope *Scope
}

type Scope struct {
	parent *Scope
	types  map[string]hir.Type
}

func NewAnalyzer() *Analyzer {
	return &Analyzer{
		scope: &Scope{
			parent: nil,
			types:  make(map[string]hir.Type),
		},
	}
}

func (a *Analyzer) Analyze(program *ast.Program) *hir.Program {
	hirProg := &hir.Program{}
	for _, g := range program.Globals {
		hirProg.Globals = append(hirProg.Globals, a.analyzeGlobal(g))
	}
	return hirProg
}
