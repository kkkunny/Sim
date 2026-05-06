package analyze

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode/utf8"

	stlmaps "github.com/kkkunny/stl/container/maps"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"
	"golang.org/x/exp/utf8string"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/locals"
	"github.com/kkkunny/Sim/compiler/hir/scopes"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
	"github.com/kkkunny/Sim/compiler/util"
)

// 带自动转换
func (a *Analyzer) analyzeExpr(expr ast.Expr, expect ...hir.Type) locals.Expr {
	v := a.analyzeStrictExpr(expr, expect...)
	if len(expect) > 0 {
		vt := v.GetType()
		expectType := stlslices.Last(expect)
		if ut, ok := expectType.(types.UnionType); ok {
			for i, e := range ut.GetElems() {
				if vt.Equal(e) {
					return locals.NewUnion(v, ut, uint8(i))
				}
			}
		} else if literal, ok := v.(locals.Literal); ok {
			literal.TryToType(expectType)
		}
	}
	return v
}

// 不带自动转换
func (a *Analyzer) analyzeStrictExpr(expr ast.Expr, expect ...hir.Type) locals.Expr {
	switch expr := expr.(type) {
	case *ast.IdentExpr:
		return a.analyzeIdentExpr(expr)
	case *ast.Integer:
		return a.analyzeInteger(expr, expect...)
	case *ast.Char:
		return a.analyzeChar(expr, expect...)
	case *ast.String:
		return a.analyzeString(expr)
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
	case *ast.Struct:
		return a.analyzeStruct(expr)
	case *ast.Member:
		return a.analyzeMember(expr)
	default:
		panic("unreachable")
	}
}

// 期待类型，两个类型必须完全相同
func (a *Analyzer) expectTypeExpr(expr ast.Expr, expect hir.Type) locals.Expr {
	v := a.analyzeExpr(expr, expect)
	if vt := v.GetType(); !vt.Equal(expect) {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.UnexpectedExpression,
			expect, vt,
		)
	}
	return v
}

// 期待类型，属于指定的类型
func expectTypeExpr[T hir.Type](a *Analyzer, expr ast.Expr, expect ...hir.Type) locals.Expr {
	v := a.analyzeExpr(expr, expect...)
	if vt := v.GetType(); !stlval.Is[T](vt) {
		typename := fmt.Sprintf("%T", stlval.Default[T]())
		typename = strings.TrimSuffix(strings.ToLower(typename), "type")
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.UnexpectedExpressionCategory,
			typename, vt,
		)
	}
	return v
}

func (a *Analyzer) analyzeIdentExpr(expr *ast.IdentExpr) *locals.IdentExpr {
	samePkg := true
	pkg := a.scope
	if pkgAst, ok := expr.Pkg.Value(); ok {
		samePkg = false
		pkg, ok = pkg.LookupPkg(pkgAst.OriginText)
		if !ok {
			a.reporter.Fatalf(
				pkgAst.Position,
				report.Errors.UnknownIdentifier,
				pkgAst.OriginText,
			)
		}
	}

	v, ok := pkg.LookupValue(expr.Name.OriginText)
	if !ok ||
		(!samePkg && !v.Public()) {
		a.reporter.Fatalf(
			expr.Name.Position,
			report.Errors.UnknownIdentifier,
			expr.Name.OriginText,
		)
	}

	if let, ok := v.(*locals.Let); ok && let.Type == nil {
		// 全局变量定义，没有显式定义类型
		// 不可能跨包还没有类型
		v = a.analyzeGlobalLetDef(a.letDef2Ast[let])
	}
	return locals.NewIdentExpr(v)
}

func (a *Analyzer) analyzeInteger(expr *ast.Integer, expect ...hir.Type) locals.Expr {
	v, _ := strconv.ParseInt(expr.Value.OriginText, 10, 64)
	if len(expect) == 0 || stlval.Is[types.IntegerType](stlslices.Last(expect)) {
		return locals.NewInteger(types.I64, big.NewInt(v))
	} else {
		return locals.NewFloat(types.F64, big.NewFloat(float64(v)))
	}
}

