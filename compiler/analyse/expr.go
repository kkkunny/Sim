package analyse

import (
	"math/big"
	"slices"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlerror "github.com/kkkunny/stl/error"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/compiler/ast"

	"github.com/kkkunny/Sim/compiler/hir"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
	"github.com/kkkunny/Sim/compiler/util"
)

func (self *Analyser) analyseExpr(expect hir.Type, node ast.Expr) hir.Expr {
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
		return self.analyseCall(exprNode)
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
		return self.analyseString(exprNode)
	case *ast.Judgment:
		return self.analyseJudgment(exprNode)
	case *ast.Lambda:
		return self.analyseLambda(expect, exprNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseInteger(expect hir.Type, node *ast.Integer) hir.ExprStmt {
	if expect == nil || !hir.IsNumberType(expect) {
		expect = self.pkgScope.Isize()
	}
	switch {
	case hir.IsIntType(expect):
		value, ok := big.NewInt(0).SetString(node.Value.Source(), 10)
		if !ok {
			panic("unreachable")
		}
		return &hir.Integer{
			Type:  expect,
			Value: value,
		}
	case hir.IsType[*hir.FloatType](expect):
		value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
		return &hir.Float{
			Type:  expect,
			Value: value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseChar(expect hir.Type, node *ast.Char) hir.ExprStmt {
	if expect == nil || !hir.IsNumberType(expect) {
		expect = self.pkgScope.I32()
	}
	s := node.Value.Source()
	char := util.ParseEscapeCharacter(s[1:len(s)-1], `\'`, `'`)[0]
	switch {
	case hir.IsIntType(expect):
		value := big.NewInt(int64(char))
		return &hir.Integer{
			Type:  expect,
			Value: value,
		}
	case hir.IsType[*hir.SintType](expect):
		value := big.NewFloat(float64(char))
		return &hir.Float{
			Type:  expect,
			Value: value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseFloat(expect hir.Type, node *ast.Float) *hir.Float {
	if expect == nil || !hir.IsType[*hir.FloatType](expect) {
		expect = self.pkgScope.F64()
	}
	value, _ := stlerror.MustWith2(big.NewFloat(0).Parse(node.Value.Source(), 10))
	return &hir.Float{
		Type:  expect,
		Value: value,
	}
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

func (self *Analyser) analyseBinary(expect hir.Type, node *ast.Binary) hir.Expr {
	left := self.analyseExpr(expect, node.Left)
	lt := left.GetType()
	right := self.analyseExpr(lt, node.Right)
	rt := right.GetType()

	switch node.Opera.Kind {
	case token.ASS:
		if lt.EqualTo(rt) {
			if left.Temporary() {
				errors.ThrowExprTemporaryError(node.Left.Position())
			} else if !left.Mutable() {
				errors.ThrowNotMutableError(node.Left.Position())
			}
			return hir.NewAssign(left, right)
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
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.And().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.And().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntAndInt{
				Left:  left,
				Right: right,
			}
		}
	case token.OR:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Or().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Or().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntOrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.XOR:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Xor().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Xor().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntXorInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHL:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Shl().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Shl().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntShlInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHR:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Shr().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Shr().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsIntType(lt) {
			return &hir.IntShrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.ADD:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Add().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Add().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumAddNum{
				Left:  left,
				Right: right,
			}
		}
	case token.SUB:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Sub().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Sub().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumSubNum{
				Left:  left,
				Right: right,
			}
		}
	case token.MUL:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Mul().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Mul().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumMulNum{
				Left:  left,
				Right: right,
			}
		}
	case token.DIV:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Div().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Div().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			if (stlbasic.Is[*hir.Integer](right) && right.(*hir.Integer).Value.Cmp(big.NewInt(0)) == 0) ||
				(stlbasic.Is[*hir.Float](right) && right.(*hir.Float).Value.Cmp(big.NewFloat(0)) == 0) {
				errors.ThrowDivZero(node.Right.Position())
			}
			return &hir.NumDivNum{
				Left:  left,
				Right: right,
			}
		}
	case token.REM:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Rem().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Rem().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			if (stlbasic.Is[*hir.Integer](right) && right.(*hir.Integer).Value.Cmp(big.NewInt(0)) == 0) ||
				(stlbasic.Is[*hir.Float](right) && right.(*hir.Float).Value.Cmp(big.NewFloat(0)) == 0) {
				errors.ThrowDivZero(node.Right.Position())
			}
			return &hir.NumRemNum{
				Left:  left,
				Right: right,
			}
		}
	case token.EQ:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Eq().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Eq().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) {
			if !lt.EqualTo(hir.NoThing) {
				return &hir.Equal{
					BoolType: self.pkgScope.Bool(),
					Left:     left,
					Right:    right,
				}
			}
		}
	case token.NE:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Eq().HasBeImpled(ct) {
			return &hir.BoolNegate{
				Value: hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Eq().FirstMethodName()).MustValue(), right),
			}
		} else if !lt.EqualTo(hir.NoThing) {
			return &hir.NotEqual{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.LT:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Lt().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Lt().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumLtNum{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.GT:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Gt().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Gt().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumGtNum{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.LE:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Lt().HasBeImpled(ct) && self.pkgScope.Eq().HasBeImpled(ct) {
			return &hir.BoolOrBool{
				Left:  hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Lt().FirstMethodName()).MustValue(), right),
				Right: hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Eq().FirstMethodName()).MustValue(), right),
			}
		} else if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumLeNum{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.GE:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Gt().HasBeImpled(ct) && self.pkgScope.Eq().HasBeImpled(ct) {
			return &hir.BoolOrBool{
				Left:  hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Gt().FirstMethodName()).MustValue(), right),
				Right: hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Eq().FirstMethodName()).MustValue(), right),
			}
		} else if lt.EqualTo(rt) && hir.IsNumberType(lt) {
			return &hir.NumGeNum{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.LAND:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Land().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Land().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && lt.EqualTo(self.pkgScope.Bool()) {
			return &hir.BoolAndBool{
				Left:  left,
				Right: right,
			}
		}
	case token.LOR:
		ct, ok := hir.TryCustomType(lt)
		if ok && self.pkgScope.Lor().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](left), self.pkgScope.Lor().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && lt.EqualTo(self.pkgScope.Bool()) {
			return &hir.BoolOrBool{
				Left:  left,
				Right: right,
			}
		}
	default:
		panic("unreachable")
	}

	errors.ThrowIllegalBinaryError(node.Position(), node.Opera, left, right)
	return nil
}

