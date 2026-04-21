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
	"github.com/kkkunny/Sim/compiler/hir"
)

func (c *CodeGenerator) buildExpr(expr hir.Expr) cir.Expr {
	switch expr := expr.(type) {
	case *hir.IdentExpr:
		return c.buildIdentExpr(expr)
	case *hir.Integer:
		return cir.NewInteger(expr.Value)
	case *hir.Float:
		return &cir.FloatExpr{Value: expr.Value}
	case *hir.Boolean:
		return cir.NewMacroExpr(stlval.If(expr.Value, "true", "false"))
	case *hir.Unary:
		return c.buildUnary(expr)
	case *hir.Binary:
		return c.buildBinary(expr)
	case *hir.Func:
		return c.buildFunc(expr)
	case *hir.Call:
		return c.buildCall(expr)
	case *hir.Tuple:
		return c.buildTuple(expr)
	case *hir.TupleIndex:
		return c.buildTupleIndex(expr)
	case *hir.Array:
		return c.buildArray(expr)
	case *hir.ArrayIndex:
		return c.buildArrayIndex(expr)
	case hir.Covert:
		return c.buildCovert(expr)
	case *hir.GetRef:
		return c.buildGetRef(expr)
	case *hir.DeRef:
		return c.buildDeRef(expr)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildIdentExpr(expr *hir.IdentExpr) cir.Expr {
	name := c.idents[expr.Define].GetName()
	var value cir.Expr = cir.NewIdentExpr(name)

	// 如果是闭包中捕获的外部变量，转换成ctx的成员变量
	if c.currentFunc != nil {
		if cv, ok := c.captureVarsMap[tuple.Pack2(c.currentFunc, expr.Define)]; ok {
			value = cv
		}
	}

	if let, ok := expr.Define.(*hir.Let); ok && stlval.Is[*hir.Func](let.Value) {
		return cir.NewMacroExpr("FUNCEXPR_F", value)
	}
	return value
}

func (c *CodeGenerator) buildUnary(expr *hir.Unary) *cir.UnaryExpr {
	var op cir.UnaryOp
	switch expr.Op {
	case hir.UnaryOpEnum.Not:
		op = cir.UnaryOpEnum.Not
	default:
		panic("unreachable")
	}
	return cir.NewUnaryExpr(op, c.buildExpr(expr.Expr))
}

func (c *CodeGenerator) buildBinary(expr *hir.Binary) cir.Expr {
	left, right := c.buildExpr(expr.Left), c.buildExpr(expr.Right)
	if strings.Contains(string(expr.Op), "=") {
		var op cir.AssignOp
		switch expr.Op {
		case hir.BinaryOpEnum.Assign:
			op = cir.AssignOpEnum.Assign
		case hir.BinaryOpEnum.AddAssign:
			op = cir.AssignOpEnum.AddAssign
		case hir.BinaryOpEnum.SubAssign:
			op = cir.AssignOpEnum.SubAssign
		case hir.BinaryOpEnum.MulAssign:
			op = cir.AssignOpEnum.MulAssign
		case hir.BinaryOpEnum.QuoAssign:
			op = cir.AssignOpEnum.QuoAssign
		case hir.BinaryOpEnum.RemAssign:
			op = cir.AssignOpEnum.RemAssign
		case hir.BinaryOpEnum.AndAssign:
			op = cir.AssignOpEnum.AndAssign
		case hir.BinaryOpEnum.OrAssign:
			op = cir.AssignOpEnum.OrAssign
		case hir.BinaryOpEnum.XorAssign:
			op = cir.AssignOpEnum.XorAssign
		default:
			panic("unreachable")
		}
		return cir.NewAssign(op, left, right)
	}

	var op cir.BinaryOp
	switch expr.Op {
	case hir.BinaryOpEnum.Add:
		op = cir.BinaryOpEnum.Add
	case hir.BinaryOpEnum.Sub:
		op = cir.BinaryOpEnum.Sub
	case hir.BinaryOpEnum.Mul:
		op = cir.BinaryOpEnum.Mul
	case hir.BinaryOpEnum.Quo:
		op = cir.BinaryOpEnum.Quo
	case hir.BinaryOpEnum.Rem:
		op = cir.BinaryOpEnum.Rem
	case hir.BinaryOpEnum.And:
		op = cir.BinaryOpEnum.And
	case hir.BinaryOpEnum.Or:
		op = cir.BinaryOpEnum.Or
	case hir.BinaryOpEnum.Xor:
		op = cir.BinaryOpEnum.Xor
	default:
		panic("unreachable")
	}
	return &cir.BinaryExpr{
		Op:    op,
		Left:  left,
		Right: right,
	}
}

func (c *CodeGenerator) buildNativeFunc(expr *hir.Func) *cir.FuncExpr {
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

func (c *CodeGenerator) buildNativeClosureFunc(expr *hir.Func, captureVars []hir.Ident) (*cir.Typedef, *cir.FuncExpr) {
	// 上下文
	fields := stlslices.Map(captureVars, func(i int, vexpr hir.Ident) *cir.Member {
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
		ctx.Value = optional.Some[cir.Expr](cir.NewUnaryExpr(cir.UnaryOpEnum.Mul, cir.NewCovert(cir.NewPointerType(cir.NewAliasType(ctxT)), cir.NewIdentExpr("_p0"))))
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

func (c *CodeGenerator) buildFunc(expr *hir.Func) *cir.MacroExpr {
	captureVars := stlslices.Filter(expr.UsedExternalVariables, func(i int, v hir.Ident) bool {
		return !stlval.Is[hir.Global](v)
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
			fields[fmt.Sprintf("_f%d", i+1)] = c.buildExpr(hir.NewIdentExpr(cv))
		}
		ctx := c.builder.BuildVarDecl(cir.NewAliasType(ctxT), "", cir.NewStruct(fields))
		return cir.NewMacroExpr("FUNCEXPR_C", &cir.IdentExpr{Name: f.Decl.Name}, cir.NewUnaryExpr(cir.UnaryOpEnum.AND, cir.NewIdentExpr(ctx.GetName())))
	}
}

func (c *CodeGenerator) buildCall(expr *hir.Call) cir.Expr {
	f := c.buildExpr(expr.Func)
	args := stlslices.Map(expr.Args, func(_ int, argExpr hir.Expr) cir.Expr {
		return c.buildExpr(argExpr)
	})
	if macroF, ok := f.(*cir.MacroExpr); ok && macroF.Name == "FUNCEXPR_F" {
		return cir.NewCall(macroF.Args[0], args...)
	} else if ok && macroF.Name == "FUNCEXPR_C" {
		return cir.NewCall(macroF.Args[0], append([]cir.Expr{macroF.Args[1]}, args...)...)
	}
	return cir.NewMacroExpr("FUNCCALL", append([]cir.Expr{f}, args...)...)
}

func (c *CodeGenerator) buildTuple(expr *hir.Tuple) *cir.Struct {
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

func (c *CodeGenerator) buildTupleIndex(expr *hir.TupleIndex) *cir.GetMember {
	from := c.buildExpr(expr.From)
	return cir.NewGetMember(from, fmt.Sprintf("_f%d", expr.Index.Int64()+1))
}

func (c *CodeGenerator) buildArray(expr *hir.Array) *cir.Struct {
	at := c.buildType(expr.GetType())
	if len(expr.Elems) == 0 {
		return cir.NewStruct(nil, at)
	}

	elems := stlslices.Map(expr.Elems, func(_ int, argExpr hir.Expr) cir.Expr {
		return c.buildExpr(argExpr)
	})
	return cir.NewStruct(map[string]cir.Expr{
		"array": cir.NewArray(elems),
	}, at)
}

func (c *CodeGenerator) buildArrayIndex(expr *hir.ArrayIndex) *cir.Offset {
	from := c.buildExpr(expr.From)
	offset := c.buildExpr(expr.Index)
	return cir.NewOffset(cir.NewGetMember(from, "array"), offset)
}

func (c *CodeGenerator) buildCovert(expr hir.Covert) cir.Expr {
	t := c.buildType(expr.GetType())
	v := c.buildExpr(expr.GetFrom())
	switch expr := expr.(type) {
	case *hir.Union:
		return cir.NewStruct(map[string]cir.Expr{
			"t": cir.NewInteger(big.NewInt(int64(expr.Index))),
			"v": cir.NewStruct(map[string]cir.Expr{
				fmt.Sprintf("t%d", expr.Index+1): v,
			}),
		}, t)
	case *hir.NumberCovert:
		return cir.NewCovert(t, v)
	default:
		panic("unreachable")
	}
}

func (c *CodeGenerator) buildGetRef(expr *hir.GetRef) cir.Expr {
	v := c.buildExpr(expr.Value)
	return cir.NewUnaryExpr(cir.UnaryOpEnum.AND, v)
}

func (c *CodeGenerator) buildDeRef(expr *hir.DeRef) cir.Expr {
	v := c.buildExpr(expr.Value)
	return cir.NewUnaryExpr(cir.UnaryOpEnum.Mul, v)
}
