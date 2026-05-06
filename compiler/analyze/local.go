package analyze

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (a *Analyzer) analyzeBlock(block *ast.Block) *locals.Block {
	hirBlock := locals.NewBlock()
	for _, stmt := range block.Stmts {
		hirBlock.Stmts = append(hirBlock.Stmts, a.analyzeLocal(stmt))
	}
	return hirBlock
}

func (a *Analyzer) analyzeLocal(local ast.Local) locals.Local {
	switch local := local.(type) {
	case *ast.Block:
		return a.analyzeBlock(local)
	case *ast.Return:
		return a.analyzeReturn(local)
	case ast.Expr:
		return a.analyzeExpr(local)
	case *ast.Let:
		return a.analyzeLetDef(local, false)
	case *ast.If:
		return a.analyzeIf(local)
	case *ast.While:
		return a.analyzeWhile(local)
	case *ast.For:
		return a.analyzeFor(local)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeReturn(local *ast.Return) *locals.Return {
	ls := a.scope.(scopes.LocalScope)
	if v, ok := local.Value.Value(); ok {
		value := a.expectTypeExpr(v, ls.FuncType().GetReturn())
		return locals.NewReturn(value)
	} else {
		return locals.NewReturn()
	}
}

func (a *Analyzer) analyzeLetDef(local *ast.Let, isGlobal bool) *locals.Let {
	var t hir.Type
	if tAst, ok := local.Type.Value(); ok {
		t = a.analyzeType(tAst)
	}

	var value optional.Optional[locals.Expr]
	if v, ok := local.Value.Value(); t != nil && ok {
		value = optional.Some(a.expectTypeExpr(v, t))
	} else if t == nil {
		value = optional.Some(a.analyzeExpr(v))
		t = value.MustValue().GetType()
	} else if extern := stlslices.Any(local.Attributes, func(_ int, a ast.Attribute) bool { return stlval.Is[*ast.Extern](a) }); !extern {
		value = optional.Some(a.getZeroExpr(local.Name.Position, t))
	}

	var let *locals.Let
	if decl, ok := a.scope.LookupValue(local.Name.OriginText); ok && isGlobal && stlval.Is[*locals.Let](decl) && decl.(*locals.Let).IsGlobal {
		let = decl.(*locals.Let)
	} else {
		let = &locals.Let{
			Pub:      local.Public,
			IsGlobal: isGlobal,
			Mut:      local.Mut,
			Type:     t,
			Name:     local.Name.OriginText,
		}
		a.scope.AddValue(let)
	}
	let.Value = value
	return let
}

func (a *Analyzer) analyzeIf(local *ast.If) *locals.If {
	cond := a.expectTypeExpr(local.Condition, types.Bool)

	a.scope = scopes.NewBlockScope(a.scope)
	body := a.analyzeBlock(local.Body)
	a.scope, _ = a.scope.Parent()

	var next []either.Either[*locals.If, *locals.Block]
	if nextLocal, ok := local.Else.Value(); ok {
		if elseif, ok := nextLocal.TryLeft(); ok {
			next = append(next, either.Left[*locals.If, *locals.Block](a.analyzeIf(elseif)))
		} else {
			a.scope = scopes.NewBlockScope(a.scope)
			next = append(next, either.Right[*locals.If, *locals.Block](a.analyzeBlock(nextLocal.Right())))
			a.scope, _ = a.scope.Parent()
		}
	}

	return locals.NewIf(cond, body, next...)
}

func (a *Analyzer) analyzeWhile(local *ast.While) *locals.While {
	var cond locals.Expr
	if condAst, ok := local.Condition.Value(); ok {
		cond = a.expectTypeExpr(condAst, types.Bool)
	} else {
		cond = locals.NewBoolean(types.Bool, true)
	}

	a.scope = scopes.NewBlockScope(a.scope)
	body := a.analyzeBlock(local.Body)
	a.scope, _ = a.scope.Parent()

	return locals.NewWhile(cond, body)
}

func (a *Analyzer) analyzeFor(local *ast.For) *locals.For {
	rangv := expectTypeExpr[types.ArrayType](a, local.Range)
	et := rangv.GetType().(types.ArrayType).GetElem()
	param := hir.NewParam(local.Mut, et, local.Variable.OriginText)

	a.scope = scopes.NewBlockScope(a.scope)
	body := a.analyzeBlock(local.Body)
	a.scope, _ = a.scope.Parent()

	return locals.NewFor(param, rangv, body)
}
