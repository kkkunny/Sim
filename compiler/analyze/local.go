package analyze

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
)

func (a *Analyzer) analyzeBlock(block *ast.Block) *hir.Block {
	hirBlock := &hir.Block{}
	for _, stmt := range block.Stmts {
		hirBlock.Stmts = append(hirBlock.Stmts, a.analyzeLocal(stmt))
	}
	return hirBlock
}

func (a *Analyzer) analyzeLocal(local ast.Local) hir.Local {
	switch local := local.(type) {
	case *ast.Block:
		return a.analyzeBlock(local)
	case *ast.Return:
		return a.analyzeReturn(local)
	case ast.Expr:
		return a.analyzeExpr(local)
	case *ast.Let:
		return a.analyzeLet(local)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeReturn(ret *ast.Return) *hir.Return {
	var value optional.Optional[hir.Expr]
	if v, ok := ret.Value.Value(); ok {
		value = optional.Some(a.analyzeExpr(v))
	}
	return &hir.Return{
		Value: value,
	}
}

func (a *Analyzer) analyzeLet(l *ast.Let) *hir.Let {
	value := a.analyzeExpr(l.Value)
	a.scope.types[l.Name.OriginText] = value.GetType()
	return &hir.Let{
		Name:  l.Name.OriginText,
		Value: value,
	}
}
