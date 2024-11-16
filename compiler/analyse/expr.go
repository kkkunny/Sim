package analyse

import (
	"math/big"

	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlerror "github.com/kkkunny/stl/error"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	errors "github.com/kkkunny/Sim/compiler/error"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
	"github.com/kkkunny/Sim/compiler/token"
	"github.com/kkkunny/Sim/compiler/util"
)

func (self *Analyser) analyseExpr(expect hir.Type, node ast.Expr) hir.Value {
	switch exprNode := node.(type) {
	case *ast.Integer:
		return self.analyseInteger(expect, exprNode)
	case *ast.Char:
		return self.analyseChar(expect, exprNode)
	case *ast.Float:
		return self.analyseFloat(expect, exprNode)
	case *ast.Binary:
		return self.analyseBinary(expect, exprNode)
	case *ast.Unary:
		return self.analyseUnary(expect, exprNode)
	case *ast.IdentExpr:
		return self.analyseIdentExpr(exprNode)
	case *ast.Call:
		return self.analyseCall(expect, exprNode)
	case *ast.Tuple:
		return self.analyseTuple(expect, exprNode)
	case *ast.Covert:
		return self.analyseCovert(exprNode)
	case *ast.Array:
		return self.analyseArray(expect, exprNode)
	case *ast.Index:
		return self.analyseIndex(exprNode)
	case *ast.Extract:
		return self.analyseExtract(expect, exprNode)
	case *ast.Struct:
		return self.analyseStruct(exprNode)
	case *ast.Dot:
		return self.analyseDot(exprNode)
	case *ast.String:
		return self.analyseString(expect, exprNode)
	case *ast.Judgment:
		return self.analyseJudgment(exprNode)
	case *ast.Lambda:
		return self.analyseLambda(exprNode)
	default:
		panic("unreachable")
	}
	return nil
}

func (self *Analyser) expectExpr(expect hir.Type, node ast.Expr) hir.Value {
	value := self.analyseExpr(expect, node)
	newValue, ok := self.autoTypeCovert(expect, value)
	if !ok {
		errors.ThrowTypeMismatchError(node.Position(), value.Type(), expect)
	}
	return newValue
}

func (self *Analyser) analyseInteger(expect hir.Type, node *ast.Integer) hir.Value {
	if !types.Is[types.NumType](expect) {
		expect = types.Isize
	}
	if expectIt, ok := types.As[types.IntType](expect); ok {
		value, ok := big.NewInt(0).SetString(node.Value.Source(), 10)
		if !ok {
			panic("unreachable")
		}
		return values.NewInteger(expectIt, value)
	} else if expectFt, ok := types.As[types.FloatType](expect); ok {
		value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
		return values.NewFloat(expectFt, value)
	} else {
		panic("unreachable")
	}
}

func (self *Analyser) analyseChar(expect hir.Type, node *ast.Char) hir.Value {
	if !types.Is[types.NumType](expect) {
		expect = types.I32
	}
	s := node.Value.Source()
	char := util.ParseEscapeCharacter(s[1:len(s)-1], `\'`, `'`)[0]
	if expectIt, ok := types.As[types.IntType](expect); ok {
		value := big.NewInt(int64(char))
		return values.NewInteger(expectIt, value)
	} else if expectFt, ok := types.As[types.FloatType](expect); ok {
		value := big.NewFloat(float64(char))
		return values.NewFloat(expectFt, value)
	} else {
		panic("unreachable")
	}
}

func (self *Analyser) analyseFloat(expect hir.Type, node *ast.Float) *values.Float {
	if !types.Is[types.FloatType](expect) {
		expect = types.F64
	}
	expectFt, _ := types.As[types.FloatType](expect)
	value, _ := stlerror.MustWith2(big.NewFloat(0).Parse(node.Value.Source(), 10))
	return values.NewFloat(expectFt, value)
}

