package analyze

import (
	"fmt"
	"math/big"
	"strconv"

	stlmaps "github.com/kkkunny/stl/container/maps"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (a *Analyzer) analyzeExprWithAutoCovert(expr ast.Expr, expect types.Type) stmts.Expr {
	v := a.analyzeExpr(expr, expect)
	vt := v.GetType()
	if ut, ok := expect.(types.UnionType); ok {
		for i, e := range ut.GetElems() {
			if vt.Equal(e) {
				return stmts.NewUnion(v, expect, uint8(i))
			}
		}
	}
	return v
}

func (a *Analyzer) analyzeExpr(expr ast.Expr, expect ...types.Type) stmts.Expr {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		return a.analyzeIdentExpr(expr)
	case *ast.Integer:
		return a.analyzeInteger(expr, expect...)
	case *ast.Unary:
		return a.analyzeUnary(expr, expect...)
	case *ast.Binary:
		return a.analyzeBinary(expr, expect...)
	case *ast.Func:
		return a.analyzeFunc(expr)
	case *ast.Call:
		return a.analyzeCall(expr)
	case *ast.Tuple:
		return a.analyzeTuple(expr, expect...)
	case *ast.Index:
		return a.analyzeIndex(expr)
	case *ast.Array:
		return a.analyzeArray(expr, expect...)
	case *ast.As:
		return a.analyzeAs(expr)
	case *ast.Boolean:
		return a.analyzeBoolean(expr)
	case *ast.GetReference:
		return a.analyzeGetReference(expr, expect...)
	case *ast.Ternary:
		return a.analyzeTernary(expr, expect...)
	default:
		panic("unreachable")
	}
}

// 期待类型，两个类型必须完全相同
func (a *Analyzer) expectTypeExpr(expr ast.Expr, expect types.Type) stmts.Expr {
	v := a.analyzeExprWithAutoCovert(expr, expect)
	if vt := v.GetType(); !vt.Equal(expect) {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.UnexpectedExpression,
			expect, vt,
		)
	}
	return v
}

// 期待类型，属于制定的类型类
func expectTypeExpr[T types.Type](a *Analyzer, expr ast.Expr, expect ...types.Type) stmts.Expr {
	v := a.analyzeExpr(expr, expect...)
	if vt := v.GetType(); !stlval.Is[T](vt) {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.UnexpectedExpressionType,
			fmt.Sprintf("%T", stlval.Default[T]()), vt,
		)
	}
	return v
}

func (a *Analyzer) analyzeIdentExpr(expr *ast.IdentExpr) *stmts.IdentExpr {
	v, ok := a.scope.Lookup(expr.Name.OriginText)
	if !ok {
		a.reporter.Fatalf(
			expr.Name.Position,
			report.Errors.UnknownIdentifier,
			expr.Name.OriginText,
		)
		return nil
	}
	return stmts.NewIdentExpr(v)
}

func (a *Analyzer) analyzeInteger(expr *ast.Integer, expect ...types.Type) stmts.Expr {
	var t types.Type
	it, ok := stlslices.Last(expect).(types.IntegerType)
	if ok {
		t = it
	} else {
		ft, ok := stlslices.Last(expect).(types.FloatType)
		if ok {
			t = ft
		} else {
			t = types.I32
		}
	}
	v, _ := strconv.ParseInt(expr.Value.OriginText, 10, 64)
	if stlval.Is[types.IntegerType](t) {
		return stmts.NewInteger(t, big.NewInt(v))
	} else {
		return stmts.NewFloat(t, big.NewFloat(float64(v)))
	}
}

