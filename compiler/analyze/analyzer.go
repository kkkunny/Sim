package analyze

import (
	"github.com/kkkunny/stl/container/set"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/report"
)

type Analyzer struct {
	reporter *report.Reporter

	scope scopes.Scope

	typedefAsts map[*stmts.TypeDef]*ast.TypeDef
	// 类型定义栈，用于在类型定义时检测循环引用
	typedefStack set.Set[*stmts.TypeDef]
}

func NewAnalyzer(reporter *report.Reporter) *Analyzer {
	return &Analyzer{
		reporter:     reporter,
		scope:        scopes.NewPkgScope(),
		typedefAsts:  make(map[*stmts.TypeDef]*ast.TypeDef),
		typedefStack: set.StdLinkedHashSetWith[*stmts.TypeDef](),
	}
}

func (a *Analyzer) Analyze(program *ast.Program) *stmts.Program {
	hirProg := &stmts.Program{}

	ts := make([]*ast.TypeDef, 0, len(program.Globals))
	for _, g := range program.Globals {
		t, ok := g.(*ast.TypeDef)
		if !ok {
			continue
		}
		hirProg.Globals = append(hirProg.Globals, a.analyzeTypeDecl(t))
		ts = append(ts, t)
	}
	for _, t := range ts {
		a.analyzeTypeDef(t)
	}

	for _, g := range program.Globals {
		if stlval.Is[*ast.TypeDef](g) {
			continue
		}
		a.analyzeGlobalDecl(g)
	}

	for _, g := range program.Globals {
		if stlval.Is[*ast.TypeDef](g) {
			continue
		}
		hirProg.Globals = append(hirProg.Globals, a.analyzeGlobalDef(g))
	}
	return hirProg
}