var assMap = map[token.Kind]token.Kind{
	token.ANDASS:  token.AND,
	token.ORASS:   token.OR,
	token.XORASS:  token.XOR,
	token.SHLASS:  token.SHL,
	token.SHRASS:  token.SHR,
	token.ADDASS:  token.ADD,
	token.SUBASS:  token.SUB,
	token.MULASS:  token.MUL,
	token.DIVASS:  token.DIV,
	token.REMASS:  token.REM,
	token.LANDASS: token.LAND,
	token.LORASS:  token.LOR,
}

func (self *Analyser) analyseBinary(expect hir.Type, node *ast.Binary) hir.Value {
	left := self.analyseExpr(expect, node.Left)
	lt := left.Type()
	right := self.analyseExpr(lt, node.Right)
	rt := right.Type()

	switch node.Opera.Kind {
	case token.ASS:
		if lt.Equal(rt) {
			if !left.Storable() {
				errors.ThrowExprTemporaryError(node.Left.Position())
			} else if !left.Mutable() {
				errors.ThrowNotMutableError(node.Left.Position())
			}
			return local.NewAssignExpr(left, right)
		}
	case token.ANDASS, token.ORASS, token.XORASS, token.SHLASS, token.SHRASS, token.ADDASS, token.SUBASS, token.MULASS, token.DIVASS, token.REMASS, token.LANDASS, token.LORASS:
		return self.analyseBinary(expect, &ast.Binary{
			Left: node.Left,
			Opera: token.Token{
				Position: node.Opera.Position,
				Kind:     token.ASS,
			},
			Right: &ast.Binary{
				Left: node.Left,
				Opera: token.Token{
					Position: node.Opera.Position,
					Kind:     assMap[node.Opera.Kind],
				},
				Right: node.Right,
			},
		})
	case token.AND:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("And")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.IntType](lt) {
			return local.NewAndExpr(left, right)
		}
	case token.OR:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Or")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.IntType](lt) {
			return local.NewOrExpr(left, right)
		}
	case token.XOR:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Xor")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.IntType](lt) {
			return local.NewXorExpr(left, right)
		}
	case token.SHL:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Shl")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.IntType](lt) {
			return local.NewShlExpr(left, right)
		}
	case token.SHR:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Shr")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.IntType](lt) {
			return local.NewShrExpr(left, right)
		}
	case token.ADD:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Add")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.NumType](lt) {
			return local.NewAddExpr(left, right)
		}
	case token.SUB:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Sub")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.NumType](lt) {
			return local.NewSubExpr(left, right)
		}
	case token.MUL:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Mul")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.NumType](lt) {
			return local.NewMulExpr(left, right)
		}
	case token.DIV:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Div")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.NumType](lt) {
			return local.NewDivExpr(left, right)
		}
	case token.REM:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Rem")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.NumType](lt) {
			return local.NewRemExpr(left, right)
		}
	case token.EQ:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Eq")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) {
			if types.Is[types.NoThingType](lt, true) || types.Is[types.NoReturnType](lt, true) {
				return values.NewBoolean(true)
			} else {
				return local.NewEqExpr(left, right)
			}
		}
	case token.NE:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Eq")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewNotExpr(local.NewCallExpr(local.NewMethodExpr(left, method), right))
		} else if lt.Equal(rt) {
			if types.Is[types.NoThingType](lt, true) || types.Is[types.NoReturnType](lt, true) {
				return values.NewBoolean(false)
			} else {
				return local.NewNeExpr(left, right)
			}
		}
	case token.LT:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Lt")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.NumType](lt) {
			return local.NewLtExpr(left, right)
		}
	case token.GT:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Gt")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.NumType](lt) {
			return local.NewGtExpr(left, right)
		}
	case token.LE:
		EQtrait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Eq")).(*global.Trait)
		LTtrait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Lt")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(EQtrait) && ct.HasImpl(LTtrait) {
			EQmethod := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(EQtrait.FirstMethod()).GetName())))
			LTmethod := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(LTtrait.FirstMethod()).GetName())))
			return local.NewLogicAndExpr(
				local.NewCallExpr(local.NewMethodExpr(left, EQmethod), right),
				local.NewCallExpr(local.NewMethodExpr(left, LTmethod), right),
			)
		} else if lt.Equal(rt) && types.Is[types.NumType](lt) {
			return local.NewLeExpr(left, right)
		}
	case token.GE:
		EQtrait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Eq")).(*global.Trait)
		GTtrait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Gt")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(EQtrait) && ct.HasImpl(GTtrait) {
			EQmethod := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(EQtrait.FirstMethod()).GetName())))
			GTmethod := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(GTtrait.FirstMethod()).GetName())))
			return local.NewLogicAndExpr(
				local.NewCallExpr(local.NewMethodExpr(left, EQmethod), right),
				local.NewCallExpr(local.NewMethodExpr(left, GTmethod), right),
			)
		} else if lt.Equal(rt) && types.Is[types.NumType](lt) {
			return local.NewGeExpr(left, right)
		}
	case token.LAND:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Land")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.BoolType](lt) {
			return local.NewLogicAndExpr(left, right)
		}
	case token.LOR:
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Lor")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](lt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(left, method), right)
		} else if lt.Equal(rt) && types.Is[types.BoolType](lt) {
			return local.NewLogicOrExpr(left, right)
		}
	default:
		panic("unreachable")
	}

	errors.ThrowIllegalBinaryError(node.Position(), node.Opera, left, right)
	return nil
}

