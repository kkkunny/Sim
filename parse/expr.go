package parse

import (
	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) parseOptionExpr() util.Option[Expr] {
	switch self.nextTok.Kind {
	case token.INTEGER:
		return util.Some[Expr](self.parseInteger())
	default:
		return util.None[Expr]()
	}
}

func (self *Parser) parseExpr() Expr {
	e, ok := self.parseOptionExpr().Value()
	if !ok {
		// TODO: 编译时异常：期待一个表达式
		panic("unreachable")
	}
	return e
}

func (self *Parser) parseInteger() *Integer {
	return &Integer{Value: self.expectNextIs(token.INTEGER)}
}
