package codegen

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/cir"
	"github.com/kkkunny/Sim/compiler/hir/stmts"
)

func (c *CodeGenerator) buildExpr(expr stmts.Expr) cir.Expr {
	switch expr := expr.(type) {
	case *stmts.IdentExpr:
		return c.buildIdentExpr(expr)
	case *stmts.Integer:
		return cir.NewInteger(expr.Value)
	case *stmts.Float:
		return &cir.FloatExpr{Value: expr.Value}
	case *stmts.Boolean:
		return cir.NewMacroExpr(stlval.If(expr.Value, "true", "false"))
	case stmts.Unary:
		return c.buildUnary(expr)
	case *stmts.Binary:
		return c.buildBinary(expr)
	case *stmts.Func:
		return c.buildFunc(expr)
	case *stmts.Call:
		return c.buildCall(expr)
	case *stmts.Tuple:
		return c.buildTuple(expr)
	case *stmts.TupleIndex:
		return c.buildTupleIndex(expr)
	case *stmts.Array:
		return c.buildArray(expr)
	case *stmts.ArrayIndex:
		return c.buildArrayIndex(expr)
	case stmts.Covert:
		return c.buildCovert(expr)
	case *stmts.Ternary:
		return c.buildTernary(expr)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildIdentExpr(expr *stmts.IdentExpr) cir.Expr {
	name := c.idents[expr.Define].GetName()
	var value cir.Expr = cir.NewIdentExpr(name)

	// 如果是闭包中捕获的外部变量，转换成ctx的成员变量
	if c.currentFunc != nil {
		if cv, ok := c.captureVarsMap[tuple.Pack2(c.currentFunc, expr.Define)]; ok {
			value = cv
		}
	}

	if let, ok := expr.Define.(*stmts.Let); ok && stlval.Is[*stmts.Func](let.Value) {
		return cir.NewMacroExpr("FUNCEXPR_F", value)
	}
	return value
}

func (c *CodeGenerator) buildUnary(expr stmts.Unary) *cir.Unary {
	switch expr := expr.(type) {
	case *stmts.BitReverse, *stmts.BooleanReverse:
		return cir.NewUnary(cir.UnaryOpEnum.Not, c.buildExpr(expr.GetOpTarget()))
	case *stmts.GetRef:
		v := c.buildExpr(expr.Target)
		return cir.NewUnary(cir.UnaryOpEnum.AND, v)
	case *stmts.DeRef:
		v := c.buildExpr(expr.Target)
		return cir.NewUnary(cir.UnaryOpEnum.Mul, v)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildBinary(expr *stmts.Binary) cir.Expr {
	left, right := c.buildExpr(expr.Left), c.buildExpr(expr.Right)
	if strings.Contains(string(expr.Op), "=") {
		var op cir.AssignOp
		switch expr.Op {
		case stmts.BinaryOpEnum.Assign:
			op = cir.AssignOpEnum.Assign
		case stmts.BinaryOpEnum.AddAssign:
			op = cir.AssignOpEnum.AddAssign
		case stmts.BinaryOpEnum.SubAssign:
			op = cir.AssignOpEnum.SubAssign
		case stmts.BinaryOpEnum.MulAssign:
			op = cir.AssignOpEnum.MulAssign
		case stmts.BinaryOpEnum.QuoAssign:
			op = cir.AssignOpEnum.QuoAssign
		case stmts.BinaryOpEnum.RemAssign:
			op = cir.AssignOpEnum.RemAssign
		case stmts.BinaryOpEnum.AndAssign:
			op = cir.AssignOpEnum.AndAssign
		case stmts.BinaryOpEnum.OrAssign:
			op = cir.AssignOpEnum.OrAssign
		case stmts.BinaryOpEnum.XorAssign:
			op = cir.AssignOpEnum.XorAssign
		default:
			panic("unreachable")
		}
		return cir.NewAssign(op, left, right)
	}

	var op cir.BinaryOp
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
	default:
		panic("unreachable")
	}
	return cir.NewBinary(op, left, right)
}

func (c *CodeGenerator) buildNativeFunc(expr *stmts.Func) *cir.FuncExpr {
	params := make([]*cir.Param, len(expr.Params))
	for i, p := range expr.Params {
		pn := fmt.Sprintf("_p%d", i+1)
		pt := c.buildType(p.Type)
		params[i] = cir.NewParam(pn, pt)
		c.idents[p] = params[i]
	}

	returnType := c.buildType(expr.Return)

	var body optional.Optional[*cir.Block]
	if b, ok := expr.Body.Value(); ok {
		prevFunc := c.currentFunc
		c.currentFunc = expr
		prevBlock, _ := c.builder.CurrentAt()
		c.builder.MoveTo(nil)
		body = optional.Some(c.buildBlock(b, nil))
		c.builder.MoveTo(prevBlock)
		c.currentFunc = prevFunc
	}

	decl := c.builder.BuildFuncDecl("", returnType, params)
	decl.Body = body
	return &cir.FuncExpr{Decl: decl}
}

func (c *CodeGenerator) buildNativeClosureFunc(expr *stmts.Func, captureVars []stmts.Ident) (*cir.Typedef, *cir.FuncExpr) {
	// 上下文
	fields := stlslices.Map(captureVars, func(i int, vexpr stmts.Ident) *cir.Member {
		fn := fmt.Sprintf("_f%d", i+1)
		c.captureVarsMap[tuple.Pack2(expr, vexpr)] = cir.NewGetMember(cir.NewIdentExpr("_ctx"), fn)
		return cir.NewMember(c.buildType(vexpr.GetType()), fn)
	})
	ctxT := c.builder.BuildTypedef(cir.NewStructType("", fields...), "")

	// 函数
	params := make([]*cir.Param, len(expr.Params)+1)
	params[0] = cir.NewParam("_p0", cir.VoidPtr)
	for i, p := range expr.Params {
		pn := fmt.Sprintf("_p%d", i+1)
		pt := c.buildType(p.Type)
		params[i+1] = cir.NewParam(pn, pt)
		c.idents[p] = params[i+1]
	}

	returnType := c.buildType(expr.Return)

	initBodyFn := func() {
		ctx := c.builder.BuildVarDecl(cir.NewAliasType(ctxT), "_ctx")
		ctx.Value = optional.Some[cir.Expr](cir.NewUnary(cir.UnaryOpEnum.Mul, cir.NewCovert(cir.NewPointerType(cir.NewAliasType(ctxT)), cir.NewIdentExpr("_p0"))))
	}

	var body optional.Optional[*cir.Block]
	if b, ok := expr.Body.Value(); ok {
		prevFunc := c.currentFunc
		c.currentFunc = expr
		prevBlock, _ := c.builder.CurrentAt()
		c.builder.MoveTo(nil)
		body = optional.Some(c.buildBlock(b, initBodyFn))
		c.builder.MoveTo(prevBlock)
		c.currentFunc = prevFunc
	}

	decl := c.builder.BuildFuncDecl("", returnType, params)
	decl.Body = body
	return ctxT, &cir.FuncExpr{Decl: decl}
}

func (c *CodeGenerator) buildFunc(expr *stmts.Func) *cir.MacroExpr {
	captureVars := stlslices.Filter(expr.UsedExternalVariables, func(i int, v stmts.Ident) bool {
		return !stlval.Is[stmts.Global](v)
	})
	var f *cir.FuncExpr
	if len(captureVars) == 0 {
		f = c.buildNativeFunc(expr)
		return cir.NewMacroExpr("FUNCEXPR_F", &cir.IdentExpr{Name: f.Decl.Name})
	} else {
		var ctxT *cir.Typedef
		ctxT, f = c.buildNativeClosureFunc(expr, captureVars)
		fields := make(map[string]cir.Expr, len(captureVars))
		for i, cv := range captureVars {
			fields[fmt.Sprintf("_f%d", i+1)] = c.buildExpr(stmts.NewIdentExpr(cv))
		}
		ctx := c.builder.BuildVarDecl(cir.NewAliasType(ctxT), "", cir.NewStruct(fields))
		return cir.NewMacroExpr("FUNCEXPR_C", &cir.IdentExpr{Name: f.Decl.Name}, cir.NewUnary(cir.UnaryOpEnum.AND, cir.NewIdentExpr(ctx.GetName())))
	}
}

func (c *CodeGenerator) buildCall(expr *stmts.Call) cir.Expr {
	f := c.buildExpr(expr.Func)
	args := stlslices.Map(expr.Args, func(_ int, argExpr stmts.Expr) cir.Expr {
		return c.buildExpr(argExpr)
	})
	if macroF, ok := f.(*cir.MacroExpr); ok && macroF.Name == "FUNCEXPR_F" {
		return cir.NewCall(macroF.Args[0], args...)
	} else if ok && macroF.Name == "FUNCEXPR_C" {
		return cir.NewCall(macroF.Args[0], append([]cir.Expr{macroF.Args[1]}, args...)...)
	}
	return cir.NewMacroExpr("FUNCCALL", append([]cir.Expr{f}, args...)...)
}

func (c *CodeGenerator) buildTuple(expr *stmts.Tuple) *cir.Struct {
	t := c.buildType(expr.Type)
	if len(expr.Elems) == 0 {
		return cir.NewStruct(nil, t)
	}

	fields := make(map[string]cir.Expr, len(expr.Elems))
	for i, e := range expr.Elems {
		fields[fmt.Sprintf("_f%d", i+1)] = c.buildExpr(e)
	}
	return cir.NewStruct(fields, t)
}

func (c *CodeGenerator) buildTupleIndex(expr *stmts.TupleIndex) *cir.GetMember {
	from := c.buildExpr(expr.From)
	return cir.NewGetMember(from, fmt.Sprintf("_f%d", expr.Index.Int64()+1))
}

func (c *CodeGenerator) buildArray(expr *stmts.Array) *cir.Struct {
	at := c.buildType(expr.GetType())
	if len(expr.Elems) == 0 {
		return cir.NewStruct(nil, at)
	}

	elems := stlslices.Map(expr.Elems, func(_ int, argExpr stmts.Expr) cir.Expr {
		return c.buildExpr(argExpr)
	})
	return cir.NewStruct(map[string]cir.Expr{
		"array": cir.NewArray(elems),
	}, at)
}

func (c *CodeGenerator) buildArrayIndex(expr *stmts.ArrayIndex) *cir.Offset {
	from := c.buildExpr(expr.From)
	offset := c.buildExpr(expr.Index)
	return cir.NewOffset(cir.NewGetMember(from, "array"), offset)
}

func (c *CodeGenerator) buildCovert(expr stmts.Covert) cir.Expr {
	t := c.buildType(expr.GetType())
	v := c.buildExpr(expr.GetFrom())
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
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildTernary(expr *stmts.Ternary) *cir.Ternary {
	cond := c.buildExpr(expr.Condition)
	trueExpr := c.buildExpr(expr.TrueExpr)
	falseExpr := c.buildExpr(expr.FalseExpr)
	return cir.NewTernaryExpr(cond, trueExpr, falseExpr)
}