func (self *Analyser) analyseUnary(expect hir.Type, node *ast.Unary) hir.Value {
	switch node.Opera.Kind {
	case token.SUB:
		value := self.analyseExpr(expect, node.Value)
		vt := value.Type()
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Neg")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](vt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(value, method))
		} else if types.Is[types.SintType](vt) || types.Is[types.FloatType](vt) {
			return local.NewOppositeExpr(value)
		}
		errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
		return nil
	case token.NOT:
		value := self.analyseExpr(expect, node.Value)
		vt := value.Type()
		trait := stlval.IgnoreWith(self.buildinPkg().GetIdent("Not")).(*global.Trait)
		ct, ok := types.As[global.CustomTypeDef](vt, true)
		if ok && ct.HasImpl(trait) {
			method := stlval.IgnoreWith(ct.GetMethod(stlval.IgnoreWith(stlval.IgnoreWith(trait.FirstMethod()).GetName())))
			return local.NewCallExpr(local.NewMethodExpr(value, method))
		}
		switch {
		case types.Is[types.IntType](vt):
			return local.NewNegExpr(value)
		case types.Is[types.BoolType](vt):
			return local.NewNotExpr(value)
		default:
			errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
			return nil
		}
	case token.AND, token.AND_WITH_MUT:
		if expect != nil && types.Is[types.RefType](expect) {
			if t, ok := types.As[types.RefType](expect); ok {
				expect = t.Pointer()
			}
		}
		value := self.analyseExpr(expect, node.Value)
		if !value.Storable() {
			errors.ThrowExprTemporaryError(node.Value.Position())
		} else if node.Opera.Is(token.AND_WITH_MUT) && !value.Mutable() {
			errors.ThrowNotMutableError(node.Value.Position())
		} else if localVarDef, ok := value.(*local.SingleVarDef); ok {
			localVarDef.SetEscaped(true)
		}
		return local.NewGetRefExpr(node.Opera.Is(token.AND_WITH_MUT), value)
	case token.MUL:
		if expect != nil {
			expect = types.NewRefType(false, expect)
		}
		value := self.analyseExpr(expect, node.Value)
		vt := value.Type()
		if !types.Is[types.RefType](vt) {
			errors.ThrowExpectReferenceError(node.Value.Position(), vt)
		}
		return local.NewDeRefExpr(value)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) tryAnalyseIdentExpr(node *ast.IdentExpr) (hir.Value, bool) {
	scope := self.scope
	if pkgToken, ok := node.Pkg.Value(); ok {
		scope, ok = self.pkg.GetExternPackage(pkgToken.Source())
		if !ok {
			errors.ThrowUnknownIdentifierError(pkgToken.Position, pkgToken)
		}
	}

	v, ok := scope.GetIdent(node.Name.Source(), true)
	if ok && stlval.Is[hir.Value](v) {
		return v.(hir.Value), true
	}
	return nil, false
}

