package parse

import (
	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) parseOptionExpr() util.Option[Expr] {
	return self.parseOptionBinary(0)
}

func (self *Parser) parseExpr() Expr {
	return self.parseBinary(0)
}

func (self *Parser) parseOptionPrimary() util.Option[Expr] {
	switch self.nextTok.Kind {
	case token.INTEGER:
		return util.Some[Expr](self.parseInteger())
	case token.FLOAT:
		return util.Some[Expr](self.parseFloat())
	default:
		return util.None[Expr]()
	}
}

func (self *Parser) parseInteger() *Integer {
	return &Integer{Value: self.expectNextIs(token.INTEGER)}
}

func (self *Parser) parseFloat() *Float {
	return &Float{Value: self.expectNextIs(token.FLOAT)}
}

func (self *Parser) parseOptionBinary(priority uint8) util.Option[Expr] {
	left, ok := self.parseOptionPrimary().Value()
	if !ok {
		return util.None[Expr]()
	}
	for {
		nextOp := self.nextTok
		if priority >= nextOp.Kind.Priority() {
			break
		}
		self.next()

		right := self.parseBinary(nextOp.Kind.Priority())
		left = &Binary{
			Left:  left,
			Opera: nextOp,
			Right: right,
		}
	}
	return util.Some(left)
}

func (self *Parser) parseBinary(priority uint8) Expr {
	v, ok := self.parseOptionBinary(priority).Value()
	if !ok {
		// TODO: 编译时异常：期待一个表达式
		panic("unreachable")
	}
	return v
}
