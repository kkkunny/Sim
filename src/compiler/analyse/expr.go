package analyse

import (
	. "github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/parse"
	"github.com/kkkunny/Sim/src/compiler/utils"
)

// 表达式
func analyseExpr(ctx *blockContext, expect Type, ast parse.Expr) (Expr, utils.Error) {
	switch expr := ast.(type) {
	case *parse.Int:
		if expect == nil || !IsNumberTypeAndSon(expect) {
			expect = Isize
		}
		if IsIntTypeAndSon(expect) {
			return &Integer{
				Type:  expect,
				Value: expr.Value,
			}, nil
		} else {
			return &Float{
				Type:  expect,
				Value: float64(expr.Value),
			}, nil
		}
	case *parse.Float:
		if expect == nil || !IsFloatTypeAndSon(expect) {
			expect = F64
		}
		return &Float{
			Type:  expect,
			Value: expr.Value,
		}, nil
	case *parse.Bool:
		if expect == nil || !IsBoolTypeAndSon(expect) {
			expect = Bool
		}
		return &Boolean{
			Type:  expect,
			Value: expr.Value,
		}, nil
	case *parse.Char:
		if expect == nil || !IsNumberTypeAndSon(expect) {
			expect = I32
		}
		return &Integer{
			Type:  expect,
			Value: int64(expr.Value),
		}, nil
	case *parse.String:
		if expect == nil || !GetDepthBaseType(expect).Equal(NewPtrType(I8)) {
			expect = NewPtrType(I8)
		}
		return &String{
			Type:  expect,
			Value: expr.Value,
		}, nil
	case *parse.Null:
		if expect == nil || (!IsPtrTypeAndSon(expect) && !IsFuncTypeAndSon(expect)) {
			return nil, utils.Errorf(expr.Position(), "expect a pointer type")
		}
		return &Null{Type: expect}, nil
	case *parse.Ident:
		return analyseIdent(ctx, expr)
	case *parse.Array:
		if len(expr.Elems) == 0 {
			if expect == nil || !IsArrayTypeAndSon(expect) {
				return nil, utils.Errorf(expr.Position(), "expect a array type")
			}
			return &EmptyArray{Type: expect}, nil
		}
		if expect != nil {
			if at, ok := GetBaseType(expect).(*TypeArray); ok && at.Size == uint(len(expr.Elems)) {
				expect = at.Elem
			}
		}

		elems := make([]Expr, len(expr.Elems))
		var errs []utils.Error
		for i, e := range expr.Elems {
			var err utils.Error
			if elems[0] == nil {
				elems[i], err = analyseExpr(ctx, expect, e)
			} else {
				elems[i], err = expectExpr(ctx, elems[0].GetType(), e)
			}
			if err != nil {
				errs = append(errs, err)
			}
		}
		if len(errs) == 1 {
			return nil, errs[0]
		} else if len(errs) > 1 {
			return nil, utils.NewMultiError(errs...)
		}

		var rt Type = NewArrayType(uint(len(elems)), elems[0].GetType())
		if expect != nil && GetDepthBaseType(expect).Equal(GetDepthBaseType(rt)) {
			rt = expect
		}
		return &Array{
			Type:  rt,
			Elems: elems,
		}, nil
	case *parse.TupleOrExpr:
		if len(expr.Elems) == 0 {
			if expect == nil || !IsTupleTypeAndSon(expect) {
				return nil, utils.Errorf(expr.Position(), "expect a tuple type")
			}
			return &EmptyTuple{Type: expect}, nil
		} else if len(expr.Elems) == 1 && (expect == nil || !IsTupleTypeAndSon(expect) || len(GetBaseType(expect).(*TypeTuple).Elems) != 1) {
			return analyseExpr(ctx, expect, expr.Elems[0])
		}
		expects := make([]Type, len(expr.Elems))
		if expect != nil {
			if tt, ok := GetBaseType(expect).(*TypeTuple); ok && len(tt.Elems) == len(expr.Elems) {
				for i := range expects {
					expects[i] = tt.Elems[i]
				}
			}
		}
		elems, err := analyseExprList(ctx, expects, expr.Elems)
		if err != nil {
			return nil, err
		}
		for i, e := range elems {
			expects[i] = e.GetType()
		}
		var rt Type = NewTupleType(expects...)
		if expect != nil && GetDepthBaseType(expect).Equal(GetDepthBaseType(rt)) {
			rt = expect
		}
		return &Tuple{
			Type:  rt,
			Elems: elems,
		}, nil
	case *parse.Struct:
		if len(expr.Fields) == 0 {
			if expect == nil || !IsStructTypeAndSon(expect) {
				return nil, utils.Errorf(expr.Position(), "expect a struct type")
			}
			return &EmptyStruct{Type: expect}, nil
		}
		if expect == nil || !IsStructTypeAndSon(expect) {
			return nil, utils.Errorf(expr.Position(), "expect a struct type")
		} else if GetBaseType(expect).(*TypeStruct).Fields.Length() != len(expr.Fields) {
			return nil, utils.Errorf(expr.Position(), "expect `%d` fields", len(expr.Fields))
		}
		expects := make([]Type, len(expr.Fields))
		for iter := GetBaseType(expect).(*TypeStruct).Fields.Begin(); iter.HasValue(); iter.Next() {
			expects[iter.Index()] = iter.Value().Second
		}
		fields, err := analyseExprList(ctx, expects, expr.Fields)
		if err != nil {
			return nil, err
		}
		for i, e := range fields {
			expects[i] = e.GetType()
		}
		return &Struct{
			Type:   expect,
			Fields: fields,
		}, nil
	case *parse.Unary:
		switch expr.Opera.Kind {
		case lex.SUB:
			value, err := analyseExpr(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			if !IsNumberTypeAndSon(value.GetType()) {
				return nil, utils.Errorf(expr.Value.Position(), "expect a number")
			}
			left, err := getDefaultExprByType(expr.Value.Position(), value.GetType())
			if err != nil {
				return nil, err
			}
			return &Binary{
				Opera: "-",
				Left:  left,
				Right: value,
			}, nil
		case lex.NEG:
			value, err := analyseExpr(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			if !IsSintTypeAndSon(value.GetType()) {
				return nil, utils.Errorf(expr.Value.Position(), "expect a signed integer")
			}
			return &Binary{
				Opera: "^",
				Left:  value,
				Right: &Integer{
					Type:  value.GetType(),
					Value: -1,
				},
			}, nil
		case lex.NOT:
			if expect == nil || !GetBaseType(expect).Equal(Bool) {
				expect = Bool
			}
			value, err := expectExprAndSon(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			return &Unary{
				Mutable: false,
				Type:    value.GetType(),
				Opera:   "!",
				Value:   value,
			}, nil
		case lex.AND:
			if expect != nil && IsPtrTypeAndSon(expect) {
				expect = GetBaseType(expect).(*TypePtr).Elem
			}
			value, err := analyseExpr(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			if value.IsTemporary() {
				return nil, utils.Errorf(expr.Value.Position(), "not expect a temporary value")
			}
			return &Unary{
				Mutable: false,
				Type:    NewPtrType(value.GetType()),
				Opera:   "&",
				Value:   value,
			}, nil
		case lex.MUL:
			if expect != nil {
				expect = NewPtrType(expect)
			}
			value, err := analyseExpr(ctx, expect, expr.Value)
			if err != nil {
				return nil, err
			}
			vt := value.GetType()
			if !IsPtrTypeAndSon(vt) {
				return nil, utils.Errorf(expr.Value.Position(), "expect a pointer")
			}
			return &Unary{
				Mutable: value.GetMut(),
				Type:    GetBaseType(vt).(*TypePtr).Elem,
				Opera:   "*",
				Value:   value,
			}, nil
		default:
			panic("")
		}
	case *parse.Binary:
		left, err := analyseExpr(ctx, nil, expr.Left)
		if err != nil {
			return nil, err
		}
		lt := left.GetType()
		right, err := expectExpr(ctx, lt, expr.Right)
		if err != nil {
			return nil, err
		}
		switch expr.Opera.Kind {
		case lex.ASS, lex.ADS, lex.SUS, lex.MUS, lex.DIS, lex.MOS, lex.ANS, lex.ORS, lex.XOS, lex.SLS, lex.SRS:
			if !left.GetMut() {
				return nil, utils.Errorf(expr.Left.Position(), "expect a mutable value")
			}
			switch expr.Opera.Kind {
			case lex.ASS:
			case lex.ADS, lex.SUS, lex.MUS, lex.DIS, lex.MOS:
				if !IsNumberTypeAndSon(lt) {
					return nil, utils.Errorf(expr.Left.Position(), "expect a number")
				}
			case lex.ANS, lex.ORS, lex.XOS, lex.SLS, lex.SRS:
				if !IsIntTypeAndSon(lt) {
					return nil, utils.Errorf(expr.Left.Position(), "expect a integer")
				}
			default:
				panic("unknown binary")
			}
			return &Assign{
				Opera: expr.Opera.Source,
				Left:  left,
				Right: right,
			}, nil
		case lex.LAN, lex.LOR:
			if !IsBoolTypeAndSon(lt) {
				return nil, utils.Errorf(expr.Left.Position(), "expect a boolean")
			}
		case lex.EQ, lex.NE:
			return &Equal{
				Opera: expr.Opera.Source,
				Left:  left,
				Right: right,
			}, nil
		case lex.LT, lex.LE, lex.GT, lex.GE:
			if !IsNumberTypeAndSon(lt) {
				return nil, utils.Errorf(expr.Left.Position(), "expect a number")
			}
			return &Equal{
				Opera: expr.Opera.Source,
				Left:  left,
				Right: right,
			}, nil
		case lex.ADD, lex.SUB, lex.MUL, lex.DIV, lex.MOD:
			if !IsNumberTypeAndSon(lt) {
				return nil, utils.Errorf(expr.Left.Position(), "expect a number")
			}
		case lex.AND, lex.OR, lex.XOR, lex.SHL, lex.SHR:
			if !IsIntTypeAndSon(lt) {
				return nil, utils.Errorf(expr.Left.Position(), "expect a integer")
			}
		default:
			panic("unknown binary")
		}
		return &Binary{
			Opera: expr.Opera.Source,
			Left:  left,
			Right: right,
		}, nil
	case *parse.Ternary:
		cond, err := expectExprAndSon(ctx, Bool, expr.Cond)
		if err != nil {
			return nil, err
		}
		tv, err := analyseExpr(ctx, expect, expr.True)
		if err != nil {
			return nil, err
		}
		fv, err := expectExpr(ctx, tv.GetType(), expr.False)
		if err != nil {
			return nil, err
		}
		return &Select{
			Cond:  cond,
			True:  tv,
			False: fv,
		}, nil
	case *parse.Call:
		f, err := analyseExpr(ctx, nil, expr.Func)
		if err != nil {
			if ident, ok := expr.Func.(*parse.Ident); ok && ident.Pkg.Path == ctx.GetPackagePath() {
				return analyseBuildInFuncCall(ctx, ident, expr.Args)
			}
			return nil, err
		}

		if method, ok := f.(*Method); ok {
			// 方法调用
			ft := method.GetMethodType()
			if (!ft.VarArg && len(ft.Params) != len(expr.Args)) ||
				(ft.VarArg && len(ft.Params) > len(expr.Args)) {
				return nil, utils.Errorf(expr.Func.Position(), "expect %d arguments", len(ft.Params))
			}

			args := make([]Expr, len(expr.Args))
			var errs []utils.Error
			for i, a := range expr.Args {
				var arg Expr
				var err utils.Error
				if i < len(ft.Params) {
					arg, err = expectExpr(ctx, ft.Params[i], a)
				} else {
					arg, err = analyseExpr(ctx, nil, a)
				}
				if err != nil {
					errs = append(errs, err)
				} else {
					args[i] = arg
				}
			}
			if len(errs) == 1 {
				return nil, errs[0]
			} else if len(errs) > 1 {
				return nil, utils.NewMultiError(errs...)
			}

			return &MethodCall{
				Method: method,
				Args:   args,
			}, nil
		} else if im, ok := f.(*GetInterfaceField); ok {
			// 接口成员方法调用
			ft := im.GetMethodType()
			if (!ft.VarArg && len(ft.Params) != len(expr.Args)) ||
				(ft.VarArg && len(ft.Params) > len(expr.Args)) {
				return nil, utils.Errorf(expr.Func.Position(), "expect %d arguments", len(ft.Params))
			}

			args := make([]Expr, len(expr.Args))
			var errs []utils.Error
			for i, a := range expr.Args {
				var arg Expr
				var err utils.Error
				if i < len(ft.Params) {
					arg, err = expectExpr(ctx, ft.Params[i], a)
				} else {
					arg, err = analyseExpr(ctx, nil, a)
				}
				if err != nil {
					errs = append(errs, err)
				} else {
					args[i] = arg
				}
			}
			if len(errs) == 1 {
				return nil, errs[0]
			} else if len(errs) > 1 {
				return nil, utils.NewMultiError(errs...)
			}

			return &InterfaceFieldCall{
				Field: im,
				Args:  args,
			}, nil
		} else {
			// 函数调用
			ft, ok := GetBaseType(f.GetType()).(*TypeFunc)
			if !ok {
				return nil, utils.Errorf(expr.Func.Position(), "expect a function")
			} else if (!ft.VarArg && len(ft.Params) != len(expr.Args)) ||
				(ft.VarArg && len(ft.Params) > len(expr.Args)) {
				return nil, utils.Errorf(expr.Func.Position(), "expect %d arguments", len(ft.Params))
			}

			args := make([]Expr, len(expr.Args))
			var errs []utils.Error
			for i, a := range expr.Args {
				var arg Expr
				var err utils.Error
				if i < len(ft.Params) {
					arg, err = expectExpr(ctx, ft.Params[i], a)
				} else {
					arg, err = analyseExpr(ctx, nil, a)
				}
				if err != nil {
					errs = append(errs, err)
				} else {
					args[i] = arg
				}
			}
			if len(errs) == 1 {
				return nil, errs[0]
			} else if len(errs) > 1 {
				return nil, utils.NewMultiError(errs...)
			}

			return &FuncCall{
				Func: f,
				Args: args,
			}, nil
		}
	case *parse.Dot:
		prefix, err := analyseExpr(ctx, nil, expr.Front)
		if err != nil {
			return nil, err
		}

		// 方法
		prefixType := prefix.GetType()
		if IsTypedef(prefixType) || (IsPtrType(prefixType) && IsTypedef(prefixType.(*TypePtr).Elem)) {
			var _selfType *Typedef
			if td, ok := prefixType.(*Typedef); ok {
				_selfType = td
			} else {
				_selfType = prefixType.(*TypePtr).Elem.(*Typedef)
			}

			if fun, ok := _selfType.Methods[expr.End.Source]; ok {
				return &Method{
					Self: prefix,
					Func: fun,
				}, nil
			}
		}

		// 属性
		switch t := GetBaseType(prefixType).(type) {
		case *TypeStruct:
			if !t.Fields.ContainKey(expr.End.Source) {
				return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
			} else if td, ok := prefixType.(*Typedef); ok && ctx.GetPackageContext().path != td.Pkg && !t.Fields.Get(expr.End.Source).First {
				return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
			}
			return &GetField{
				From:  prefix,
				Index: expr.End.Source,
			}, nil
		case *TypeInterface:
			if !t.Fields.ContainKey(expr.End.Source) {
				return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
			}
			return &GetInterfaceField{
				From:  prefix,
				Index: expr.End.Source,
			}, nil
		case *TypePtr:
			if st, ok := GetBaseType(t.Elem).(*TypeStruct); ok {
				if !st.Fields.ContainKey(expr.End.Source) {
					return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
				} else if td, ok := t.Elem.(*Typedef); ok && ctx.GetPackageContext().path != td.Pkg && !st.Fields.Get(expr.End.Source).First {
					return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
				}
				return &GetField{
					From: &Unary{
						Mutable: prefix.GetMut(),
						Type:    t.Elem,
						Opera:   "*",
						Value:   prefix,
					},
					Index: expr.End.Source,
				}, nil
			} else if st, ok := GetBaseType(t.Elem).(*TypeInterface); ok {
				if !st.Fields.ContainKey(expr.End.Source) {
					return nil, utils.Errorf(expr.End.Pos, "unknown identifier")
				}
				return &GetInterfaceField{
					From: &Unary{
						Mutable: prefix.GetMut(),
						Type:    t.Elem,
						Opera:   "*",
						Value:   prefix,
					},
					Index: expr.End.Source,
				}, nil
			}
		}
		return nil, utils.Errorf(expr.Front.Position(), "expect a struct")
	case *parse.Index:
		prefix, err := analyseExpr(ctx, nil, expr.Front)
		if err != nil {
			return nil, err
		}
		switch pt := GetBaseType(prefix.GetType()).(type) {
		case *TypeArray:
			index, err := autoExpectExpr(ctx, Usize, expr.Index)
			if err != nil {
				return nil, err
			}
			return &Index{
				Type:  pt.Elem,
				From:  prefix,
				Index: index,
			}, nil
		case *TypePtr:
			index, err := autoExpectExpr(ctx, Usize, expr.Index)
			if err != nil {
				return nil, err
			}
			return &Index{
				Type:  pt.Elem,
				From:  prefix,
				Index: index,
			}, nil
		case *TypeTuple:
			index, err := analyseExpr(ctx, Usize, expr.Index)
			if err != nil {
				return nil, err
			}
			literal, ok := index.(*Integer)
			if !ok {
				return nil, utils.Errorf(expr.Index.Position(), "expect a integer literal")
			}
			return &Index{
				Type:  pt.Elems[literal.Value],
				From:  prefix,
				Index: literal,
			}, nil
		default:
			return nil, utils.Errorf(expr.Front.Position(), "expect a array or tuple")
		}
	case *parse.Covert:
		to, err := analyseType(ctx.GetPackageContext(), expr.To)
		if err != nil {
			return nil, err
		}
		from, err := analyseExpr(ctx, to, expr.From)
		if err != nil {
			return nil, err
		}
		res := analyseCovert(from, to)
		if res == nil {
			return nil, utils.Errorf(expr.From.Position(), "can not covert to type `%s`", to)
		}
		return res, nil
	default:
		panic("unknown expression")
	}
}

// 期待指定类型的表达式
func expectExprWithType(pos utils.Position, expect Type, expr Expr) (Expr, utils.Error) {
	if vv := analyseAutoCovert(expr, expect); vv != nil {
		return vv, nil
	}
	exprType := expr.GetType()
	if !exprType.Equal(expect) {
		return nil, utils.Errorf(pos, "expect type `%s` but there is `%s`", expect, exprType)
	}
	return expr, nil
}

// 期待指定类型的表达式及其子类型
func expectExprWithTypeAndSon(pos utils.Position, expect Type, expr Expr) (Expr, utils.Error) {
	if vv := analyseAutoCovert(expr, expect); vv != nil {
		return vv, nil
	}
	exprType := expr.GetType()
	if !GetDepthBaseType(exprType).Equal(GetDepthBaseType(expect)) {
		return nil, utils.Errorf(pos, "expect type `%s` but there is `%s`", expect, exprType)
	}
	return expr, nil
}

// 期待指定类型的表达式
func expectExpr(ctx *blockContext, expect Type, ast parse.Expr) (Expr, utils.Error) {
	expr, err := analyseExpr(ctx, expect, ast)
	if err != nil {
		return nil, err
	}
	return expectExprWithType(ast.Position(), expect, expr)
}

// 期待指定类型的表达式及其子类型
func expectExprAndSon(ctx *blockContext, expect Type, ast parse.Expr) (Expr, utils.Error) {
	expr, err := analyseExpr(ctx, expect, ast)
	if err != nil {
		return nil, err
	}
	return expectExprWithTypeAndSon(ast.Position(), expect, expr)
}

// 自动转换成期待的类型
func autoExpectExpr(ctx *blockContext, expect Type, ast parse.Expr) (Expr, utils.Error) {
	expr, err := analyseExpr(ctx, expect, ast)
	if err != nil {
		return nil, err
	}
	v := analyseCovert(expr, expect)
	if v == nil {
		return nil, utils.Errorf(ast.Position(), "expect type `%s` but there is `%s`", expect, expr.GetType())
	}
	return v, nil
}

// 获取类型默认值
func getDefaultExprByType(pos utils.Position, t Type) (Expr, utils.Error) {
	switch GetBaseType(t).(type) {
	case *TypeBasic:
		switch {
		case IsNoneType(t):
			panic("")
		case IsIntType(t):
			return &Integer{
				Type:  t,
				Value: 0,
			}, nil
		case IsFloatType(t):
			return &Float{
				Type:  t,
				Value: 0,
			}, nil
		case IsBoolType(t):
			return &Boolean{
				Type:  t,
				Value: false,
			}, nil
		default:
			panic("")
		}
	case *TypeFunc:
		return &Null{Type: t}, nil
	case *TypeArray:
		return &EmptyArray{Type: t}, nil
	case *TypeTuple:
		return &EmptyTuple{Type: t}, nil
	case *TypeStruct:
		return &EmptyStruct{Type: t}, nil
	case *TypePtr:
		return &Null{Type: t}, nil
	case *TypeInterface:
		return nil, utils.Errorf(pos, "can not get default value for this type")
	default:
		panic("")
	}
}

// 表达式列表
func analyseExprList(ctx *blockContext, expects []Type, asts []parse.Expr) ([]Expr, utils.Error) {
	exprs := make([]Expr, len(asts))
	var errors []utils.Error
	for i, e := range asts {
		var expect Type
		if len(expects) == len(asts) {
			expect = expects[i]
		}
		expr, err := analyseExpr(ctx, expect, e)
		if err != nil {
			errors = append(errors, err)
		} else {
			exprs[i] = expr
		}
	}

	if len(errors) == 0 {
		return exprs, nil
	} else if len(errors) == 1 {
		return nil, errors[0]
	} else {
		return nil, utils.NewMultiError(errors...)
	}
}

// 标识符
func analyseIdent(ctx *blockContext, ast *parse.Ident) (Expr, utils.Error) {
	if ast.Pkg.Path == ctx.GetPackagePath() {
		value := ctx.GetValue(ast.Name.Source)
		if value != nil {
			return value, nil
		}
	} else {
		pkg := ctx.GetPackageContext().f.Pkgs[ast.Pkg.Path]
		value := pkg.GetValue(ast.Name.Source)
		if value.Second != nil && value.First {
			return value.Second, nil
		}
	}
	return nil, utils.Errorf(ast.Name.Pos, "unknown identifier")
}

// 内置函数调用
func analyseBuildInFuncCall(ctx *blockContext, ident *parse.Ident, paramAsts []parse.Expr) (Expr, utils.Error) {
	switch ident.Name.Source {
	case "len":
		if len(paramAsts) != 1 {
			return nil, utils.Errorf(ident.Position(), "expect 1 arguments")
		}
		param, err := analyseExpr(ctx, nil, paramAsts[0])
		if err != nil {
			return nil, err
		}
		pt := param.GetType()
		array, ok := GetBaseType(pt).(*TypeArray)
		if ok {
			return &Integer{
				Type:  Usize,
				Value: int64(array.Size),
			}, nil
		}
		return nil, utils.Errorf(paramAsts[0].Position(), "expect a array")
	case "typename":
		if len(paramAsts) != 1 {
			return nil, utils.Errorf(ident.Position(), "expect 1 arguments")
		}
		param, err := analyseExpr(ctx, nil, paramAsts[0])
		if err != nil {
			return nil, err
		}
		return &String{
			Type:  NewPtrType(I8),
			Value: param.GetType().String(),
		}, nil
	case "alloc":
		if len(paramAsts) != 1 {
			return nil, utils.Errorf(ident.Position(), "expect 1 arguments")
		}
		param, err := expectExpr(ctx, Usize, paramAsts[0])
		if err != nil {
			return nil, err
		}
		return &Alloc{Size: param}, nil
	default:
		return nil, utils.Errorf(ident.Position(), "unknown identifier")
	}
}

// 类型转换
func analyseCovert(v Expr, t Type) *Covert {
	if vv := analyseAutoCovert(v, t); vv != nil {
		return vv
	}

	ft := v.GetType()
	switch {
	case GetDepthBaseType(ft).Equal(GetDepthBaseType(t)):
	case IsNumberTypeAndSon(ft) && IsNumberTypeAndSon(t):
	case GetBaseType(ft).Equal(Usize) && (IsPtrTypeAndSon(t) || IsFuncTypeAndSon(t)):
	case (IsPtrTypeAndSon(ft) || IsFuncTypeAndSon(ft)) && GetBaseType(t).Equal(Usize):
	case (IsPtrTypeAndSon(ft) || IsFuncTypeAndSon(ft)) && (IsPtrTypeAndSon(t) || IsFuncTypeAndSon(t)):
	default:
		return nil
	}

	return &Covert{
		From: v,
		To:   t,
	}
}

// 自动类型转换
func analyseAutoCovert(v Expr, t Type) *Covert {
	ft := v.GetType()
	switch {
	case IsPtrTypeAndSon(ft) && IsTypedef(GetBaseType(ft).(*TypePtr).Elem) && IsInterfaceTypeAndSon(t) && GetBaseType(ft).(*TypePtr).Elem.(*Typedef).IsImpl(GetBaseType(t).(*TypeInterface)):
		// 类型定义指针 --> 接口类型定义
	default:
		return nil
	}

	return &Covert{
		From: v,
		To:   t,
	}
}
