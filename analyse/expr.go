package analyse

import (
	"math/big"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/token"
)

func (self *Analyser) analyseExpr(expect Type, node ast.Expr) Expr {
	switch exprNode := node.(type) {
	case *ast.Integer:
		return self.analyseInteger(expect, exprNode)
	case *ast.Float:
		return self.analyseFloat(expect, exprNode)
	case *ast.Boolean:
		return self.analyseBool(expect, exprNode)
	case *ast.Binary:
		return self.analyseBinary(expect, exprNode)
	case *ast.Unary:
		return self.analyseUnary(expect, exprNode)
	case *ast.Ident:
		return self.analyseIdent(expect, exprNode)
	case *ast.Call:
		return self.analyseCall(expect, exprNode)
	case *ast.Unit:
		return self.analyseUnit(expect, exprNode)
	case *ast.Covert:
		return self.analyseCovert(exprNode)
	case *ast.Array:
		return self.analyseArray(exprNode)
	case *ast.Index:
		return self.analyseIndex(exprNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseInteger(expect Type, node *ast.Integer) Expr {
	if expect == nil || !TypeIs[NumberType](expect) {
		expect = Isize
	}
	switch t := expect.(type) {
	case IntType:
		value, ok := big.NewInt(0).SetString(node.Value.Source(), 10)
		if !ok {
			panic("unreachable")
		}
		return &Integer{
			Type:  t,
			Value: *value,
		}
	case *FloatType:
		value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
		return &Float{
			Type:  t,
			Value: *value,
		}
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseFloat(expect Type, node *ast.Float) *Float {
	if expect == nil || !TypeIs[*FloatType](expect) {
		expect = F64
	}
	value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
	return &Float{
		Type:  expect.(*FloatType),
		Value: *value,
	}
}

func (self *Analyser) analyseBool(expect Type, node *ast.Boolean) *Boolean {
	return &Boolean{Value: node.Value.Is(token.TRUE)}
}

func (self *Analyser) analyseBinary(expect Type, node *ast.Binary) Binary {
	left, right := self.analyseExpr(expect, node.Left), self.analyseExpr(expect, node.Right)
	lt, rt := left.GetType(), right.GetType()

	switch node.Opera.Kind {
	case token.AND:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			return &IntAndInt{
				Left:  left,
				Right: right,
			}
		}
	case token.OR:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			return &IntOrInt{
				Left:  left,
				Right: right,
			}
		}
	case token.XOR:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			return &IntXorInt{
				Left:  left,
				Right: right,
			}
		}
	case token.ADD:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumAddNum{
				Left:  left,
				Right: right,
			}
		}
	case token.SUB:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumSubNum{
				Left:  left,
				Right: right,
			}
		}
	case token.MUL:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumMulNum{
				Left:  left,
				Right: right,
			}
		}
	case token.DIV:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumDivNum{
				Left:  left,
				Right: right,
			}
		}
	case token.REM:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumRemNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LT:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumLtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GT:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumGtNum{
				Left:  left,
				Right: right,
			}
		}
	case token.LE:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumLeNum{
				Left:  left,
				Right: right,
			}
		}
	case token.GE:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			return &NumGeNum{
				Left:  left,
				Right: right,
			}
		}
	default:
		panic("unreachable")
	}

	// TODO: 编译时异常：不能对两个类型进行二元运算
	panic("unreachable")
}

func (self *Analyser) analyseUnary(expect Type, node *ast.Unary) Unary {
	value := self.analyseExpr(expect, node.Value)
	vt := value.GetType()

	switch node.Opera.Kind {
	case token.SUB:
		if TypeIs[*SintType](vt) || TypeIs[*FloatType](vt) {
			return &NumNegate{Value: value}
		}
	default:
		panic("unreachable")
	}

	// TODO: 编译时异常：不能对两个类型进行一元运算
	panic("unreachable")
}

func (self *Analyser) analyseIdent(expect Type, node *ast.Ident) Ident {
	value, ok := self.localScope.GetValue(node.Name.Source())
	if !ok {
		// TODO: 编译时异常：未知的变量
		panic("unreachable")
	}
	return value
}

func (self *Analyser) analyseCall(expect Type, node *ast.Call) *Call {
	f := self.analyseExpr(nil, node.Func)
	return &Call{Func: f}
}

func (self *Analyser) analyseUnit(expect Type, node *ast.Unit) Expr {
	return self.analyseExpr(expect, node.Value)
}

func (self *Analyser) analyseCovert(node *ast.Covert) Expr {
	tt := self.analyseType(node.Type)
	from := self.analyseExpr(tt, node.Value)
	ft := from.GetType()
	if ft.Equal(tt) {
		return from
	}

	switch {
	case TypeIs[NumberType](ft) && TypeIs[NumberType](tt):
		return &Num2Num{
			From: from,
			To:   tt.(NumberType),
		}
	default:
		// TODO: 编译期异常：类型A无法转换成B
		panic("unreachable")
	}
}

func (self *Analyser) expectExpr(expect Type, node ast.Expr) Expr {
	value := self.analyseExpr(expect, node)
	if !value.GetType().Equal(expect) {
		// TODO: 编译时异常：期待类型是A，但是这里是B
		panic("unreachable")
	}
	return value
}

func (self *Analyser) analyseArray(node *ast.Array) *Array {
	t := self.analyseArrayType(node.Type)
	elems := make([]Expr, len(node.Elems))
	for i, en := range node.Elems {
		elems[i] = self.expectExpr(t.Elem, en)
	}
	return &Array{
		Type:  t,
		Elems: elems,
	}
}

func (self *Analyser) analyseIndex(node *ast.Index) *Index {
	from := self.analyseExpr(nil, node.From)
	if !TypeIs[*ArrayType](from.GetType()) {
		// TODO: 编译时异常：不能获取类型A的索引
		panic("unreachable")
	}
	index := self.expectExpr(Usize, node.Index)
	return &Index{
		From:  from,
		Index: index,
	}
}
