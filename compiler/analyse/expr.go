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

func (self *Analyser) analyseExpr(expect oldhir.Type, node ast.Expr) oldhir.Expr {
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

func (self *Analyser) analyseInteger(expect oldhir.Type, node *ast.Integer) oldhir.ExprStmt {
	if expect == nil || !oldhir.IsNumberType(expect) {
		expect = self.pkgScope.Isize()
	}
	switch {
	case oldhir.IsIntType(expect):
		value, ok := big.NewInt(0).SetString(node.Value.Source(), 10)
		if !ok {
			panic("unreachable")
		}
		return &oldhir.Integer{
			Type:  expect,
			Value: value,
		}
	case oldhir.IsType[*oldhir.FloatType](expect):
		value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
		return &oldhir.Float{
			Type:  expect,
			Value: value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseChar(expect oldhir.Type, node *ast.Char) oldhir.ExprStmt {
	if expect == nil || !oldhir.IsNumberType(expect) {
		expect = self.pkgScope.I32()
	}
	s := node.Value.Source()
	char := util.ParseEscapeCharacter(s[1:len(s)-1], `\'`, `'`)[0]
	switch {
	case oldhir.IsIntType(expect):
		value := big.NewInt(int64(char))
		return &oldhir.Integer{
			Type:  expect,
			Value: value,
		}
	case oldhir.IsType[*oldhir.SintType](expect):
		value := big.NewFloat(float64(char))
		return &oldhir.Float{
			Type:  expect,
			Value: value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseFloat(expect oldhir.Type, node *ast.Float) *oldhir.Float {
	if expect == nil || !oldhir.IsType[*oldhir.FloatType](expect) {
		expect = self.pkgScope.F64()
	}
	value, _ := stlerror.MustWith2(big.NewFloat(0).Parse(node.Value.Source(), 10))
	return &oldhir.Float{
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

func (self *Analyser) analyseBinary(expect oldhir.Type, node *ast.Binary) oldhir.Expr {
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
			return oldhir.NewAssign(left, right)
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
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.And().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.And().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsIntType(lt) {
			return &oldhir.IntAndInt{
				Left:  left,
				Right: right,
			}
		}
	case token.OR:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Or().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Or().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsIntType(lt) {
			return &oldhir.IntOrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.XOR:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Xor().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Xor().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsIntType(lt) {
			return &oldhir.IntXorInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHL:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Shl().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Shl().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsIntType(lt) {
			return &oldhir.IntShlInt{
				Left:  left,
				Right: right,
			}
		}
	case token.SHR:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Shr().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Shr().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsIntType(lt) {
			return &oldhir.IntShrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.ADD:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Add().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Add().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsNumberType(lt) {
			return &oldhir.NumAddNum{
				Left:  left,
				Right: right,
			}
		}
	case token.SUB:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Sub().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Sub().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsNumberType(lt) {
			return &oldhir.NumSubNum{
				Left:  left,
				Right: right,
			}
		}
	case token.MUL:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Mul().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Mul().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsNumberType(lt) {
			return &oldhir.NumMulNum{
				Left:  left,
				Right: right,
			}
		}
	case token.DIV:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Div().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Div().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsNumberType(lt) {
			if (stlbasic.Is[*oldhir.Integer](right) && right.(*oldhir.Integer).Value.Cmp(big.NewInt(0)) == 0) ||
				(stlbasic.Is[*oldhir.Float](right) && right.(*oldhir.Float).Value.Cmp(big.NewFloat(0)) == 0) {
				errors.ThrowDivZero(node.Right.Position())
			}
			return &oldhir.NumDivNum{
				Left:  left,
				Right: right,
			}
		}
	case token.REM:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Rem().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Rem().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsNumberType(lt) {
			if (stlbasic.Is[*oldhir.Integer](right) && right.(*oldhir.Integer).Value.Cmp(big.NewInt(0)) == 0) ||
				(stlbasic.Is[*oldhir.Float](right) && right.(*oldhir.Float).Value.Cmp(big.NewFloat(0)) == 0) {
				errors.ThrowDivZero(node.Right.Position())
			}
			return &oldhir.NumRemNum{
				Left:  left,
				Right: right,
			}
		}
	case token.EQ:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Eq().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Eq().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) {
			if !lt.EqualTo(oldhir.NoThing) {
				return &oldhir.Equal{
					BoolType: self.pkgScope.Bool(),
					Left:     left,
					Right:    right,
				}
			}
		}
	case token.NE:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Eq().HasBeImpled(ct) {
			return &oldhir.BoolNegate{
				Value: oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Eq().FirstMethodName()).MustValue(), right),
			}
		} else if !lt.EqualTo(oldhir.NoThing) {
			return &oldhir.NotEqual{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.LT:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Lt().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Lt().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsNumberType(lt) {
			return &oldhir.NumLtNum{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.GT:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Gt().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Gt().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && oldhir.IsNumberType(lt) {
			return &oldhir.NumGtNum{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.LE:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Lt().HasBeImpled(ct) && self.pkgScope.Eq().HasBeImpled(ct) {
			return &oldhir.BoolOrBool{
				Left:  oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Lt().FirstMethodName()).MustValue(), right),
				Right: oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Eq().FirstMethodName()).MustValue(), right),
			}
		} else if lt.EqualTo(rt) && oldhir.IsNumberType(lt) {
			return &oldhir.NumLeNum{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.GE:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Gt().HasBeImpled(ct) && self.pkgScope.Eq().HasBeImpled(ct) {
			return &oldhir.BoolOrBool{
				Left:  oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Gt().FirstMethodName()).MustValue(), right),
				Right: oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Eq().FirstMethodName()).MustValue(), right),
			}
		} else if lt.EqualTo(rt) && oldhir.IsNumberType(lt) {
			return &oldhir.NumGeNum{
				BoolType: self.pkgScope.Bool(),
				Left:     left,
				Right:    right,
			}
		}
	case token.LAND:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Land().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Land().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && lt.EqualTo(self.pkgScope.Bool()) {
			return &oldhir.BoolAndBool{
				Left:  left,
				Right: right,
			}
		}
	case token.LOR:
		ct, ok := oldhir.TryCustomType(lt)
		if ok && self.pkgScope.Lor().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](left), self.pkgScope.Lor().FirstMethodName()).MustValue(), right)
		} else if lt.EqualTo(rt) && lt.EqualTo(self.pkgScope.Bool()) {
			return &oldhir.BoolOrBool{
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

func (self *Analyser) analyseUnary(expect oldhir.Type, node *ast.Unary) oldhir.Expr {
	switch node.Opera.Kind {
	case token.SUB:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		ct, ok := oldhir.TryCustomType(vt)
		if ok && self.pkgScope.Neg().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](value), self.pkgScope.Neg().FirstMethodName()).MustValue())
		} else if oldhir.IsType[*oldhir.SintType](vt) || oldhir.IsType[*oldhir.FloatType](vt) {
			return &oldhir.NumNegate{Value: value}
		}
		errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
		return nil
	case token.NOT:
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		ct, ok := oldhir.TryCustomType(vt)
		if ok && self.pkgScope.Not().HasBeImpled(ct) {
			return oldhir.NewCall(oldhir.LoopFindMethodWithSelf(ct, optional.Some[oldhir.Expr](value), self.pkgScope.Not().FirstMethodName()).MustValue())
		}
		switch {
		case oldhir.IsIntType(vt):
			return &oldhir.IntBitNegate{Value: value}
		case vt.EqualTo(self.pkgScope.Bool()):
			return &oldhir.BoolNegate{Value: value}
		default:
			errors.ThrowIllegalUnaryError(node.Position(), node.Opera, vt)
			return nil
		}
	case token.AND, token.AND_WITH_MUT:
		if expect != nil && oldhir.IsType[*oldhir.RefType](expect) {
			expect = oldhir.AsType[*oldhir.RefType](expect).Elem
		}
		value := self.analyseExpr(expect, node.Value)
		if value.Temporary() {
			errors.ThrowExprTemporaryError(node.Value.Position())
		} else if node.Opera.Is(token.AND_WITH_MUT) && !value.Mutable() {
			errors.ThrowNotMutableError(node.Value.Position())
		} else if localVarDef, ok := value.(*oldhir.LocalVarDef); ok {
			localVarDef.Escaped = true
		}
		return &oldhir.GetRef{
			Mut:   node.Opera.Is(token.AND_WITH_MUT),
			Value: value,
		}
	case token.MUL:
		if expect != nil {
			expect = oldhir.NewRefType(false, expect)
		}
		value := self.analyseExpr(expect, node.Value)
		vt := value.GetType()
		if !oldhir.IsType[*oldhir.RefType](vt) {
			errors.ThrowExpectReferenceError(node.Value.Position(), vt)
		}
		return &oldhir.DeRef{Value: value}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseIdentExpr(node *ast.IdentExpr) oldhir.Ident {
	expr := self.analyseIdent((*ast.Ident)(node), true)
	if expr.IsNone() {
		errors.ThrowUnknownIdentifierError(node.Name.Position, node.Name)
	}
	return stlbasic.IgnoreWith(expr.MustValue().Left())
}

func (self *Analyser) analyseCall(node *ast.Call) oldhir.Expr {
	if dotNode, ok := node.Func.(*ast.Dot); ok && len(node.Args) == 1 {
		if identNode, ok := dotNode.From.(*ast.IdentExpr); ok {
			ident, ok := self.analyseIdent((*ast.Ident)(identNode)).Value()
			if ok {
				if t, ok := ident.Right(); ok {
					if et, ok := oldhir.TryType[*oldhir.EnumType](t); ok {
						// 枚举值
						fieldName := dotNode.Index.Source()
						if et.Fields.ContainKey(fieldName) {
							caseDef := et.Fields.Get(fieldName)
							if caseDef.Elem.IsSome() {
								elem := self.analyseExpr(caseDef.Elem.MustValue(), stlslices.First(node.Args))
								return &oldhir.Enum{
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
	ct, ok := oldhir.TryType[oldhir.CallableType](f.GetType())
	if !ok {
		errors.ThrowExpectCallableError(node.Func.Position(), f.GetType())
	}
	vararg := stlbasic.Is[*oldhir.FuncDef](f) && f.(*oldhir.FuncDef).VarArg
	if (!vararg && len(node.Args) != len(ct.GetParams())) || (vararg && len(node.Args) < len(ct.GetParams())) {
		errors.ThrowParameterNumberNotMatchError(node.Position(), uint(len(ct.GetParams())), uint(len(node.Args)))
	}
	args := stlslices.Map(node.Args, func(index int, item ast.Expr) oldhir.Expr {
		if vararg && index >= len(ct.GetParams()) {
			return self.analyseExpr(nil, item)
		} else {
			return self.expectExpr(ct.GetParams()[index], item)
		}
	})
	return oldhir.NewCall(f, args...)
}

func (self *Analyser) analyseTuple(expect oldhir.Type, node *ast.Tuple) oldhir.Expr {
	if len(node.Elems) == 1 && (expect == nil || !oldhir.IsType[*oldhir.TupleType](expect)) {
		return self.analyseExpr(expect, node.Elems[0])
	}

	elemExpects := make([]oldhir.Type, len(node.Elems))
	if expect != nil {
		if oldhir.IsType[*oldhir.TupleType](expect) {
			tt := oldhir.AsType[*oldhir.TupleType](expect)
			if len(tt.Elems) < len(node.Elems) {
				copy(elemExpects, tt.Elems)
			} else if len(tt.Elems) > len(node.Elems) {
				elemExpects = tt.Elems[:len(node.Elems)]
			} else {
				elemExpects = tt.Elems
			}
		}
	}
	elems := lo.Map(node.Elems, func(item ast.Expr, index int) oldhir.Expr {
		return self.analyseExpr(elemExpects[index], item)
	})
	return &oldhir.Tuple{Elems: elems}
}

func (self *Analyser) analyseCovert(node *ast.Covert) oldhir.Expr {
	tt := self.analyseType(node.Type)
	from := self.analyseExpr(tt, node.Value)
	ft := from.GetType()
	if v, ok := self.autoTypeCovert(tt, from); ok {
		return v
	}

	switch {
	case oldhir.ToBuildInType(ft).EqualTo(oldhir.ToBuildInType(tt)):
		return &oldhir.DoNothingCovert{
			From: from,
			To:   tt,
		}
	case oldhir.IsNumberType(ft) && oldhir.IsNumberType(tt):
		// i8 -> u8
		return &oldhir.Num2Num{
			From: from,
			To:   tt,
		}
	default:
		errors.ThrowIllegalCovertError(node.Position(), ft, tt)
		return nil
	}
}

func (self *Analyser) expectExpr(expect oldhir.Type, node ast.Expr) oldhir.Expr {
	value := self.analyseExpr(expect, node)
	newValue, ok := self.autoTypeCovert(expect, value)
	if !ok {
		errors.ThrowTypeMismatchError(node.Position(), value.GetType(), expect)
	}
	return newValue
}

// 自动类型转换
func (self *Analyser) autoTypeCovert(expect oldhir.Type, v oldhir.Expr) (oldhir.Expr, bool) {
	vt := v.GetType()
	if vt.EqualTo(expect) {
		return v, true
	}

	switch {
	case oldhir.IsType[*oldhir.RefType](vt) && oldhir.AsType[*oldhir.RefType](vt).Elem.EqualTo(expect):
		// &i8 -> i8
		return &oldhir.DeRef{Value: v}, true
	case oldhir.IsType[*oldhir.RefType](vt) && oldhir.IsType[*oldhir.RefType](expect) && oldhir.NewRefType(false, oldhir.AsType[*oldhir.RefType](vt).Elem).EqualTo(expect):
		// &mut i8 -> &i8
		return &oldhir.DoNothingCovert{
			From: v,
			To:   expect,
		}, true
	case vt.EqualTo(oldhir.NoReturn) && (self.hasTypeDefault(expect) || expect.EqualTo(oldhir.NoThing)):
		// X -> any
		return &oldhir.NoReturn2Any{
			From: v,
			To:   expect,
		}, true
	case oldhir.IsType[*oldhir.FuncType](vt) && oldhir.IsType[*oldhir.LambdaType](expect) && oldhir.AsType[*oldhir.FuncType](vt).EqualTo(oldhir.AsType[*oldhir.LambdaType](expect).ToFuncType()):
		// func -> ()->void
		return &oldhir.Func2Lambda{
			From: v,
			To:   expect,
		}, true
	case oldhir.IsType[*oldhir.EnumType](vt) && oldhir.IsType[*oldhir.UintType](expect) && oldhir.AsType[*oldhir.EnumType](vt).IsSimple() && oldhir.AsType[*oldhir.UintType](expect).EqualTo(oldhir.NewUintType(8)):
		// simple enum -> u8
		return &oldhir.Enum2Number{
			From: v,
			To:   expect,
		}, true
	case oldhir.IsType[*oldhir.UintType](vt) && oldhir.IsType[*oldhir.EnumType](expect) && oldhir.AsType[*oldhir.UintType](vt).EqualTo(oldhir.NewUintType(8)) && oldhir.AsType[*oldhir.EnumType](expect).IsSimple():
		// u8 -> simple enum
		return &oldhir.Number2Enum{
			From: v,
			To:   expect,
		}, true
	default:
		return v, false
	}
}

func (self *Analyser) analyseArray(expect oldhir.Type, node *ast.Array) *oldhir.Array {
	var expectArray *oldhir.ArrayType
	var expectElem oldhir.Type
	if expect != nil && oldhir.IsType[*oldhir.ArrayType](expect) {
		expectArray = oldhir.AsType[*oldhir.ArrayType](expect)
		expectElem = expectArray.Elem
	}
	if expectArray == nil && len(node.Elems) == 0 {
		errors.ThrowExpectArrayTypeError(node.Position(), oldhir.NoThing)
	}
	elems := make([]oldhir.Expr, len(node.Elems))
	for i, elemNode := range node.Elems {
		elems[i] = stlbasic.TernaryAction(i == 0, func() oldhir.Expr {
			return self.analyseExpr(expectElem, elemNode)
		}, func() oldhir.Expr {
			return self.expectExpr(elems[0].GetType(), elemNode)
		})
	}
	return &oldhir.Array{
		Type: stlbasic.TernaryAction(len(elems) != 0, func() oldhir.Type {
			return oldhir.NewArrayType(uint64(len(elems)), elems[0].GetType())
		}, func() oldhir.Type { return expectArray }),
		Elems: elems,
	}
}

func (self *Analyser) analyseIndex(node *ast.Index) *oldhir.Index {
	from := self.analyseExpr(nil, node.From)
	if !oldhir.IsType[*oldhir.ArrayType](from.GetType()) {
		errors.ThrowExpectArrayError(node.From.Position(), from.GetType())
	}
	at := oldhir.AsType[*oldhir.ArrayType](from.GetType())
	index := self.expectExpr(self.pkgScope.Usize(), node.Index)
	if stlbasic.Is[*oldhir.Integer](index) && index.(*oldhir.Integer).Value.Cmp(big.NewInt(int64(at.Size))) >= 0 {
		errors.ThrowIndexOutOfRange(node.Index.Position())
	}
	return &oldhir.Index{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseExtract(expect oldhir.Type, node *ast.Extract) *oldhir.Extract {
	indexValue, ok := big.NewInt(0).SetString(node.Index.Source(), 10)
	if !ok {
		panic("unreachable")
	}
	if !indexValue.IsUint64() {
		panic("unreachable")
	}
	index := uint(indexValue.Uint64())

	expectFrom := &oldhir.TupleType{Elems: make([]oldhir.Type, index+1)}
	expectFrom.Elems[index] = expect

	from := self.analyseExpr(expectFrom, node.From)
	if !oldhir.IsType[*oldhir.TupleType](from.GetType()) {
		errors.ThrowExpectTupleError(node.From.Position(), from.GetType())
	}

	tt := oldhir.AsType[*oldhir.TupleType](from.GetType())
	if index >= uint(len(tt.Elems)) {
		errors.ThrowInvalidIndexError(node.Index.Position, index)
	}
	return &oldhir.Extract{
		From:  from,
		Index: index,
	}
}

func (self *Analyser) analyseStruct(node *ast.Struct) *oldhir.Struct {
	stObj := self.analyseType(node.Type)
	st, ok := oldhir.TryType[*oldhir.StructType](stObj)
	if !ok {
		errors.ThrowExpectStructTypeError(node.Type.Position(), stObj)
	}

	existedFields := make(map[string]oldhir.Expr)
	for _, nf := range node.Fields {
		fn := nf.First.Source()
		if !st.Fields.ContainKey(fn) || (!self.pkgScope.pkg.Equal(st.Def.Pkg) && !st.Fields.Get(fn).Public) {
			errors.ThrowUnknownIdentifierError(nf.First.Position, nf.First)
		}
		existedFields[fn] = self.expectExpr(st.Fields.Get(fn).Type, nf.Second)
	}

	fields := make([]oldhir.Expr, st.Fields.Length())
	var i int
	for iter := st.Fields.Iterator(); iter.Next(); i++ {
		fn, ft := iter.Value().First, iter.Value().Second.Type
		if fv, ok := existedFields[fn]; ok {
			fields[i] = fv
		} else {
			fields[i] = self.getTypeDefaultValue(node.Type.Position(), ft)
		}
	}

	return &oldhir.Struct{
		Type:   stObj,
		Fields: fields,
	}
}

func (self *Analyser) analyseDot(node *ast.Dot) oldhir.Expr {
	fieldName := node.Index.Source()

	if identNode, ok := node.From.(*ast.IdentExpr); ok {
		ident, ok := self.analyseIdent((*ast.Ident)(identNode)).Value()
		if ok {
			if t, ok := ident.Right(); ok {
				if ct, ok := oldhir.TryCustomType(t); ok {
					// 静态方法
					f := ct.Methods.Get(fieldName)
					if f != nil && (f.GetPublic() || self.pkgScope.pkg.Equal(f.GetPackage())) {
						return f
					}
				}
				if et, ok := oldhir.TryType[*oldhir.EnumType](t); ok {
					// 枚举值
					if et.Fields.ContainKey(fieldName) {
						if et.Fields.Get(fieldName).Elem.IsNone() {
							return &oldhir.Enum{
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
	var structVal oldhir.Expr
	if oldhir.IsType[*oldhir.StructType](ft) {
		structVal = fromValObj
	} else if oldhir.IsType[*oldhir.RefType](ft) && oldhir.IsType[*oldhir.StructType](oldhir.AsType[*oldhir.RefType](ft).Elem) {
		structVal = &oldhir.DeRef{Value: fromValObj}
	} else {
		errors.ThrowExpectStructError(node.From.Position(), ft)
	}
	st := oldhir.AsType[*oldhir.StructType](structVal.GetType())
	if field := st.Fields.Get(fieldName); !st.Fields.ContainKey(fieldName) || (!field.Public && !self.pkgScope.pkg.Equal(st.Def.Pkg)) {
		errors.ThrowUnknownIdentifierError(node.Index.Position, node.Index)
	}
	return &oldhir.GetField{
		Internal: self.pkgScope.pkg.Equal(st.Def.Pkg),
		From:     structVal,
		Index:    uint(slices.Index(st.Fields.Keys().ToSlice(), fieldName)),
	}
}

func (self *Analyser) analyseMethod(node *ast.Dot) (oldhir.Expr, bool) {
	fieldName := node.Index.Source()
	fromObj := self.analyseExpr(nil, node.From)
	ft := fromObj.GetType()

	var customType *oldhir.CustomType
	if oldhir.IsCustomType(ft) {
		customType = oldhir.AsCustomType(ft)
	} else if oldhir.IsType[*oldhir.RefType](ft) && oldhir.IsCustomType(oldhir.AsType[*oldhir.RefType](ft).Elem) {
		customType = oldhir.AsCustomType(oldhir.AsType[*oldhir.RefType](ft).Elem)
	}
	if customType == nil {
		return nil, false
	}

	method := customType.Methods.Get(fieldName)
	if method == nil || (!method.GetPublic() && !self.pkgScope.pkg.Equal(customType.Pkg)) {
		return nil, false
	}

	var customFrom oldhir.Expr
	if method.IsStatic() {
		return method, true
	} else if method.IsRef() && oldhir.IsCustomType(ft) {
		if fromObj.Temporary() {
			errors.ThrowExprTemporaryError(node.From.Position())
		}
		customFrom = &oldhir.GetRef{
			Mut:   method.GetSelfParam().MustValue().Mutable(),
			Value: fromObj,
		}
	} else if method.IsRef() && !oldhir.IsCustomType(ft) {
		customFrom = fromObj
	} else if !method.IsRef() && oldhir.IsCustomType(ft) {
		customFrom = fromObj
	} else if !method.IsRef() && !oldhir.IsCustomType(ft) {
		customFrom = &oldhir.DeRef{Value: fromObj}
	} else {
		panic("unreachable")
	}

	return &oldhir.Method{
		Self:   customFrom,
		Define: method,
	}, true
}

func (self *Analyser) analyseString(node *ast.String) *oldhir.String {
	s := node.Value.Source()
	s = util.ParseEscapeCharacter(s[1:len(s)-1], `\"`, `"`)
	return &oldhir.String{
		Type:  self.pkgScope.Str(),
		Value: s,
	}
}

func (self *Analyser) analyseJudgment(node *ast.Judgment) oldhir.Expr {
	target := self.analyseType(node.Type)
	value := self.analyseExpr(target, node.Value)
	vt := value.GetType()

	switch {
	case vt.EqualTo(target):
		return self.pkgScope.True()
	default:
		return &oldhir.TypeJudgment{
			BoolType: self.pkgScope.Bool(),
			Value:    value,
			Type:     target,
		}
	}
}

func (self *Analyser) analyseLambda(expect oldhir.Type, node *ast.Lambda) *oldhir.Lambda {
	params := stlslices.Map(node.Params, func(_ int, e ast.Param) *oldhir.Param { return self.analyseParam(e) })
	ret := self.analyseOptionTypeWith(optional.Some(node.Ret), voidTypeAnalyser, noReturnTypeAnalyser)
	f := &oldhir.Lambda{
		Params: params,
		Ret:    ret,
	}

	if expect == nil || !oldhir.IsType[*oldhir.LambdaType](expect) || !oldhir.AsType[*oldhir.LambdaType](expect).ToFuncType().EqualTo(f.GetFuncType()) {
		expect = oldhir.NewLambdaType(f.Ret, stlslices.Map(f.Params, func(_ int, e *oldhir.Param) oldhir.Type {
			return e.GetType()
		})...)
	}
	f.Type = expect

	captureIdents := hashset.NewHashSet[oldhir.Ident]()
	self.localScope = _NewLambdaScope(self.localScope, f, func(ident oldhir.Ident) {
		if v, ok := ident.(*oldhir.LocalVarDef); ok {
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

func (self *Analyser) analyseParam(node ast.Param) *oldhir.Param {
	return &oldhir.Param{
		Mut:  !node.Mutable.IsNone(),
		Type: self.analyseType(node.Type),
		Name: stlbasic.TernaryAction(node.Name.IsNone(), func() optional.Optional[string] {
			return optional.None[string]()
		}, func() optional.Optional[string] {
			return optional.Some(node.Name.MustValue().Source())
		}),
	}
}