func (a *Analyzer) analyzeUnary(expr *ast.Unary, expect ...types.Type) stmts.Unary {
	switch expr.Op.Kind {
	case token.KindEnum.Not:
		v := a.analyzeExpr(expr.Expr, expect...)
		if vt := v.GetType(); stlval.Is[types.IntegerType](vt) {
			return stmts.NewBitReverse(v)
		} else if stlval.Is[types.BooleanType](vt) {
			return stmts.NewBooleanReverse(v)
		} else {
			a.reporter.Fatalf(
				expr.Position(),
				report.Errors.UnexpectedExpressionType,
				"IntegerType or BooleanType", vt,
			)
			return nil
		}
	case token.KindEnum.Mul:
		if len(expect) > 0 {
			expect = []types.Type{types.NewRefType(false, stlslices.Last(expect))}
		}
		v := expectTypeExpr[types.RefType](a, expr.Expr, expect...)
		return stmts.NewDeRef(v)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeBinary(expr *ast.Binary, expect ...types.Type) *stmts.Binary {
	var left stmts.Expr
	var op stmts.BinaryOp
	switch expr.Op.Kind {
	case token.KindEnum.Add:
		op = stmts.BinaryOpEnum.Add
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Sub:
		op = stmts.BinaryOpEnum.Sub
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Mul:
		op = stmts.BinaryOpEnum.Mul
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Quo:
		op = stmts.BinaryOpEnum.Quo
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Rem:
		op = stmts.BinaryOpEnum.Rem
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.And:
		op = stmts.BinaryOpEnum.And
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Or:
		op = stmts.BinaryOpEnum.Or
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Xor:
		op = stmts.BinaryOpEnum.Xor
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Shl:
		op = stmts.BinaryOpEnum.Shl
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Shr:
		op = stmts.BinaryOpEnum.Shr
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Eq:
		op = stmts.BinaryOpEnum.Eq
		left = a.analyzeExpr(expr.Left, expect...)
	case token.KindEnum.Neq:
		op = stmts.BinaryOpEnum.Neq
		left = a.analyzeExpr(expr.Left, expect...)
	case token.KindEnum.Lt:
		op = stmts.BinaryOpEnum.Lt
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Lte:
		op = stmts.BinaryOpEnum.Lte
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Gt:
		op = stmts.BinaryOpEnum.Gt
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Gte:
		op = stmts.BinaryOpEnum.Gte
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.LogicAnd:
		op = stmts.BinaryOpEnum.LogicAnd
		left = expectTypeExpr[types.BooleanType](a, expr.Left, expect...)
	case token.KindEnum.LogicOr:
		op = stmts.BinaryOpEnum.LogicOr
		left = expectTypeExpr[types.BooleanType](a, expr.Left, expect...)
	case token.KindEnum.Assign:
		op = stmts.BinaryOpEnum.Assign
		left = a.analyzeExpr(expr.Left, expect...)
	case token.KindEnum.AddAssign:
		op = stmts.BinaryOpEnum.AddAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.SubAssign:
		op = stmts.BinaryOpEnum.SubAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.MulAssign:
		op = stmts.BinaryOpEnum.MulAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.QuoAssign:
		op = stmts.BinaryOpEnum.QuoAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.RemAssign:
		op = stmts.BinaryOpEnum.RemAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.AndAssign:
		op = stmts.BinaryOpEnum.AndAssign
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.OrAssign:
		op = stmts.BinaryOpEnum.OrAssign
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.XorAssign:
		op = stmts.BinaryOpEnum.XorAssign
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.ShlAssign:
		op = stmts.BinaryOpEnum.ShlAssign
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.ShrAssign:
		op = stmts.BinaryOpEnum.ShrAssign
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	default:
		panic("unreachable")
	}

	right := a.expectTypeExpr(expr.Right, left.GetType())

	if stlslices.Contain(
		[]token.Kind{
			token.KindEnum.Assign,
			token.KindEnum.AddAssign,
			token.KindEnum.SubAssign,
			token.KindEnum.QuoAssign,
			token.KindEnum.RemAssign,
			token.KindEnum.AndAssign,
			token.KindEnum.OrAssign,
			token.KindEnum.XorAssign,
			token.KindEnum.ShlAssign,
			token.KindEnum.ShrAssign,
		},
		expr.Op.Kind,
	) {
		if left.Temporary() {
			a.reporter.Fatalf(
				expr.Left.Position(),
				report.Errors.MustNotTemporary,
			)
		}
		if !left.Mutable() {
			a.reporter.Fatalf(
				expr.Left.Position(),
				report.Errors.MustMutable,
			)
		}
	}

	return &stmts.Binary{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

func (a *Analyzer) analyzeFunc(expr *ast.Func) *stmts.Func {
	scope := scopes.NewBlockScope(a.scope)
	a.scope = scope

	ft := a.analyzeFuncDecl(expr)

	var params []*stmts.Param
	for i, p := range expr.Params {
		params = append(params, stmts.NewParam(p.Mut, ft.GetParams()[i], p.Name.OriginText))
	}

	scope.SetFuncType(ft)

	for _, p := range params {
		a.scope.AddValue(p)
	}

	var body optional.Optional[*stmts.Block]
	if b, ok := expr.Body.Value(); ok {
		body = optional.Some(a.analyzeBlock(b))
	}

	externalVars := stlslices.DiffTo(a.scope.UsedValues(), stlmaps.Values(a.scope.Values()))

	a.scope, _ = a.scope.Parent()

	f := stmts.NewFunc(ft.GetReturn(), params...)
	f.Body = body
	f.UsedExternalVariables = externalVars
	return f
}

func (a *Analyzer) analyzeCall(expr *ast.Call) *stmts.Call {
	f := expectTypeExpr[types.FuncType](a, expr.Func)
	ft := f.GetType().(types.FuncType)
	params := ft.GetParams()
	if len(params) != len(expr.Args) {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.InsufficientArguments,
			len(params), len(expr.Args),
		)
	}
	args := stlslices.Map(expr.Args, func(i int, expr ast.Expr) stmts.Expr {
		return a.analyzeExpr(expr, params[i])
	})
	return &stmts.Call{
		Func: f,
		Args: args,
	}
}

func (a *Analyzer) analyzeTuple(expr *ast.Tuple, expect ...types.Type) stmts.Expr {
	if len(expr.Elems) == 1 {
		return a.analyzeExpr(expr.Elems[0], expect...)
	}

	var hasType bool
	expectElemTypes := make([]types.Type, len(expr.Elems))
	if len(expect) > 0 {
		if tt, ok := stlslices.Last(expect).(types.TupleType); ok && (len(expr.Elems) == 0 || len(expr.Elems) == len(tt.GetElems())) {
			hasType = true
			expectElemTypes = tt.GetElems()
		}
	}

	elems := stlslices.Map(expr.Elems, func(i int, e ast.Expr) stmts.Expr {
		var elemExpect []types.Type
		if et := expectElemTypes[i]; et != nil {
			elemExpect = []types.Type{et}
		}
		return a.analyzeExpr(e, elemExpect...)
	})

	var t types.TupleType
	if len(elems) == 0 && !hasType {
		t = types.NewTupleType()
	} else if len(elems) == 0 {
		t = types.NewTupleType(expectElemTypes...)
	} else {
		t = types.NewTupleType(stlslices.Map(elems, func(_ int, e stmts.Expr) types.Type {
			return e.GetType()
		})...)
	}

	return stmts.NewTuple(t, elems...)
}

func (a *Analyzer) analyzeIndex(expr *ast.Index) stmts.Expr {
	from := a.analyzeExpr(expr.From)
	ft := from.GetType()

	if stlval.Is[types.TupleType](ft) {
		index := expectTypeExpr[types.IntegerType](a, expr.Index)
		indexValue, ok := index.(*stmts.Integer)
		if !ok {
			a.reporter.Fatalf(
				expr.Index.Position(),
				report.Errors.ExpectedIntegerConstant,
			)
		}
		return stmts.NewTupleIndex(from, indexValue.Value)
	}

	at, ok := ft.(types.ArrayType)
	if !ok {
		a.reporter.Fatalf(
			expr.From.Position(),
			report.Errors.UnexpectedExpressionType,
			"ArrayType", at,
		)
	}
	index := expectTypeExpr[types.IntegerType](a, expr.Index)
	return stmts.NewArrayIndex(from, index)
}

func (a *Analyzer) analyzeArray(expr *ast.Array, expect ...types.Type) *stmts.Array {
	var size *big.Int
	var expectElemType types.Type
	if len(expect) > 0 {
		if at, ok := stlslices.Last(expect).(types.ArrayType); ok && (len(expr.Elems) == 0 || strconv.FormatInt(int64(len(expr.Elems)), 10) == at.GetSize().String()) {
			size = at.GetSize()
			expectElemType = at.GetElem()
		}
	}

	elems := stlslices.Map(expr.Elems, func(i int, e ast.Expr) stmts.Expr {
		if i == 0 {
			var elemExpect []types.Type
			if expectElemType != nil {
				elemExpect = []types.Type{expectElemType}
			}
			v := a.analyzeExpr(e, elemExpect...)
			expectElemType = v.GetType()
			return v
		} else {
			return a.expectTypeExpr(e, expectElemType)
		}
	})

	if size == nil || expectElemType == nil {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.MissingType,
		)
	}

	var t types.ArrayType
	if len(elems) == 0 {
		t = types.NewArrayType(size, expectElemType)
	} else {
		t = types.NewArrayType(big.NewInt(int64(len(elems))), expectElemType)
	}
	return stmts.NewArray(t, elems...)
}

// 尝试获取类型的零值
func (a *Analyzer) tryGetZeroExpr(t types.Type) (stmts.Expr, bool) {
	switch t := t.(type) {
	case types.IntegerType:
		return stmts.NewInteger(t, big.NewInt(0)), true
	case types.FloatType:
		return stmts.NewFloat(t, big.NewFloat(0)), true
	case types.BooleanType:
		return stmts.NewBoolean(false), true
	case types.FuncType:
		var returnValue stmts.Expr
		if !t.GetReturn().Equal(types.Unit) {
			var ok bool
			returnValue, ok = a.tryGetZeroExpr(t.GetReturn())
			if !ok {
				return nil, false
			}
		}
		f := stmts.NewFunc(t.GetReturn(), stlslices.Map(t.GetParams(), func(i int, pt types.Type) *stmts.Param {
			return stmts.NewParam(false, pt, fmt.Sprintf("p%d", i+1))
		})...)
		block := stmts.NewBlock()
		if returnValue != nil {
			block.Stmts = append(block.Stmts, stmts.NewReturn(returnValue))
		}
		f.Body = optional.Some(block)
		return f, true
	case types.TupleType:
		for _, e := range t.GetElems() {
			if _, ok := a.tryGetZeroExpr(e); !ok {
				return nil, false
			}
		}
		return stmts.NewTuple(t), true
	case types.ArrayType:
		if t.GetSize().String() != "0" {
			if _, ok := a.tryGetZeroExpr(t.GetElem()); !ok {
				return nil, false
			}
		}
		return stmts.NewArray(t), true
	case types.UnionType:
		v, ok := a.tryGetZeroExpr(t.GetElems()[0])
		if !ok {
			return nil, false
		}
		return stmts.NewUnion(v, t, 0), true
	default:
		return nil, false
	}
}

// 获取类型的零值
func (a *Analyzer) getZeroExpr(pos reader.Position, t types.Type) stmts.Expr {
	v, ok := a.tryGetZeroExpr(t)
	if !ok {
		a.reporter.Fatalf(
			pos,
			report.Errors.TypeMissingDefaultValue,
			t,
		)
	}
	return v
}

func (a *Analyzer) analyzeAs(expr *ast.As) stmts.Expr {
	to := a.analyzeType(expr.Right)
	v := a.analyzeExprWithAutoCovert(expr.Left, to)
	from := v.GetType()

	if from.Equal(to) {
		return v
	}

	switch {
	case stlval.Is[types.NumberType](from) && stlval.Is[types.NumberType](to):
		return stmts.NewNumberCovert(v, to)
	}

	a.reporter.Fatalf(
		expr.Left.Position(),
		report.Errors.InvalidTypeCovert,
		from, to,
	)
	return nil
}

func (a *Analyzer) analyzeBoolean(expr *ast.Boolean) *stmts.Boolean {
	return stmts.NewBoolean(expr.Value.Kind == token.KindEnum.True)
}

func (a *Analyzer) analyzeGetReference(expr *ast.GetReference, expect ...types.Type) *stmts.GetRef {
	from := a.analyzeExpr(expr.Value)
	if from.Temporary() {
		a.reporter.Fatalf(
			expr.Value.Position(),
			report.Errors.MustNotTemporary,
		)
	}
	return stmts.NewGetRef(expr.Mut, from)
}

func (a *Analyzer) analyzeTernary(expr *ast.Ternary, expect ...types.Type) *stmts.Ternary {
	cond := a.expectTypeExpr(expr.Condition, types.Bool)
	trueExpr := a.analyzeExpr(expr.TrueExpr, expect...)
	falseExpr := a.expectTypeExpr(expr.FalseExpr, trueExpr.GetType())
	return stmts.NewTernary(cond, trueExpr, falseExpr)
}
