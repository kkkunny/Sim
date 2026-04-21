package analyze

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	stlmaps "github.com/kkkunny/stl/container/maps"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (a *Analyzer) analyzeExprWithAutoCovert(expr ast.Expr, expect hir.Type) hir.Expr {
	v := a.analyzeExpr(expr, expect)
	vt := v.GetType()
	if ut, ok := expect.(*hir.UnionType); ok {
		for i, e := range ut.Elems {
			if vt.Equal(e) {
				return hir.NewUnion(v, expect, uint8(i))
			}
		}
	}
	return v
}

func (a *Analyzer) analyzeExpr(expr ast.Expr, expect ...hir.Type) hir.Expr {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		return a.analyzeIdentExpr(expr)
	case *ast.Integer:
		return a.analyzeInteger(expr, expect...)
	case *ast.Unary:
		return a.analyzeUnary(expr)
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
	default:
		panic("unreachable")
	}
}

// 期待类型，两个类型必须完全相同
func (a *Analyzer) expectTypeExpr(expr ast.Expr, expect hir.Type) hir.Expr {
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
func expectTypeExpr[TYPE hir.Type](a *Analyzer, expr ast.Expr) hir.Expr {
	v := a.analyzeExpr(expr)
	if vt := v.GetType(); !stlval.Is[TYPE](vt) {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.UnexpectedExpressionType,
			fmt.Sprintf("%T", stlval.Default[TYPE]()), vt,
		)
	}
	return v
}

func (a *Analyzer) analyzeIdentExpr(expr *ast.IdentExpr) *hir.IdentExpr {
	v, ok := a.scope.Lookup(expr.Name.OriginText)
	if !ok {
		a.reporter.Fatalf(
			expr.Name.Position,
			report.Errors.UnknownIdentifier,
			expr.Name.OriginText,
		)
		return nil
	}
	return hir.NewIdentExpr(v)
}

func (a *Analyzer) analyzeInteger(expr *ast.Integer, expect ...hir.Type) hir.Expr {
	var t hir.Type
	it, ok := stlslices.Last(expect).(hir.IntegerType)
	if ok {
		t = it
	} else {
		ft, ok := stlslices.Last(expect).(*hir.FloatType)
		if ok {
			t = ft
		} else {
			t = hir.I32
		}
	}
	v, _ := strconv.ParseInt(expr.Value.OriginText, 10, 64)
	if stlval.Is[hir.IntegerType](t) {
		return hir.NewInteger(t, big.NewInt(v))
	} else {
		return hir.NewFloat(t, big.NewFloat(float64(v)))
	}
}

func (a *Analyzer) analyzeUnary(expr *ast.Unary) *hir.Unary {
	subExpr := a.analyzeExpr(expr.Expr)
	var op hir.UnaryOp
	switch expr.Op.Kind {
	case token.KindEnum.Not:
		op = hir.UnaryOpEnum.Not
	default:
		panic("unreachable")
	}
	return &hir.Unary{
		Op:   op,
		Expr: subExpr,
	}
}