func (self *Analyser) analyseIdentExpr(node *ast.IdentExpr) hir.Value {
	v, ok := self.tryAnalyseIdentExpr(node)
	if !ok {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	return v
}

func (self *Analyser) tryAnalyseEnum(node *ast.Call) (*local.EnumExpr, bool) {
	dotNode, ok := node.Func.(*ast.Dot)
	if !ok || len(node.Args) != 1 {
		return nil, false
	}
	identNode, ok := dotNode.From.(*ast.IdentExpr)
	if !ok {
		return nil, false
	}
	ident, ok := self.tryAnalyseIdent((*ast.Ident)(identNode))
	if !ok {
		return nil, false
	}
	t, ok := ident.Left()
	if !ok {
		return nil, false
	}
	et, ok := types.As[types.EnumType](t)
	if !ok {
		return nil, false
	}
	fieldName := dotNode.Index.Source()
	if !et.EnumFields().Contain(fieldName) {
		return nil, false
	}
	filed := et.EnumFields().Get(fieldName)
	fieldElem, ok := filed.Elem()
	if !ok {
		return nil, false
	}
	elem := self.analyseExpr(fieldElem, stlslices.First(node.Args))
	return local.NewEnumExpr(et, fieldName, elem), true
}

func (self *Analyser) analyseCall(expect hir.Type, node *ast.Call) hir.Value {
	// 枚举值
	if enum, ok := self.tryAnalyseEnum(node); ok {
		return enum
	}

	// 函数调用
	if expect != nil {
		expect = types.NewFuncType(expect)
	}
	f := self.analyseExpr(expect, node.Func)
	ct, ok := types.As[types.CallableType](f.Type())
	if !ok {
		errors.ThrowExpectCallableError(node.Func.Position(), f.Type())
	}
	pts := ct.Params()
	vararg := stlval.Is[*global.FuncDef](f) && stlslices.Exist(f.(*global.FuncDef).Attrs(), func(_ int, attr global.FuncAttr) bool {
		return stlval.Is[*global.FuncAttrVararg](attr)
	})
	if (!vararg && len(node.Args) != len(pts)) || (vararg && len(node.Args) < len(pts)) {
		errors.ThrowParameterNumberNotMatchError(node.Position(), uint(len(pts)), uint(len(node.Args)))
	}
	args := stlslices.Map(node.Args, func(i int, item ast.Expr) hir.Value {
		if vararg && i >= len(pts) {
			return self.analyseExpr(nil, item)
		} else {
			return self.expectExpr(pts[i], item)
		}
	})
	return local.NewCallExpr(f, args...)
}

func (self *Analyser) analyseTuple(expect hir.Type, node *ast.Tuple) hir.Value {
	// 括号表达式
	if len(node.Elems) == 1 && !types.Is[types.TupleType](expect) {
		return self.analyseExpr(expect, node.Elems[0])
	}

	// 元组
	expectElems := make([]hir.Type, len(node.Elems))
	if expectTt, ok := types.As[types.TupleType](expect); ok {
		elems := expectTt.Elems()
		if len(elems) < len(node.Elems) {
			copy(expectElems, elems)
		} else if len(elems) > len(node.Elems) {
			expectElems = elems[:len(node.Elems)]
		} else {
			expectElems = elems
		}
	}
	elems := stlslices.Map(node.Elems, func(i int, item ast.Expr) hir.Value {
		return self.analyseExpr(expectElems[i], item)
	})
	return local.NewTupleExpr(elems...)
}

// 自动类型转换
func (self *Analyser) autoTypeCovert(tt hir.Type, v hir.Value) (hir.Value, bool) {
	ft := v.Type()
	if ft.Equal(tt) {
		return v, true
	}

	if fromRt, fromOk := types.As[types.RefType](ft, true); fromOk && fromRt.Pointer().Equal(tt) {
		// &type -> type
		return local.NewDeRefExpr(v), true
	} else if toRt, toOk := types.As[types.RefType](tt, true); fromOk && toOk && fromRt.Mutable() && !toRt.Mutable() && fromRt.Pointer().Equal(toRt.Pointer()) {
		// &mut i8 -> &i8
		return local.NewWrapTypeExpr(v, toRt), true
	} else if types.Is[types.NoReturnType](ft, true) && (types.Is[types.NoThingType](tt, true) || self.hasTypeDefault(tt)) {
		// X -> default
		return local.NewNoReturn2AnyExpr(v, tt), true
	} else if fromFt, fromOk := types.As[types.FuncType](ft, true); fromOk && false {
		panic("unreachable")
	} else if toLt, toOk := types.As[types.LambdaType](tt, true); fromOk && toOk && fromFt.Equal(toLt.ToFunc()) {
		// func -> lambda
		return local.NewFunc2LambdaExpr(v, toLt), true
	} else {
		return v, false
	}
}

func (self *Analyser) analyseCovert(node *ast.Covert) hir.Value {
	tt := self.analyseType(node.Type)
	from := self.analyseExpr(tt, node.Value)
	ft := from.Type()
	if v, ok := self.autoTypeCovert(tt, from); ok {
		return v
	}

	if fromBt, fromOk := types.As[types.BuildInType](ft); fromOk && false {
		panic("unreachable")
	} else if toBt, toOk := types.As[types.BuildInType](tt); fromOk && toOk && fromBt.Equal(toBt) {
		// build-in -> build-in
		return local.NewWrapTypeExpr(from, tt)
	} else if toIt, toOk := types.As[types.IntType](tt); toOk && types.Is[types.IntType](ft) {
		// int -> int
		return local.NewInt2IntExpr(from, toIt)
	} else if toOk && types.Is[types.FloatType](ft) {
		// float -> int
		return local.NewFloat2IntExpr(from, toIt)
	} else if toFt, toOk := types.As[types.FloatType](tt); toOk && types.Is[types.IntType](ft) {
		// int -> float
		return local.NewInt2FloatExpr(from, toFt)
	} else if toOk && types.Is[types.FloatType](ft) {
		// float -> float
		return local.NewFloat2FloatExpr(from, toFt)
	} else if fromEt, fromOk := types.As[types.EnumType](ft); fromOk && false {
		panic("unreachable")
	} else if toUt, toOk := types.As[types.UintType](tt, true); fromOk && toOk && fromEt.Simple() && toUt.Equal(types.U8) {
		// simple enum -> u8
		return local.NewEnum2NumberExpr(from, toUt)
	} else if fromUt, fromOk := types.As[types.UintType](ft, true); fromOk && false {
		panic("unreachable")
	} else if toEt, toOk := types.As[types.EnumType](tt); fromOk && toOk && toEt.Simple() && fromUt.Equal(types.U8) {
		// u8 -> simple enum
		return local.NewNumber2EnumExpr(from, toEt)
	} else {
		errors.ThrowIllegalCovertError(node.Position(), ft, tt)
		return nil
	}
}

func (self *Analyser) analyseArray(expect hir.Type, node *ast.Array) *local.ArrayExpr {
	expectAt, _ := types.As[types.ArrayType](expect)
	if expectAt == nil && len(node.Elems) == 0 {
		errors.ThrowExpectArrayTypeError(node.Position(), types.NoThing)
	}

	elems := make([]hir.Value, len(node.Elems))
	for i, elemNode := range node.Elems {
		elems[i] = stlval.TernaryAction(i == 0, func() hir.Value {
			var expectElem hir.Type
			if expectAt != nil {
				expectElem = expectAt.Elem()
			}
			return self.analyseExpr(expectElem, elemNode)
		}, func() hir.Value {
			return self.expectExpr(stlslices.First(elems).Type(), elemNode)
		})
	}

	if expectAt == nil || expectAt.Size() != uint(len(elems)) {
		expectAt = types.NewArrayType(stlslices.First(elems).Type(), uint(len(elems)))
	}
	return local.NewArrayExpr(expectAt, elems...)
}

func (self *Analyser) analyseIndex(node *ast.Index) *local.IndexExpr {
	from := self.analyseExpr(nil, node.From)
	at, ok := types.As[types.ArrayType](from.Type())
	if !ok {
		errors.ThrowExpectArrayError(node.From.Position(), from.Type())
	}
	index := self.expectExpr(types.Usize, node.Index)
	if stlval.Is[*values.Integer](index) && index.(*values.Integer).Value().Cmp(big.NewInt(int64(at.Size()))) >= 0 {
		errors.ThrowIndexOutOfRange(node.Index.Position())
	}
	return local.NewIndexExpr(from, index)
}

func (self *Analyser) analyseExtract(expect hir.Type, node *ast.Extract) *local.ExtractExpr {
	indexValue, ok := big.NewInt(0).SetString(node.Index.Source(), 10)
	if !ok || !indexValue.IsUint64() {
		panic("unreachable")
	}
	// TODO: 溢出
	index := uint(indexValue.Uint64())

	if expect != nil {
		elems := make([]hir.Type, index+1)
		elems[index] = expect
		expect = types.NewTupleType(elems...)
	}
	from := self.analyseExpr(expect, node.From)
	ft, ok := types.As[types.TupleType](from.Type())
	if !ok {
		errors.ThrowExpectTupleError(node.From.Position(), from.Type())
	} else if index >= uint(len(ft.Elems())) {
		errors.ThrowInvalidIndexError(node.Index.Position, index)
	}
	return local.NewExtractExpr(from, index)
}

func (self *Analyser) analyseStruct(node *ast.Struct) *local.StructExpr {
	stObj := self.analyseType(node.Type)
	st, ok := types.As[types.StructType](stObj)
	if !ok {
		errors.ThrowExpectStructTypeError(node.Type.Position(), stObj)
	}
	td, ok := types.As[types.TypeDef](stObj, true)
	if !ok {
		panic("unreachable")
	}

	existedFields := make(map[string]hir.Value)
	for _, fieldNode := range node.Fields {
		fn := fieldNode.E1().Source()
		if !st.Fields().Contain(fn) || (!self.pkg.Equal(td.Package()) && !st.Fields().Get(fn).Public()) {
			errors.ThrowUnknownIdentifierError(fieldNode.E1().Position, fieldNode.E1())
		}
		existedFields[fn] = self.expectExpr(st.Fields().Get(fn).Type(), fieldNode.E2())
	}

	fields := make([]hir.Value, st.Fields().Length())
	var i int
	for iter := st.Fields().Iterator(); iter.Next(); i++ {
		fn, ft := iter.Value().E1(), iter.Value().E2().Type()
		if fv, ok := existedFields[fn]; ok {
			fields[i] = fv
		} else {
			fields[i] = self.getTypeDefaultValue(node.Type.Position(), ft)
		}
	}

	return local.NewStructExpr(st, fields...)
}

func (self *Analyser) analyseDot(node *ast.Dot) hir.Value {
	fieldName := node.Index.Source()

	if identNode, ok := node.From.(*ast.IdentExpr); ok {
		identType, ok := self.tryAnalyseIdentType((*ast.IdentType)(identNode))
		if ok {
			if ctd, ok := types.As[global.CustomTypeDef](identType, true); ok {
				// 静态方法
				method, ok := ctd.GetMethod(fieldName)
				if ok && (method.Public() || self.pkg.Equal(method.Package())) {
					return method
				}
			}
			if et, ok := types.As[types.EnumType](identType); ok {
				// 枚举值
				if et.EnumFields().Contain(fieldName) {
					_, ok = et.EnumFields().Get(fieldName).Elem()
					if !ok {
						return local.NewEnumExpr(et, fieldName)
					}
				}
			}
		}
	}

	if method, ok := self.analyseMethod(node); ok {
		return method
	}

	from := self.analyseExpr(nil, node.From)
	ft := from.Type()

	// 字段
	var fromStVal hir.Value
	if types.Is[types.StructType](ft) {
		fromStVal = from
	} else if fromRt, fromOk := types.As[types.RefType](ft, true); fromOk && false {
		panic("unreachable")
	} else if fromOk && types.Is[types.StructType](fromRt.Pointer()) {
		fromStVal = local.NewDeRefExpr(from)
	} else {
		errors.ThrowExpectStructError(node.From.Position(), ft)
	}
	fromSt, ok := types.As[types.StructType](fromStVal.Type())
	if !ok {
		panic("unreachable")
	}
	fromTd, ok := types.As[types.TypeDef](fromStVal.Type(), true)
	if !ok {
		panic("unreachable")
	}

	if field := fromSt.Fields().Get(fieldName); !fromSt.Fields().Contain(fieldName) || (!field.Public() && !self.pkg.Equal(fromTd.Package())) {
		errors.ThrowUnknownIdentifierError(node.Index.Position, node.Index)
	}
	return local.NewFieldExpr(fromStVal, fieldName)
}

func (self *Analyser) analyseMethod(node *ast.Dot) (values.Callable, bool) {
	fieldName := node.Index.Source()
	from := self.analyseExpr(nil, node.From)
	ft := from.Type()

	var fromCtd global.CustomTypeDef
	if fromCtd2, fromOk := types.As[global.CustomTypeDef](ft, true); fromOk {
		fromCtd = fromCtd2
	} else if fromRt, fromOk := types.As[types.RefType](ft, true); fromOk && false {
		panic("unreachable")
	} else if fromOk && types.Is[global.CustomTypeDef](fromRt.Pointer(), true) {
		fromCtd, _ = types.As[global.CustomTypeDef](fromRt.Pointer(), true)
	} else {
		return nil, false
	}

	method, ok := fromCtd.GetMethod(fieldName)
	if !ok || (!method.Public() && !self.pkg.Equal(method.Package())) {
		return nil, false
	}

	if method.Static() {
		return method, true
	}

	selfParam, ok := method.SelfParam()
	if !ok {
		panic("unreachable")
	}
	var selfVal hir.Value
	if fromIsRef := types.Is[types.RefType](ft, true); method.SelfParamIsRef() && !fromIsRef {
		if !from.Storable() {
			errors.ThrowExprTemporaryError(node.From.Position())
		}
		selfVal = local.NewGetRefExpr(selfParam.Mutable(), from)
	} else if !method.SelfParamIsRef() && fromIsRef {
		selfVal = local.NewDeRefExpr(from)
	} else {
		selfVal = from
	}

	return local.NewMethodExpr(selfVal, method), true
}

func (self *Analyser) analyseString(expect hir.Type, node *ast.String) *values.String {
	if !types.Is[types.StrType](expect) {
		expect = types.Str
	}
	expectSt, ok := types.As[types.StrType](expect)
	if !ok {
		panic("unreachable")
	}

	s := node.Value.Source()
	s = util.ParseEscapeCharacter(s[1:len(s)-1], `\"`, `"`)
	return values.NewString(expectSt, s)
}

func (self *Analyser) analyseJudgment(node *ast.Judgment) hir.Value {
	tt := self.analyseType(node.Type)
	from := self.analyseExpr(tt, node.Value)
	ft := from.Type()

	if ft.Equal(tt) {
		return values.NewBoolean(true)
	} else {
		return local.NewTypeJudgmentExpr(from, tt)
	}
}

func (self *Analyser) analyseLambda(node *ast.Lambda) hir.Value {
	params := stlslices.Map(node.Params, func(_ int, e ast.Param) *local.Param { return self.analyseParam(e) })
	ret := self.analyseType(node.Ret, self.voidTypeAnalyser(), self.noReturnTypeAnalyser())
	lt := types.NewLambdaType(ret, stlslices.Map(params, func(_ int, param *local.Param) hir.Type { return param.Type() })...)

	captureIdents := set.StdHashSetWith[values.Ident]()
	onCapture := func(ident any) {
		exprIdent, ok := ident.(values.Ident)
		if !ok {
			return
		}
		if varDecl, ok := exprIdent.(values.VarDecl); ok {
			varDecl.SetEscaped(true)
		}
		captureIdents.Add(exprIdent)
	}

	l := local.NewLambdaExpr(self.scope, lt, params, onCapture)
	l.SetBody(self.analyseFuncBody(l, node.Params, node.Body))
	l.SetContext(captureIdents.ToSlice()...)
	return l
}
