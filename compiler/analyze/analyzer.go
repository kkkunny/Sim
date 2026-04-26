package analyze

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/report"
)

type Analyzer struct {
	reporter *report.Reporter

	scope scopes.Scope
}

func NewAnalyzer(reporter *report.Reporter) *Analyzer {
	return &Analyzer{
		reporter: reporter,
		scope:    scopes.NewPkgScope(),
	}
}

func (a *Analyzer) Analyze(program *ast.Program) *stmts.Program {
	for _, g := range program.Globals {
		a.analyzeGlobalDecl(g)
	}

	hirProg := &stmts.Program{}
	for _, g := range program.Globals {
		hirProg.Globals = append(hirProg.Globals, a.analyzeGlobalDef(g))
	}
	return hirProg
}
