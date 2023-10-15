package analyse

import (
	"math/big"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
)

func (self *Analyser) analyseExpr(node ast.Expr) Expr {
	switch exprNode := node.(type) {
	case *ast.Integer:
		return self.analyseInteger(exprNode)
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
