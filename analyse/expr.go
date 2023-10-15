package analyse

import (
	"math/big"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
)

func (self *Analyser) analyseExpr(node ast.Expr) Expr {
	switch exprNode := node.(type) {
	case *ast.Integer:
		return self.analyseInteger(exprNode)
	case *ast.Float:
		return self.analyseFloat(exprNode)
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