func (a *Analyzer) analyzeChar(expr *ast.Char, expect ...hir.Type) locals.Expr {
	charText := expr.Value.OriginText[1 : len(expr.Value.OriginText)-1]
	charText = util.ParseEscapeCharacter(charText, `\'`, `'`)
	chars := utf8string.NewString(charText)
	if !utf8.Valid([]byte(charText)) || chars.RuneCount() != 1 {
		a.reporter.Fatalf(
			expr.Value.Position,
			report.Errors.InvalidChar,
			charText,
		)
	}
	char := chars.At(0)

	if len(expect) == 0 || stlval.Is[types.IntegerType](stlslices.Last(expect)) {
		return locals.NewInteger(types.I32, big.NewInt(int64(char)))
	} else {
		return locals.NewFloat(types.F64, big.NewFloat(float64(char)))
	}
}

func (a *Analyzer) analyzeString(expr *ast.String) locals.Expr {
	strText := expr.Value.OriginText[1 : len(expr.Value.OriginText)-1]
	strText = util.ParseEscapeCharacter(strText, `\"`, `"`)
	if !utf8.Valid([]byte(strText)) {
		a.reporter.Fatalf(
			expr.Value.Position,
			report.Errors.InvalidChar,
			strText,
		)
	}
	return locals.NewString(types.Str, strText)
}

func (a *Analyzer) analyzeUnary(expr *ast.Unary, expect ...hir.Type) locals.Unary {
	switch expr.Op.Kind {
	case token.KindEnum.Not:
		v := a.analyzeExpr(expr.Expr, expect...)
		if vt := v.GetType(); stlval.Is[types.IntegerType](vt) {
			return locals.NewBitReverse(v)
		} else if stlval.Is[types.BooleanType](vt) {
			return locals.NewBooleanReverse(v)
		} else {
			a.reporter.Fatalf(
				expr.Position(),
				report.Errors.UnexpectedExpressionCategory,
				"integer or boolean", vt,
			)
			return nil
		}
	case token.KindEnum.Mul:
		if len(expect) > 0 {
			expect = []hir.Type{types.NewRefType(false, stlslices.Last(expect))}
		}
		v := expectTypeExpr[types.RefType](a, expr.Expr, expect...)
		return locals.NewDeRef(v)
	default:
		panic("unreachable")
	}
}

