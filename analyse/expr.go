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
	case *ast.Ident:
		return self.analyseIdent(exprNode)
	case *ast.Call:
		return self.analyseCall(exprNode)
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

func (self *Analyser) analyseBinary(node *ast.Binary) Binary {
	left, right := self.analyseExpr(node.Left), self.analyseExpr(node.Right)
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

func (self *Analyser) analyseUnary(node *ast.Unary) Unary {
	value := self.analyseExpr(node.Value)
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

func (self *Analyser) analyseIdent(node *ast.Ident) Ident {
	value, ok := self.localScope.GetValue(node.Name.Source())
	if !ok {
		// TODO: 编译时异常：未知的变量
		panic("unreachable")
	}
	return value
}

func (self *Analyser) analyseCall(node *ast.Call) *Call {
	f := self.analyseExpr(node.Func)
	return &Call{Func: f}
}
