package analyze

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/report"
)

type Analyzer struct {
	reporter *report.Reporter

	scope hir.Scope
}

func NewAnalyzer(reporter *report.Reporter) *Analyzer {
	return &Analyzer{
		reporter: reporter,
		scope:    hir.NewPkgScope(),
	}
}

func (a *Analyzer) Analyze(program *ast.Program) *hir.Program {
	hirProg := &hir.Program{}
	for _, g := range program.Globals {
		hirProg.Globals = append(hirProg.Globals, a.analyzeGlobal(g))
	}
	return hirProg
}