func (a *Analyzer) analyzeBinary(expr *ast.Binary, expect ...hir.Type) *locals.Binary {
	var left locals.Expr
	var op locals.BinaryOp
	switch expr.Op.Kind {
	case token.KindEnum.Add:
		op = locals.BinaryOpEnum.Add
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Sub:
		op = locals.BinaryOpEnum.Sub
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Mul:
		op = locals.BinaryOpEnum.Mul
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Quo:
		op = locals.BinaryOpEnum.Quo
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Rem:
		op = locals.BinaryOpEnum.Rem
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.And:
		op = locals.BinaryOpEnum.And
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Or:
		op = locals.BinaryOpEnum.Or
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Xor:
		op = locals.BinaryOpEnum.Xor
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Shl:
		op = locals.BinaryOpEnum.Shl
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Shr:
		op = locals.BinaryOpEnum.Shr
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.Eq:
		op = locals.BinaryOpEnum.Eq
		left = a.analyzeExpr(expr.Left, expect...)
	case token.KindEnum.Neq:
		op = locals.BinaryOpEnum.Neq
		left = a.analyzeExpr(expr.Left, expect...)
	case token.KindEnum.Lt:
		op = locals.BinaryOpEnum.Lt
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Lte:
		op = locals.BinaryOpEnum.Lte
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Gt:
		op = locals.BinaryOpEnum.Gt
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.Gte:
		op = locals.BinaryOpEnum.Gte
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.LogicAnd:
		op = locals.BinaryOpEnum.LogicAnd
		left = expectTypeExpr[types.BooleanType](a, expr.Left, expect...)
	case token.KindEnum.LogicOr:
		op = locals.BinaryOpEnum.LogicOr
		left = expectTypeExpr[types.BooleanType](a, expr.Left, expect...)
	case token.KindEnum.Assign:
		op = locals.BinaryOpEnum.Assign
		left = a.analyzeExpr(expr.Left, expect...)
	case token.KindEnum.AddAssign:
		op = locals.BinaryOpEnum.AddAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.SubAssign:
		op = locals.BinaryOpEnum.SubAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.MulAssign:
		op = locals.BinaryOpEnum.MulAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.QuoAssign:
		op = locals.BinaryOpEnum.QuoAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.RemAssign:
		op = locals.BinaryOpEnum.RemAssign
		left = expectTypeExpr[types.NumberType](a, expr.Left, expect...)
	case token.KindEnum.AndAssign:
		op = locals.BinaryOpEnum.AndAssign
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.OrAssign:
		op = locals.BinaryOpEnum.OrAssign
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.XorAssign:
		op = locals.BinaryOpEnum.XorAssign
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.ShlAssign:
		op = locals.BinaryOpEnum.ShlAssign
		left = expectTypeExpr[types.IntegerType](a, expr.Left, expect...)
	case token.KindEnum.ShrAssign:
		op = locals.BinaryOpEnum.ShrAssign
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

	return &locals.Binary{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

func (a *Analyzer) analyzeFunc(expr *ast.Func) *locals.Func {
	scope := scopes.NewBlockScope(a.scope)
	a.scope = scope

	ft := a.analyzeFuncDecl(expr)

	params := stlslices.Map(expr.Params, func(i int, pAst *ast.ParamDecl) *hir.Param {
		p := hir.NewParam(pAst.Mut, ft.GetParams()[i], pAst.Name.OriginText)
		a.scope.AddValue(p)
		return p
	})

	scope.SetFuncType(ft)

	var body optional.Optional[*locals.Block]
	if b, ok := expr.Body.Value(); ok {
		body = optional.Some(a.analyzeBlock(b))
	}

	externalVars := stlslices.DiffTo(a.scope.UsedValues(), stlmaps.Values(a.scope.Values()))

	a.scope, _ = a.scope.Parent()

	f := locals.NewFunc(ft, params...)
	f.Body = body
	f.UsedExternalVariables = externalVars
	return f
}

func (a *Analyzer) analyzeCall(expr *ast.Call) *locals.Call {
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
	args := stlslices.Map(expr.Args, func(i int, expr ast.Expr) locals.Expr {
		return a.analyzeExpr(expr, params[i])
	})
	return &locals.Call{
		Func: f,
		Args: args,
	}
}

func (a *Analyzer) analyzeTuple(expr *ast.Tuple, expect ...hir.Type) locals.Expr {
	if len(expr.Elems) == 1 {
		return a.analyzeExpr(expr.Elems[0], expect...)
	}

	var hasType bool
	expectElemTypes := make([]hir.Type, len(expr.Elems))
	if len(expect) > 0 {
		if tt, ok := stlslices.Last(expect).(types.TupleType); ok && (len(expr.Elems) == 0 || len(expr.Elems) == len(tt.GetElems())) {
			hasType = true
			expectElemTypes = tt.GetElems()
		}
	}

	elems := stlslices.Map(expr.Elems, func(i int, e ast.Expr) locals.Expr {
		var elemExpect []hir.Type
		if et := expectElemTypes[i]; et != nil {
			elemExpect = []hir.Type{et}
		}
		return a.analyzeExpr(e, elemExpect...)
	})

	var t types.TupleType
	if len(elems) == 0 && !hasType {
		t = types.NewTupleType()
	} else if len(elems) == 0 {
		t = types.NewTupleType(expectElemTypes...)
	} else {
		t = types.NewTupleType(stlslices.Map(elems, func(_ int, e locals.Expr) hir.Type {
			return e.GetType()
		})...)
	}

	return locals.NewTuple(t, elems...)
}

func (a *Analyzer) analyzeIndex(expr *ast.Index) locals.Expr {
	from := a.analyzeExpr(expr.From)
	ft := from.GetType()

	if stlval.Is[types.TupleType](ft) {
		index := expectTypeExpr[types.IntegerType](a, expr.Index)
		indexValue, ok := index.(*locals.Integer)
		if !ok {
			a.reporter.Fatalf(
				expr.Index.Position(),
				report.Errors.ExpectedIntegerConstant,
			)
		}
		return locals.NewTupleIndex(from, indexValue.Value)
	}

	at, ok := ft.(types.ArrayType)
	if !ok {
		a.reporter.Fatalf(
			expr.From.Position(),
			report.Errors.UnexpectedExpressionCategory,
			"array", at,
		)
	}
	index := expectTypeExpr[types.IntegerType](a, expr.Index)
	return locals.NewArrayIndex(from, index)
}

func (a *Analyzer) analyzeArray(expr *ast.Array, expect ...hir.Type) *locals.Array {
	var size *big.Int
	var expectElemType hir.Type
	if len(expect) > 0 {
		if at, ok := stlslices.Last(expect).(types.ArrayType); ok && (len(expr.Elems) == 0 || strconv.FormatInt(int64(len(expr.Elems)), 10) == at.GetSize().String()) {
			size = at.GetSize()
			expectElemType = at.GetElem()
		}
	}

	elems := stlslices.Map(expr.Elems, func(i int, e ast.Expr) locals.Expr {
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

	var t types.ArrayType
	if len(elems) == 0 {
		t = types.NewArrayType(size, expectElemType)
	} else {
		t = types.NewArrayType(big.NewInt(int64(len(elems))), expectElemType)
	}
	return locals.NewArray(t, elems...)
}

// 尝试获取类型的零值
func (a *Analyzer) tryGetZeroExpr(t hir.Type) (locals.Expr, bool) {
	switch t := t.(type) {
	case types.IntegerType:
		return locals.NewInteger(t, big.NewInt(0)), true
	case types.FloatType:
		return locals.NewFloat(t, big.NewFloat(0)), true
	case types.BooleanType:
		return locals.NewBoolean(t, false), true
	case types.StringType:
		return locals.NewString(t, ""), true
	case types.FuncType:
		var returnValue locals.Expr
		if !t.GetReturn().Equal(types.Unit) {
			var ok bool
			returnValue, ok = a.tryGetZeroExpr(t.GetReturn())
			if !ok {
				return nil, false
			}
		}
		f := locals.NewFunc(t, stlslices.Map(t.GetParams(), func(i int, pt hir.Type) *hir.Param {
			return hir.NewParam(false, pt, fmt.Sprintf("p%d", i+1))
		})...)
		block := locals.NewBlock()
		if returnValue != nil {
			block.Stmts = append(block.Stmts, locals.NewReturn(returnValue))
		}
		f.Body = optional.Some(block)
		return f, true
	case types.TupleType:
		for _, e := range t.GetElems() {
			if _, ok := a.tryGetZeroExpr(e); !ok {
				return nil, false
			}
		}
		return locals.NewTuple(t), true
	case types.ArrayType:
		if t.GetSize().String() != "0" {
			if _, ok := a.tryGetZeroExpr(t.GetElem()); !ok {
				return nil, false
			}
		}
		return locals.NewArray(t), true
	case types.UnionType:
		v, ok := a.tryGetZeroExpr(t.GetElems()[0])
		if !ok {
			return nil, false
		}
		return locals.NewUnion(v, t, 0), true
	case types.StructType:
		fields := make(map[string]locals.Expr, len(t.GetFields()))
		for _, f := range t.GetFields() {
			v, ok := a.tryGetZeroExpr(f.Type)
			if !ok {
				return nil, false
			}
			fields[f.Name] = v
		}
		return locals.NewStruct(t, fields), true
	default:
		return nil, false
	}
}

// 获取类型的零值
func (a *Analyzer) getZeroExpr(pos reader.Position, t hir.Type) locals.Expr {
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

func (a *Analyzer) analyzeAs(expr *ast.As) locals.Expr {
	to := a.analyzeType(expr.Right)
	v := a.analyzeExpr(expr.Left, to)
	from := v.GetType()

	if from.Equal(to) {
		return v
	}

	switch {
	case stlval.Is[types.NumberType](from) && stlval.Is[types.NumberType](to):
		return locals.NewNumberCovert(v, to)
	case types.GetUnderlying(from).Equal(types.GetUnderlying(to)):
		return locals.NewTypedefCovert(v, to)
	}

	a.reporter.Fatalf(
		expr.Left.Position(),
		report.Errors.InvalidTypeCovert,
		from, to,
	)
	return nil
}

func (a *Analyzer) analyzeBoolean(expr *ast.Boolean) *locals.Boolean {
	return locals.NewBoolean(types.Bool, expr.Value.Kind == token.KindEnum.True)
}

func (a *Analyzer) analyzeGetReference(expr *ast.GetReference, expect ...hir.Type) *locals.GetRef {
	var expectElemType []hir.Type
	if len(expect) > 0 {
		if rt, ok := stlslices.Last(expect).(types.RefType); ok {
			expectElemType = append(expectElemType, rt.PtrTo())
		}
	}
	from := a.analyzeExpr(expr.Value, expectElemType...)

	if from.Temporary() {
		a.reporter.Fatalf(
			expr.Value.Position(),
			report.Errors.MustNotTemporary,
		)
	} else if expr.Mut && !from.Mutable() {
		a.reporter.Fatalf(
			expr.Value.Position(),
			report.Errors.MustMutable,
		)
	}
	return locals.NewGetRef(expr.Mut, from)
}

func (a *Analyzer) analyzeTernary(expr *ast.Ternary, expect ...hir.Type) *locals.Ternary {
	cond := a.expectTypeExpr(expr.Condition, types.Bool)
	trueExpr := a.analyzeExpr(expr.TrueExpr, expect...)
	falseExpr := a.expectTypeExpr(expr.FalseExpr, trueExpr.GetType())
	return locals.NewTernary(cond, trueExpr, falseExpr)
}

func (a *Analyzer) analyzeStruct(expr *ast.Struct) *locals.Struct {
	t := a.analyzeType(expr.Type)
	st, ok := t.(types.StructType)
	if !ok {
		a.reporter.Fatalf(
			expr.Position(),
			report.Errors.UnexpectedExpressionCategory,
			"struct", t,
		)
	}

	fields := make(map[string]locals.Expr, len(expr.Fields))
	for _, f := range expr.Fields {
		field, ok := stlslices.FindFirst(st.GetFields(), func(_ int, sf *types.StructField) bool {
			return f.Name.OriginText == sf.Name
		})
		if !ok {
			a.reporter.Fatalf(
				f.Name.Position,
				report.Errors.UnknownIdentifier,
				f.Name.OriginText,
			)
		}
		if _, ok = fields[f.Name.OriginText]; ok {
			a.reporter.Fatalf(
				f.Name.Position,
				report.Errors.RepeatedIdentifier,
				f.Name.OriginText,
			)
		}
		fields[f.Name.OriginText] = a.expectTypeExpr(f.Value, field.Type)
	}
	return locals.NewStruct(st, fields)
}

func (a *Analyzer) analyzeMember(expr *ast.Member) locals.Expr {
	from := a.analyzeExpr(expr.From)
	fromType := from.GetType()
	for {
		refT, ok := fromType.(types.RefType)
		if !ok {
			break
		}
		fromType = refT.PtrTo()
		from = locals.NewDeRef(from)
	}

	// 绑定
	ct, ok := fromType.(types.CustomType)
	if ok {
		let, ok := a.scope.LookupBind(ct.GetDef(), expr.Name.OriginText)
		if ok {
			return locals.NewGetBind(from, let)
		}
	}

	// 结构体字段
	st, ok := fromType.(types.StructType)
	if !ok {
		a.reporter.Fatalf(
			expr.From.Position(),
			report.Errors.UnexpectedExpressionCategory,
			"struct", fromType,
		)
	}
	field, ok := stlslices.FindFirst(st.GetFields(), func(_ int, f *types.StructField) bool {
		return f.Name == expr.Name.OriginText
	})
	if !ok {
		a.reporter.Fatalf(
			expr.Name.Position,
			report.Errors.UnknownIdentifier,
			expr.Name.OriginText,
		)
	}
	return locals.NewGetField(from, field.Name)
}