func (a *Analyzer) analyzeBinary(expr *ast.Binary, expect ...hir.Type) *hir.Binary {
	left := a.analyzeExpr(expr.Left, expect...)
	right := a.expectTypeExpr(expr.Right, left.GetType())
	var op hir.BinaryOp
	switch expr.Op.Kind {
	case token.KindEnum.Add:
		op = hir.BinaryOpEnum.Add
	case token.KindEnum.Sub:
		op = hir.BinaryOpEnum.Sub
	case token.KindEnum.Mul:
		op = hir.BinaryOpEnum.Mul
	case token.KindEnum.Quo:
		op = hir.BinaryOpEnum.Quo
	case token.KindEnum.Rem:
		op = hir.BinaryOpEnum.Rem
	case token.KindEnum.And:
		op = hir.BinaryOpEnum.And
	case token.KindEnum.Or:
		op = hir.BinaryOpEnum.Or
	case token.KindEnum.Xor:
		op = hir.BinaryOpEnum.Xor
	case token.KindEnum.Assign:
		op = hir.BinaryOpEnum.Assign
	case token.KindEnum.AddAssign:
		op = hir.BinaryOpEnum.AddAssign
	case token.KindEnum.SubAssign:
		op = hir.BinaryOpEnum.SubAssign
	case token.KindEnum.MulAssign:
		op = hir.BinaryOpEnum.MulAssign
	case token.KindEnum.QuoAssign:
		op = hir.BinaryOpEnum.QuoAssign
	case token.KindEnum.RemAssign:
		op = hir.BinaryOpEnum.RemAssign
	case token.KindEnum.AndAssign:
		op = hir.BinaryOpEnum.AndAssign
	case token.KindEnum.OrAssign:
		op = hir.BinaryOpEnum.OrAssign
	case token.KindEnum.XorAssign:
		op = hir.BinaryOpEnum.XorAssign
	default:
		panic("unreachable")
	}

	if strings.Contains(expr.Op.Kind.String(), "=") {
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

	return &hir.Binary{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

func (a *Analyzer) analyzeFunc(expr *ast.Func) *hir.Func {
	scope := hir.NewBlockScope(a.scope)
	a.scope = scope

	var params []*hir.Param
	for _, p := range expr.Params {
		paramType := a.analyzeType(p.Type)
		params = append(params, hir.NewParam(p.Mut, paramType, p.Name.OriginText))
	}

	var returnType hir.Type = hir.Unit
	if rtAst, ok := expr.ReturnType.Value(); ok {
		returnType = a.analyzeTypeWithUnit(rtAst)
	}

	ft := hir.NewFuncType(returnType, stlslices.Map(params, func(i int, p *hir.Param) hir.Type {
		return p.Type
	})...)
	scope.SetFuncType(ft)

	for _, p := range params {
		a.scope.AddValue(p)
	}

	var body optional.Optional[*hir.Block]
	if b, ok := expr.Body.Value(); ok {
		body = optional.Some(a.analyzeBlock(b))
	}

	externalVars := stlslices.DiffTo(a.scope.UsedValues(), stlmaps.Values(a.scope.Values()))

	a.scope = stlval.IgnoreWith(a.scope.Parent())

	f := hir.NewFunc(returnType, params...)
	f.Body = body
	f.UsedExternalVariables = externalVars
	return f
}

func (a *Analyzer) analyzeCall(expr *ast.Call) *hir.Call {
	f := expectTypeExpr[*hir.FuncType](a, expr.Func)
	ft := f.GetType().(*hir.FuncType)
	if len(ft.Params) != len(expr.Args) {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.InsufficientArguments,
			len(ft.Params), len(expr.Args),
		)
	}
	args := stlslices.Map(expr.Args, func(i int, expr ast.Expr) hir.Expr {
		return a.analyzeExpr(expr, ft.Params[i])
	})
	return &hir.Call{
		Func: f,
		Args: args,
	}
}

func (a *Analyzer) analyzeTuple(expr *ast.Tuple, expect ...hir.Type) hir.Expr {
	if len(expr.Elems) == 1 {
		return a.analyzeExpr(expr.Elems[0], expect...)
	}

	var hasType bool
	expectElemTypes := make([]hir.Type, len(expr.Elems))
	if len(expect) > 0 {
		if tt, ok := stlslices.Last(expect).(*hir.TupleType); ok && (len(expr.Elems) == 0 || len(expr.Elems) == len(tt.Elems)) {
			hasType = true
			expectElemTypes = tt.Elems
		}
	}

	elems := stlslices.Map(expr.Elems, func(i int, e ast.Expr) hir.Expr {
		var elemExpect []hir.Type
		if et := expectElemTypes[i]; et != nil {
			elemExpect = []hir.Type{et}
		}
		return a.analyzeExpr(e, elemExpect...)
	})

	var t *hir.TupleType
	if len(elems) == 0 && !hasType {
		t = hir.NewTupleType()
	} else if len(elems) == 0 {
		t = hir.NewTupleType(expectElemTypes...)
	} else {
		t = hir.NewTupleType(stlslices.Map(elems, func(_ int, e hir.Expr) hir.Type {
			return e.GetType()
		})...)
	}

	return hir.NewTuple(t, elems...)
}

func (a *Analyzer) analyzeIndex(expr *ast.Index) hir.Expr {
	from := a.analyzeExpr(expr.From)
	ft := from.GetType()

	if stlval.Is[*hir.TupleType](ft) {
		index := expectTypeExpr[hir.IntegerType](a, expr.Index)
		indexValue, ok := index.(*hir.Integer)
		if !ok {
			a.reporter.Fatalf(
				expr.Index.Position(),
				report.Errors.ExpectedIntegerConstant,
			)
		}
		return hir.NewTupleIndex(from, indexValue.Value)
	}

	at, ok := ft.(*hir.ArrayType)
	if !ok {
		a.reporter.Fatalf(
			expr.From.Position(),
			report.Errors.UnexpectedExpressionType,
			fmt.Sprintf("%T", hir.ArrayType{}), at,
		)
	}
	index := expectTypeExpr[hir.IntegerType](a, expr.Index)
	return hir.NewArrayIndex(from, index)
}

func (a *Analyzer) analyzeArray(expr *ast.Array, expect ...hir.Type) *hir.Array {
	var size *big.Int
	var expectElemType hir.Type
	if len(expect) > 0 {
		if at, ok := stlslices.Last(expect).(*hir.ArrayType); ok && (len(expr.Elems) == 0 || strconv.FormatInt(int64(len(expr.Elems)), 10) == at.Size.String()) {
			size = at.Size
			expectElemType = at.Elem
		}
	}

	elems := stlslices.Map(expr.Elems, func(i int, e ast.Expr) hir.Expr {
		if i == 0 {
			var elemExpect []hir.Type
			if expectElemType != nil {
				elemExpect = []hir.Type{expectElemType}
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

	var t *hir.ArrayType
	if len(elems) == 0 {
		t = hir.NewArrayType(size, expectElemType)
	} else {
		t = hir.NewArrayType(big.NewInt(int64(len(elems))), expectElemType)
	}
	return hir.NewArray(t, elems...)
}

// 尝试获取类型的零值
func (a *Analyzer) tryGetZeroExpr(t hir.Type) (hir.Expr, bool) {
	switch t := t.(type) {
	case hir.IntegerType:
		return hir.NewInteger(t, big.NewInt(0)), true
	case *hir.FloatType:
		return hir.NewFloat(t, big.NewFloat(0)), true
	case *hir.FuncType:
		var returnValue hir.Expr
		if !t.Return.Equal(hir.Unit) {
			var ok bool
			returnValue, ok = a.tryGetZeroExpr(t.Return)
			if !ok {
				return nil, false
			}
		}
		f := hir.NewFunc(t.Return, stlslices.Map(t.Params, func(i int, pt hir.Type) *hir.Param {
			return hir.NewParam(false, pt, fmt.Sprintf("p%d", i+1))
		})...)
		block := hir.NewBlock()
		if returnValue != nil {
			block.Stmts = append(block.Stmts, hir.NewReturn(returnValue))
		}
		f.Body = optional.Some(block)
		return f, true
	case *hir.TupleType:
		for _, e := range t.Elems {
			if _, ok := a.tryGetZeroExpr(e); !ok {
				return nil, false
			}
		}
		return hir.NewTuple(t), true
	case *hir.ArrayType:
		if t.Size.String() != "0" {
			if _, ok := a.tryGetZeroExpr(t.Elem); !ok {
				return nil, false
			}
		}
		return hir.NewArray(t), true
	case *hir.UnionType:
		v, ok := a.tryGetZeroExpr(t.Elems[0])
		if !ok {
			return nil, false
		}
		return hir.NewUnion(v, t, 0), true
	default:
		return nil, false
	}
}

// 获取类型的零值
func (a *Analyzer) getZeroExpr(pos reader.Position, t hir.Type) hir.Expr {
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

func (a *Analyzer) analyzeAs(expr *ast.As) hir.Expr {
	to := a.analyzeType(expr.Right)
	v := a.analyzeExprWithAutoCovert(expr.Left, to)
	from := v.GetType()

	if from.Equal(to) {
		return v
	}

	switch {
	case stlval.Is[hir.NumberType](from) && stlval.Is[hir.NumberType](to):
		return hir.NewNumberCovert(v, to)
	}

	a.reporter.Fatalf(
		expr.Left.Position(),
		report.Errors.InvalidTypeCovert,
		from, to,
	)
	return nil
}
