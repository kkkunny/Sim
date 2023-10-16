package analyse

import (
	"math/big"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/token"
)

func (self *Analyser) analyseExpr(node ast.Expr) Expr {
	switch exprNode := node.(type) {
	case *ast.Integer:
		return self.analyseInteger(exprNode)
	case *ast.Float:
		return self.analyseFloat(exprNode)
	case *ast.Boolean:
		return self.analyseBool(exprNode)
	case *ast.Binary:
		return self.analyseBinary(exprNode)
	case *ast.Unary:
		return self.analyseUnary(exprNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseInteger(node *ast.Integer) *Integer {
	value, ok := big.NewInt(0).SetString(node.Value.Source(), 10)
	if !ok {
		panic("unreachable")
	}
	return &Integer{
		Type:  Isize,
		Value: *value,
	}
}

func (self *Analyser) analyseFloat(node *ast.Float) *Float {
	value, _ := stlerror.MustWith2(big.ParseFloat(node.Value.Source(), 10, big.MaxPrec, big.ToZero))
	return &Float{
		Type:  F64,
		Value: *value,
	}
}

func (self *Analyser) analyseBool(node *ast.Boolean) *Boolean {
	return &Boolean{Value: node.Value.Is(token.TRUE)}
}

func (self *Analyser) analyseBinary(node *ast.Binary) *Binary {
	left, right := self.analyseExpr(node.Left), self.analyseExpr(node.Right)
	lt, rt := left.GetType(), right.GetType()

	var kind BinaryType
	switch node.Opera.Kind {
	case token.AND:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			kind = BinaryAnd
		}
	case token.OR:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			kind = BinaryOr
		}
	case token.XOR:
		if lt.Equal(rt) && TypeIs[IntType](lt) {
			kind = BinaryXor
		}
	case token.ADD:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			kind = BinaryAdd
		}
	case token.SUB:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			kind = BinarySub
		}
	case token.MUL:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			kind = BinaryMul
		}
	case token.DIV:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			kind = BinaryDiv
		}
	case token.REM:
		if lt.Equal(rt) && TypeIs[NumberType](lt) {
			kind = BinaryRem
		}
	default:
		panic("unreachable")
	}

	if kind == BinaryInvalid {
		// TODO: 编译时异常：不能对两个类型进行二元运算
		panic("unreachable")
	}
	return &Binary{
		Kind:  kind,
		Left:  left,
		Right: right,
	}
}

func (self *Analyser) analyseUnary(node *ast.Unary) *Unary {
	value := self.analyseExpr(node.Value)
	vt := value.GetType()

	var kind UnaryType
	switch node.Opera.Kind {
	case token.SUB:
		if TypeIs[NumberType](vt) {
			kind = UnaryNegate
		}
	default:
		panic("unreachable")
	}

	if kind == UnaryInvalid {
		// TODO: 编译时异常：不能对两个类型进行一元运算
		panic("unreachable")
	}
	return &Unary{
		Kind:  kind,
		Value: value,
	}
}