func (self *Analyser) analyseUnary(expect hir.Type, node *ast.Unary) hir.Expr {
	switch node.Opera.Kind {
	case token.SUB:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		ct, ok := hir.TryCustomType(vt)
		if ok && self.pkgScope.Neg().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](value), self.pkgScope.Neg().FirstMethodName()).MustValue())
		} else if hir.IsType[*hir.SintType](vt) || hir.IsType[*hir.FloatType](vt) {
			return &hir.NumNegate{Value: value}
		}
		errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
		return nil
	case token.NOT:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		ct, ok := hir.TryCustomType(vt)
		if ok && self.pkgScope.Not().HasBeImpled(ct) {
			return hir.NewCall(hir.LoopFindMethodWithSelf(ct, optional.Some[hir.Expr](value), self.pkgScope.Not().FirstMethodName()).MustValue())
		}
		switch {
		case hir.IsIntType(vt):
			return &hir.IntBitNegate{Value: value}
		case vt.EqualTo(self.pkgScope.Bool()):
			return &hir.BoolNegate{Value: value}
		default:
			errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
			return nil
		}
	case token.AND, token.AND_WITH_MUT:
		if expect != nil && hir.IsType[*hir.RefType](expect) {
			expect = hir.AsType[*hir.RefType](expect).Elem
		}
		value := self.analyseExpr(expect, node.Value)
		if value.Temporary() {
			errors.ThrowExprTemporaryError(node.Value.Position())
		} else if node.Opera.Is(token.AND_WITH_MUT) && !value.Mutable() {
			errors.ThrowNotMutableError(node.Value.Position())
		} else if localVarDef, ok := value.(*hir.LocalVarDef); ok {
			localVarDef.Escaped = true
		}
		return &hir.GetRef{
			Mut:   node.Opera.Is(token.AND_WITH_MUT),
			Value: value,
		}
	case token.MUL:
		if expect != nil {
			expect = hir.NewRefType(false, expect)
		}
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		if !hir.IsType[*hir.RefType](vt) {
			errors.ThrowExpectReferenceError(node.Value.Position(), vt)
		}
		return &hir.DeRef{Value: value}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseIdentExpr(node *ast.IdentExpr) hir.Ident {
	expr := self.analyseIdent((*ast.Ident)(node), true)
	if expr.IsNone() {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	return stlbasic.IgnoreWith(expr.MustValue().Left())
}

func (self *Analyser) analyseCall(node *ast.Call) hir.Expr {
	if dotNode, ok := node.Func.(*ast.Dot); ok && len(node.Args) == 1 {
		if identNode, ok := dotNode.From.(*ast.IdentExpr); ok {
			ident, ok := self.analyseIdent((*ast.Ident)(identNode)).Value()
			if ok {
				if t, ok := ident.Right(); ok {
					if et, ok := hir.TryType[*hir.EnumType](t); ok {
						// 枚举值
						fieldName := dotNode.Index.Source()
						if et.Fields.ContainKey(fieldName) {
							caseDef := et.Fields.Get(fieldName)
							if caseDef.Elem.IsSome() {
								elem := self.analyseExpr(caseDef.Elem.MustValue(), stlslices.First(node.Args))
								return &hir.Enum{
									From:  t,
									Field: fieldName,
									Elem:  optional.Some(elem),
								}
							}
						}
					}
				}
			}
		}
	}

	f := self.analyseExpr(nil, node.Func)
	ct, ok := hir.TryType[hir.CallableType](f.GetType())
	if !ok {
		errors.ThrowExpectCallableError(node.Func.Position(), f.GetType())
	}
	vararg := stlbasic.Is[*hir.FuncDef](f) && f.(*hir.FuncDef).VarArg
	if (!vararg && len(node.Args) != len(ct.GetParams())) || (vararg && len(node.Args) < len(ct.GetParams())) {
		errors.ThrowParameterNumberNotMatchError(node.Position(), uint(len(ct.GetParams())), uint(len(node.Args)))
	}
	args := stlslices.Map(node.Args, func(index int, item ast.Expr) hir.Expr {
		if vararg && index >= len(ct.GetParams()) {
			return self.analyseExpr(nil, item)
		} else {
			return self.expectExpr(ct.GetParams()[index], item)
		}
	})
	return hir.NewCall(f, args...)
}

func (self *Analyser) analyseTuple(expect hir.Type, node *ast.Tuple) hir.Expr {
	if len(node.Elems) == 1 && (expect == nil || !hir.IsType[*hir.TupleType](expect)) {
		return self.analyseExpr(expect, node.Elems[0])
	}

	elemExpects := make([]hir.Type, len(node.Elems))
	if expect != nil {
		if hir.IsType[*hir.TupleType](expect) {
			tt := hir.AsType[*hir.TupleType](expect)
			if len(tt.Elems) < len(node.Elems) {
				copy(elemExpects, tt.Elems)
			} else if len(tt.Elems) > len(node.Elems) {
				elemExpects = tt.Elems[:len(node.Elems)]
			} else {
				elemExpects = tt.Elems
			}
		}
	}
	elems := lo.Map(node.Elems, func(item ast.Expr, index int) hir.Expr {
		return self.analyseExpr(elemExpects[index], item)
	})
	return &hir.Tuple{Elems: elems}
}

func (self *Analyser) analyseCovert(node *ast.Covert) hir.Expr {
	tt := self.analyseType(node.Type)
	from := self.analyseExpr(tt, node.Value)
	ft := from.GetType()
	if v, ok := self.autoTypeCovert(tt, from); ok {
		return v
	}

	switch {
	case hir.ToBuildInType(ft).EqualTo(hir.ToBuildInType(tt)):
		return &hir.DoNothingCovert{
			From: from,
			To:   tt,
		}
	case hir.IsNumberType(ft) && hir.IsNumberType(tt):
		// i8 -> u8
		return &hir.Num2Num{
			From: from,
			To:   tt,
		}
	default:
		errors.ThrowIllegalCovertError(node.Position(), ft, tt)
		return nil
	}
}

func (self *Analyser) expectExpr(expect hir.Type, node ast.Expr) hir.Expr {
	value := self.analyseExpr(expect, node)
	newValue, ok := self.autoTypeCovert(expect, value)
	if !ok {
		errors.ThrowTypeMismatchError(node.Position(), value.GetType(), expect)
	}
	return newValue
}

// 自动类型转换
func (self *Analyser) autoTypeCovert(expect hir.Type, v hir.Expr) (hir.Expr, bool) {
	vt := v.GetType()
	if vt.EqualTo(expect) {
		return v, true
	}

	switch {
	case hir.IsType[*hir.RefType](vt) && hir.AsType[*hir.RefType](vt).Elem.EqualTo(expect):
		// &i8 -> i8
		return &hir.DeRef{Value: v}, true
	case hir.IsType[*hir.RefType](vt) && hir.IsType[*hir.RefType](expect) && hir.NewRefType(false, hir.AsType[*hir.RefType](vt).Elem).EqualTo(expect):
		// &mut i8 -> &i8
		return &hir.DoNothingCovert{
			From: v,
			To:   expect,
		}, true
	case vt.EqualTo(hir.NoReturn) && (self.hasTypeDefault(expect) || expect.EqualTo(hir.NoThing)):
		// X -> any
		return &hir.NoReturn2Any{
			From: v,
			To:   expect,
		}, true
	case hir.IsType[*hir.FuncType](vt) && hir.IsType[*hir.LambdaType](expect) && hir.AsType[*hir.FuncType](vt).EqualTo(hir.AsType[*hir.LambdaType](expect).ToFuncType()):
		// func -> ()->void
		return &hir.Func2Lambda{
			From: v,
			To:   expect,
		}, true
	case hir.IsType[*hir.EnumType](vt) && hir.IsType[*hir.UintType](expect) && hir.AsType[*hir.EnumType](vt).IsSimple() && hir.AsType[*hir.UintType](expect).EqualTo(hir.NewUintType(8)):
		// simple enum -> u8
		return &hir.Enum2Number{
			From: v,
			To:   expect,
		}, true
	case hir.IsType[*hir.UintType](vt) && hir.IsType[*hir.EnumType](expect) && hir.AsType[*hir.UintType](vt).EqualTo(hir.NewUintType(8)) && hir.AsType[*hir.EnumType](expect).IsSimple():
		// u8 -> simple enum
		return &hir.Number2Enum{
			From: v,
			To:   expect,
		}, true
	default:
		return v, false
	}
}

func (self *Analyser) analyseArray(expect hir.Type, node *ast.Array) *hir.Array {
	var expectArray *hir.ArrayType
	var expectElem hir.Type
	if expect != nil && hir.IsType[*hir.ArrayType](expect) {
		expectArray = hir.AsType[*hir.ArrayType](expect)
		expectElem = expectArray.Elem
	}
	if expectArray == nil && len(node.Elems) == 0 {
		errors.ThrowExpectArrayTypeError(node.Position(), hir.NoThing)
	}
	elems := make([]hir.Expr, len(node.Elems))
	for i, elemNode := range node.Elems {
		elems[i] = stlbasic.TernaryAction(i == 0, func() hir.Expr {
			return self.analyseExpr(expectElem, elemNode)
		}, func() hir.Expr {
			return self.expectExpr(elems[0].GetType(), elemNode)
		})
	}
	return &hir.Array{
		Type: stlbasic.TernaryAction(len(elems) != 0, func() hir.Type {
			return hir.NewArrayType(uint64(len(elems)), elems[0].GetType())
		}, func() hir.Type { return expectArray }),
		Elems: elems,
	}
}

func (self *Analyser) analyseIndex(node *ast.Index) *hir.Index {
	from := self.analyseExpr(nil, node.From)
	if !hir.IsType[*hir.ArrayType](from.GetType()) {
		errors.ThrowExpectArrayError(node.From.Position(), from.GetType())
	}
	at := hir.AsType[*hir.ArrayType](from.GetType())
	index := self.expectExpr(self.pkgScope.Usize(), node.Index)
	if stlbasic.Is[*hir.Integer](index) && index.(*hir.Integer).Value.Cmp(big.NewInt(int64(at.Size))) >= 0 {
		errors.ThrowIndexOutOfRange(node.Index.Position())
	}
	return &hir.Index{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseExtract(expect hir.Type, node *ast.Extract) *hir.Extract {
	indexValue, ok := big.NewInt(0).SetString(node.Index.Source(), 10)
	if !ok {
		panic("unreachable")
	}
	if !indexValue.IsUint64() {
		panic("unreachable")
	}
	index := uint(indexValue.Uint64())

	expectFrom := &hir.TupleType{Elems: make([]hir.Type, index+1)}
	expectFrom.Elems[index] = expect

	from := self.analyseExpr(expectFrom, node.From)
	if !hir.IsType[*hir.TupleType](from.GetType()) {
		errors.ThrowExpectTupleError(node.From.Position(), from.GetType())
	}

	tt := hir.AsType[*hir.TupleType](from.GetType())
	if index >= uint(len(tt.Elems)) {
		errors.ThrowInvalidIndexError(node.Index.Position, index)
	}
	return &hir.Extract{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseStruct(node *ast.Struct) *hir.Struct {
	stObj := self.analyseType(node.Type)
	st, ok := hir.TryType[*hir.StructType](stObj)
	if !ok {
		errors.ThrowExpectStructTypeError(node.Type.Position(), stObj)
	}

	existedFields := make(map[string]hir.Expr)
	for _, nf := range node.Fields {
		fn := nf.First.Source()
		if !st.Fields.ContainKey(fn) || (!self.pkgScope.pkg.Equal(st.Def.Pkg) && !st.Fields.Get(fn).Public) {
			errors.ThrowUnknownIdentifierError(nf.First.Position, nf.First)
		}
		existedFields[fn] = self.expectExpr(st.Fields.Get(fn).Type, nf.Second)
	}

	fields := make([]hir.Expr, st.Fields.Length())
	var i int
	for iter := st.Fields.Iterator(); iter.Next(); i++ {
		fn, ft := iter.Value().First, iter.Value().Second.Type
		if fv, ok := existedFields[fn]; ok {
			fields[i] = fv
		} else {
			fields[i] = self.getTypeDefaultValue(node.Type.Position(), ft)
		}
	}

	return &hir.Struct{
		Type:   stObj,
		Fields: fields,
	}
}

func (self *Analyser) analyseDot(node *ast.Dot) hir.Expr {
	fieldName := node.Index.Source()

	if identNode, ok := node.From.(*ast.IdentExpr); ok {
		ident, ok := self.analyseIdent((*ast.Ident)(identNode)).Value()
		if ok {
			if t, ok := ident.Right(); ok {
				if ct, ok := hir.TryCustomType(t); ok {
					// 静态方法
					f := ct.Methods.Get(fieldName)
					if f != nil && (f.GetPublic() || self.pkgScope.pkg.Equal(f.GetPackage())) {
						return f
					}
				}
				if et, ok := hir.TryType[*hir.EnumType](t); ok {
					// 枚举值
					if et.Fields.ContainKey(fieldName) {
						if et.Fields.Get(fieldName).Elem.IsNone() {
							return &hir.Enum{
								From:  t,
								Field: fieldName,
							}
						}
					}
				}
			}
		}
	}

	if method, ok := self.analyseMethod(node); ok {
		return method
	}

	fromValObj := self.analyseExpr(nil, node.From)
	ft := fromValObj.GetType()

	// 字段
	var structVal hir.Expr
	if hir.IsType[*hir.StructType](ft) {
		structVal = fromValObj
	} else if hir.IsType[*hir.RefType](ft) && hir.IsType[*hir.StructType](hir.AsType[*hir.RefType](ft).Elem) {
		structVal = &hir.DeRef{Value: fromValObj}
	} else {
		errors.ThrowExpectStructError(node.From.Position(), ft)
	}
	st := hir.AsType[*hir.StructType](structVal.GetType())
	if field := st.Fields.Get(fieldName); !st.Fields.ContainKey(fieldName) || (!field.Public && !self.pkgScope.pkg.Equal(st.Def.Pkg)) {
		errors.ThrowUnknownIdentifierError(node.Index.Position, node.Index)
	}
	return &hir.GetField{
		Internal: self.pkgScope.pkg.Equal(st.Def.Pkg),
		From:     structVal,
		Index:    uint(slices.Index(st.Fields.Keys().ToSlice(), fieldName)),
	}
}

func (self *Analyser) analyseMethod(node *ast.Dot) (hir.Expr, bool) {
	fieldName := node.Index.Source()
	fromObj := self.analyseExpr(nil, node.From)
	ft := fromObj.GetType()

	var customType *hir.CustomType
	if hir.IsCustomType(ft) {
		customType = hir.AsCustomType(ft)
	} else if hir.IsType[*hir.RefType](ft) && hir.IsCustomType(hir.AsType[*hir.RefType](ft).Elem) {
		customType = hir.AsCustomType(hir.AsType[*hir.RefType](ft).Elem)
	}
	if customType == nil {
		return nil, false
	}

	method := customType.Methods.Get(fieldName)
	if method == nil || (!method.GetPublic() && !self.pkgScope.pkg.Equal(customType.Pkg)) {
		return nil, false
	}

	var customFrom hir.Expr
	if method.IsStatic() {
		return method, true
	} else if method.IsRef() && hir.IsCustomType(ft) {
		if fromObj.Temporary() {
			errors.ThrowExprTemporaryError(node.From.Position())
		}
		customFrom = &hir.GetRef{
			Mut:   method.GetSelfParam().MustValue().Mutable(),
			Value: fromObj,
		}
	} else if method.IsRef() && !hir.IsCustomType(ft) {
		customFrom = fromObj
	} else if !method.IsRef() && hir.IsCustomType(ft) {
		customFrom = fromObj
	} else if !method.IsRef() && !hir.IsCustomType(ft) {
		customFrom = &hir.DeRef{Value: fromObj}
	} else {
		panic("unreachable")
	}

	return &hir.Method{
		Self:   customFrom,
		Define: method,
	}, true
}

func (self *Analyser) analyseString(node *ast.String) *hir.String {
	s := node.Value.Source()
	s = util.ParseEscapeCharacter(s[1:len(s)-1], `\"`, `"`)
	return &hir.String{
		Type:  self.pkgScope.Str(),
		Value: s,
	}
}

func (self *Analyser) analyseJudgment(node *ast.Judgment) hir.Expr {
	target := self.analyseType(node.Type)
	value := self.analyseExpr(target, node.Value)
	vt := value.GetType()

	switch {
	case vt.EqualTo(target):
		return self.pkgScope.True()
	default:
		return &hir.TypeJudgment{
			BoolType: self.pkgScope.Bool(),
			Value:    value,
			Type:     target,
		}
	}
}

func (self *Analyser) analyseLambda(expect hir.Type, node *ast.Lambda) *hir.Lambda {
	params := stlslices.Map(node.Params, func(_ int, e ast.Param) *hir.Param { return self.analyseParam(e) })
	ret := self.analyseOptionTypeWith(optional.Some(node.Ret), voidTypeAnalyser, noReturnTypeAnalyser)
	f := &hir.Lambda{
		Params: params,
		Ret:    ret,
	}

	if expect == nil || !hir.IsType[*hir.LambdaType](expect) || !hir.AsType[*hir.LambdaType](expect).ToFuncType().EqualTo(f.GetFuncType()) {
		expect = hir.NewLambdaType(f.Ret, stlslices.Map(f.Params, func(_ int, e *hir.Param) hir.Type {
			return e.GetType()
		})...)
	}
	f.Type = expect

	captureIdents := hashset.NewHashSet[hir.Ident]()
	self.localScope = _NewLambdaScope(self.localScope, f, func(ident hir.Ident) {
		if v, ok := ident.(*hir.LocalVarDef); ok {
			v.Escaped = true
		}
		captureIdents.Add(ident)
	})
	defer func() {
		self.localScope = self.localScope.GetParent().(_LocalScope)
	}()

	for i, p := range f.Params {
		if p.Name.IsSome() && !self.localScope.SetValue(p.Name.MustValue(), p) {
			errors.ThrowIdentifierDuplicationError(node.Params[i].Name.MustValue().Position, node.Params[i].Name.MustValue())
		}
	}

	f.Body = self.analyseFuncBody(node.Body)
	f.Context = captureIdents.ToSlice().ToSlice()
	return f
}

func (self *Analyser) analyseParam(node ast.Param) *hir.Param {
	return &hir.Param{
		Mut:  !node.Mutable.IsNone(),
		Type: self.analyseType(node.Type),
		Name: stlbasic.TernaryAction(node.Name.IsNone(), func() optional.Optional[string] {
			return optional.None[string]()
		}, func() optional.Optional[string] {
			return optional.Some(node.Name.MustValue().Source())
		}),
	}
}
