package codegen

import (
	"fmt"
	"math/big"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
	"github.com/kkkunny/Sim/compiler/hir/types"
)

func (c *CodeGenerator) genExpr(expr stmts.Expr) cir.Expr {
	switch expr := expr.(type) {
	case *stmts.IdentExpr:
		return c.genIdentExpr(expr)
	case *stmts.Integer:
		return cir.NewInteger(expr.Value)
	case *stmts.Float:
		return &cir.FloatExpr{Value: expr.Value}
	case *stmts.Boolean:
		return stlval.If(expr.Value, cir.True, cir.False)
	case stmts.Unary:
		return c.genUnary(expr)
	case *stmts.Binary:
		return c.genBinary(expr)
	case *stmts.Func:
		return c.genFunc(expr)
	case *stmts.Call:
		return c.genCall(expr)
	case *stmts.Tuple:
		return c.genTuple(expr)
	case *stmts.TupleIndex:
		from := c.genExpr(expr.From)
		return c.buildTupleIndex(from, expr.Index)
	case *stmts.Array:
		return c.genArray(expr)
	case *stmts.ArrayIndex:
		from := c.genExpr(expr.From)
		offset := c.genExpr(expr.Index)
		return c.buildArrayIndex(from, offset)
	case stmts.Covert:
		return c.genCovert(expr)
	case *stmts.Ternary:
		return c.genTernary(expr)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) genIdentExpr(expr *stmts.IdentExpr) cir.Expr {
	name := c.idents[expr.Define].GetName()
	var value cir.Expr = cir.NewIdentExpr(name)

	// 如果是闭包中捕获的外部变量，转换成ctx的成员变量
	if c.currentFunc != nil {
		if cv, ok := c.captureVarsMap[tuple.Pack2(c.currentFunc, expr.Define)]; ok {
			value = cv
		}
	}

	if let, ok := expr.Define.(*stmts.Let); ok && stlval.Is[*stmts.Func](let.Value) {
		return cir.NewMacroExpr("FUNC_EXPR_F", value)
	}
	return value
}

func (c *CodeGenerator) genUnary(expr stmts.Unary) cir.Expr {
	switch expr := expr.(type) {
	case *stmts.BitsReverse, *stmts.BooleanReverse:
		return cir.NewUnary(cir.UnaryOpEnum.Not, c.genExpr(expr.GetOpTarget()))
	case *stmts.GetRef:
		v := c.genExpr(expr.Target)
		return cir.NewMacroExpr("GET_PTR", v)
	case *stmts.DeRef:
		v := c.genExpr(expr.Target)
		return cir.NewMacroExpr("DE_PTR", v)
	default:
		panic("unreachable")
	}
}

var AssignOp2BinaryOp = map[stmts.BinaryOp]cir.BinaryOp{
	stmts.BinaryOpEnum.AddAssign: cir.BinaryOpEnum.Add,
	stmts.BinaryOpEnum.SubAssign: cir.BinaryOpEnum.Sub,
	stmts.BinaryOpEnum.MulAssign: cir.BinaryOpEnum.Mul,
	stmts.BinaryOpEnum.QuoAssign: cir.BinaryOpEnum.Quo,
	stmts.BinaryOpEnum.RemAssign: cir.BinaryOpEnum.Rem,
	stmts.BinaryOpEnum.AndAssign: cir.BinaryOpEnum.And,
	stmts.BinaryOpEnum.OrAssign:  cir.BinaryOpEnum.Or,
	stmts.BinaryOpEnum.XorAssign: cir.BinaryOpEnum.Xor,
	stmts.BinaryOpEnum.ShlAssign: cir.BinaryOpEnum.Shl,
	stmts.BinaryOpEnum.ShrAssign: cir.BinaryOpEnum.Shr,
}

func (c *CodeGenerator) genBinary(expr *stmts.Binary) cir.Expr {
	left, right := c.genExpr(expr.Left), c.genExpr(expr.Right)

	op, ok := AssignOp2BinaryOp[expr.Op]
	if ok {
		return cir.NewAssign(left, cir.NewBinary(op, left, right))
	} else if expr.Op == stmts.BinaryOpEnum.Assign {
		return cir.NewAssign(left, right)
	}

	switch expr.Op {
	case stmts.BinaryOpEnum.Add:
		op = cir.BinaryOpEnum.Add
	case stmts.BinaryOpEnum.Sub:
		op = cir.BinaryOpEnum.Sub
	case stmts.BinaryOpEnum.Mul:
		op = cir.BinaryOpEnum.Mul
	case stmts.BinaryOpEnum.Quo:
		op = cir.BinaryOpEnum.Quo
	case stmts.BinaryOpEnum.Rem:
		op = cir.BinaryOpEnum.Rem
	case stmts.BinaryOpEnum.And:
		op = cir.BinaryOpEnum.And
	case stmts.BinaryOpEnum.Or:
		op = cir.BinaryOpEnum.Or
	case stmts.BinaryOpEnum.Xor:
		op = cir.BinaryOpEnum.Xor
	case stmts.BinaryOpEnum.Shl:
		op = cir.BinaryOpEnum.Shl
	case stmts.BinaryOpEnum.Shr:
		op = cir.BinaryOpEnum.Shr
	case stmts.BinaryOpEnum.Eq:
		return c.genEqual(false, expr.Left.GetType(), left, right)
	case stmts.BinaryOpEnum.Neq:
		return c.genEqual(true, expr.Left.GetType(), left, right)
	case stmts.BinaryOpEnum.Lt:
		op = cir.BinaryOpEnum.Lt
	case stmts.BinaryOpEnum.Lte:
		op = cir.BinaryOpEnum.Lte
	case stmts.BinaryOpEnum.Gt:
		op = cir.BinaryOpEnum.Gt
	case stmts.BinaryOpEnum.Gte:
		op = cir.BinaryOpEnum.Gte
	case stmts.BinaryOpEnum.LogicAnd:
		op = cir.BinaryOpEnum.LogicAnd
	case stmts.BinaryOpEnum.LogicOr:
		op = cir.BinaryOpEnum.LogicOr
	default:
		panic("unreachable")
	}
	return cir.NewBinary(op, left, right)
}

func (c *CodeGenerator) genNativeFuncDecl(expr *stmts.Func) *cir.FuncDecl {
	returnType := c.genType(expr.Type.GetReturn())
	params := make([]*cir.Param, len(expr.Params))
	for i, p := range expr.Params {
		pn := fmt.Sprintf("_p%d", i+1)
		pt := c.genType(p.Type)
		params[i] = cir.NewParam(pn, pt)
		c.idents[p] = params[i]
	}
	return cir.BuildStmt(c.builder, cir.NewFuncDecl("", returnType, params...))
}

func (c *CodeGenerator) genNativeClosureFunc(expr *stmts.Func, captureVars []stmts.Ident) (*cir.Typedef, *cir.FuncExpr) {
	// 上下文
	fields := stlslices.Map(captureVars, func(i int, vexpr stmts.Ident) *cir.Member {
		fn := fmt.Sprintf("_f%d", i+1)
		c.captureVarsMap[tuple.Pack2(expr, vexpr)] = cir.NewGetMember(cir.NewIdentExpr("_ctx"), fn)
		return cir.NewMember(c.genType(vexpr.GetType()), fn)
	})
	ctxT := cir.BuildStmt(c.builder, cir.NewTypedef(cir.NewStructType("", optional.Some(fields)), ""))

	// 函数
	params := make([]*cir.Param, len(expr.Params)+1)
	params[0] = cir.NewParam("_p0", cir.VoidPtr)
	for i, p := range expr.Params {
		pn := fmt.Sprintf("_p%d", i+1)
		pt := c.genType(p.Type)
		params[i+1] = cir.NewParam(pn, pt)
		c.idents[p] = params[i+1]
	}

	returnType := c.genType(expr.Type.GetReturn())

	initBodyFn := func() {
		ctx := cir.BuildStmt(c.builder, cir.NewVarDecl(cir.NewAliasType(ctxT), "_ctx"))
		ctx.Value = optional.Some[cir.Expr](cir.NewUnary(cir.UnaryOpEnum.Mul, cir.NewCovert(cir.NewPointerType(cir.NewAliasType(ctxT)), cir.NewIdentExpr("_p0"))))
	}

	var body optional.Optional[*cir.Block]
	if b, ok := expr.Body.Value(); ok {
		prevFunc := c.currentFunc
		c.currentFunc = expr
		body = optional.Some(c.genFuncBlock(b, initBodyFn))
		c.currentFunc = prevFunc
	}

	decl := cir.BuildStmt(c.builder, cir.NewFuncDecl("", returnType, params...))
	decl.Body = body
	return ctxT, cir.NewFuncExpr(decl)
}

func (c *CodeGenerator) genFunc(expr *stmts.Func) *cir.MacroExpr {
	captureVars := stlslices.Filter(expr.UsedExternalVariables, func(i int, v stmts.Ident) bool {
		return !stlval.Is[stmts.Global](v)
	})
	if len(captureVars) == 0 {
		decl := c.genNativeFuncDecl(expr)
		if b, ok := expr.Body.Value(); ok {
			prevFunc := c.currentFunc
			c.currentFunc = expr
			decl.Body = optional.Some(c.genFuncBlock(b, nil))
			c.currentFunc = prevFunc
		}
		return cir.NewMacroExpr("FUNC_EXPR_F", &cir.IdentExpr{Name: decl.Name})
	} else {
		ctxT, f := c.genNativeClosureFunc(expr, captureVars)
		fields := make(map[string]cir.Expr, len(captureVars))
		for i, cv := range captureVars {
			fields[fmt.Sprintf("_f%d", i+1)] = c.genExpr(stmts.NewIdentExpr(cv))
		}
		ctx := cir.BuildStmt(c.builder, cir.NewVarDecl(cir.NewAliasType(ctxT), "", cir.NewStruct(fields)))
		return cir.NewMacroExpr("FUNC_EXPR_C", &cir.IdentExpr{Name: f.Decl.Name}, cir.NewUnary(cir.UnaryOpEnum.AND, cir.NewIdentExpr(ctx.GetName())))
	}
}

func (c *CodeGenerator) genCall(expr *stmts.Call) cir.Expr {
	f := c.genExpr(expr.Func)
	args := stlslices.Map(expr.Args, func(_ int, argExpr stmts.Expr) cir.Expr {
		return c.genExpr(argExpr)
	})
	if macroF, ok := f.(*cir.MacroExpr); ok && macroF.Name == "FUNC_EXPR_F" {
		return cir.NewCall(macroF.Args[0].(cir.Expr), args...)
	} else if ok && macroF.Name == "FUNC_EXPR_C" {
		return cir.NewCall(macroF.Args[0].(cir.Expr), append([]cir.Expr{macroF.Args[1].(cir.Expr)}, args...)...)
	}
	call := cir.NewMacroExpr("FUNC_CALL", append([]any{f}, stlslices.AsAny(args)...)...)
	return call
}

func (c *CodeGenerator) genTuple(expr *stmts.Tuple) *cir.Struct {
	t := c.genType(expr.Type)
	if len(expr.Elems) == 0 {
		return cir.NewStruct(nil, t)
	}

	fields := make(map[string]cir.Expr, len(expr.Elems))
	for i, e := range expr.Elems {
		fields[fmt.Sprintf("_f%d", i+1)] = c.genExpr(e)
	}
	return cir.NewStruct(fields, t)
}

func (c *CodeGenerator) genArray(expr *stmts.Array) *cir.Struct {
	at := c.genType(expr.GetType())
	if len(expr.Elems) == 0 {
		return cir.NewStruct(nil, at)
	}

	elems := stlslices.Map(expr.Elems, func(_ int, argExpr stmts.Expr) cir.Expr {
		return c.genExpr(argExpr)
	})
	return cir.NewStruct(map[string]cir.Expr{
		"array": cir.NewArray(elems),
	}, at)
}

func (c *CodeGenerator) genCovert(expr stmts.Covert) cir.Expr {
	t := c.genType(expr.GetType())
	v := c.genExpr(expr.GetFrom())
	switch expr := expr.(type) {
	case *stmts.Union:
		return cir.NewStruct(map[string]cir.Expr{
			"t": cir.NewInteger(big.NewInt(int64(expr.Index))),
			"v": cir.NewStruct(map[string]cir.Expr{
				fmt.Sprintf("t%d", expr.Index+1): v,
			}),
		}, t)
	case *stmts.NumberCovert:
		return cir.NewCovert(t, v)
	case *stmts.TypedefCovert:
		return cir.NewCovert(t, v)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) genTernary(expr *stmts.Ternary) *cir.Ternary {
	cond := c.genExpr(expr.Condition)
	trueExpr := c.genExpr(expr.TrueExpr)
	falseExpr := c.genExpr(expr.FalseExpr)
	return cir.NewTernaryExpr(cond, trueExpr, falseExpr)
}

func (c *CodeGenerator) genEqual(not bool, t types.Type, left, right cir.Expr) cir.Expr {
	switch t := t.(type) {
	case types.CustomType:
		return c.genEqual(not, t.GetUnderlying(), left, right)
	case types.NumberType, types.BooleanType, types.RefType:
		return cir.NewBinary(stlval.If(!not, cir.BinaryOpEnum.Eq, cir.BinaryOpEnum.Neq), left, right)
	case types.ArrayType:
		at := c.genType(t)
		f := cir.BuildStmt(c.builder, cir.NewFuncDecl("", cir.Bool, cir.NewParam("x", at), cir.NewParam("y", at)))
		f.Body = optional.Some(c.buildFuncBlock(func() {
			init := cir.BuildStmt(c.builder, cir.NewVarDecl(cir.I64, "", cir.NewInteger(big.NewInt(0))))
			cond := cir.NewBinary(cir.BinaryOpEnum.Lt, cir.NewIdentExpr(init.Name), cir.NewInteger(t.GetSize()))
			action := cir.NewUnary(cir.UnaryOpEnum.SelfAdd, cir.NewIdentExpr(init.Name))
			loopBlock := c.buildBlock(true, func() {
				lv := c.buildArrayIndex(cir.NewIdentExpr("x"), cir.NewIdentExpr(init.Name))
				rv := c.buildArrayIndex(cir.NewIdentExpr("y"), cir.NewIdentExpr(init.Name))
				ifcond := c.genEqual(true, t.GetElem(), lv, rv)
				ifBlock := c.buildBlock(true, func() {
					cir.BuildStmt(c.builder, cir.NewReturn(stlval.If(!not, cir.False, cir.True)))
				})
				cir.BuildStmt(c.builder, cir.NewIf(ifcond, ifBlock))
			})
			cir.BuildStmt(c.builder, cir.NewFor(optional.None[*cir.VarDecl](), optional.Some[cir.Expr](cond), optional.Some[cir.Expr](action), loopBlock))
			cir.BuildStmt(c.builder, cir.NewReturn(stlval.If(!not, cir.True, cir.False)))
		}))
		return cir.NewCall(cir.NewIdentExpr(f.Name), left, right)
	case types.TupleType:
		tt := c.genType(t)
		f := cir.BuildStmt(c.builder, cir.NewFuncDecl("", cir.Bool, cir.NewParam("x", tt), cir.NewParam("y", tt)))
		f.Body = optional.Some(c.buildFuncBlock(func() {
			for i, et := range t.GetElems() {
				lv := c.buildTupleIndex(cir.NewIdentExpr("x"), big.NewInt(int64(i)))
				rv := c.buildTupleIndex(cir.NewIdentExpr("y"), big.NewInt(int64(i)))
				ifcond := c.genEqual(true, et, lv, rv)
				ifBlock := c.buildBlock(true, func() {
					cir.BuildStmt(c.builder, cir.NewReturn(stlval.If(!not, cir.False, cir.True)))
				})
				cir.BuildStmt(c.builder, cir.NewIf(ifcond, ifBlock))
			}
			cir.BuildStmt(c.builder, cir.NewReturn(stlval.If(!not, cir.True, cir.False)))
		}))
		return cir.NewCall(cir.NewIdentExpr(f.Name), left, right)
	case types.FuncType:
		ft := c.genType(t)
		f := cir.BuildStmt(c.builder, cir.NewFuncDecl("", cir.Bool, cir.NewParam("x", ft), cir.NewParam("y", ft)))
		f.Body = optional.Some(c.buildFuncBlock(func() {
			cir.BuildStmt(c.builder, cir.NewReturn(cir.NewMacroExpr(stlval.If(!not, "FUNC_EQ", "FUNC_NEQ"), cir.NewIdentExpr("x"), cir.NewIdentExpr("y"))))
		}))
		return cir.NewCall(cir.NewIdentExpr(f.Name), left, right)
	case types.UnionType:
		ut := c.genType(t)
		f := cir.BuildStmt(c.builder, cir.NewFuncDecl("", cir.Bool, cir.NewParam("x", ut), cir.NewParam("y", ut)))
		f.Body = optional.Some(c.buildFuncBlock(func() {
			lvi := c.getUnionTypeIndex(cir.NewIdentExpr("x"))
			rvi := c.getUnionTypeIndex(cir.NewIdentExpr("y"))
			ifcond := c.genEqual(true, types.U8, lvi, rvi)
			ifBlock := c.buildBlock(true, func() {
				cir.BuildStmt(c.builder, cir.NewReturn(stlval.If(!not, cir.False, cir.True)))
			})
			cir.BuildStmt(c.builder, cir.NewIf(ifcond, ifBlock))

			branch := cir.NewSwitch(lvi)
			for i, et := range t.GetElems() {
				cond := cir.NewInteger(big.NewInt(int64(i)))
				body := c.buildBlock(true, func() {
					lv := c.getUnionValueIndex(cir.NewIdentExpr("x"), big.NewInt(int64(i)))
					rv := c.getUnionValueIndex(cir.NewIdentExpr("y"), big.NewInt(int64(i)))
					cir.BuildStmt(c.builder, cir.NewReturn(c.genEqual(not, et, lv, rv)))
				})
				branch.Cases = append(branch.Cases, cir.NewCase(cond, body))
			}
			branch.Default = optional.Some(c.buildBlock(true, func() {
				cir.BuildStmt(c.builder, cir.NewReturn(stlval.If(!not, cir.True, cir.False)))
			}))
			cir.BuildStmt(c.builder, branch)
		}))
		return cir.NewCall(cir.NewIdentExpr(f.Name), left, right)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildArrayIndex(array, index cir.Expr) cir.Expr {
	return cir.NewMacroExpr("ARRAY_INDEX", array, index)
}

func (c *CodeGenerator) buildTupleIndex(tuple cir.Expr, index *big.Int) cir.Expr {
	return cir.NewMacroExpr("TUPLE_INDEX", tuple, cir.NewInteger(index.Add(index, big.NewInt(1))))
}

func (c *CodeGenerator) getUnionTypeIndex(union cir.Expr) cir.Expr {
	return cir.NewMacroExpr("UNION_TYPE_INDEX", union)
}

func (c *CodeGenerator) getUnionValueIndex(union cir.Expr, index *big.Int) cir.Expr {
	return cir.NewMacroExpr("UNION_VALUE_INDEX", union, cir.NewInteger(index.Add(index, big.NewInt(1))))
}
