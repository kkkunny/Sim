package parse

import (
	. "github.com/kkkunny/Sim/ast"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) mustExpr(op util.Option[Expr]) Expr {
	v, ok := op.Value()
	if !ok {
		// TODO: 编译时异常：期待一个表达式
		panic("unreachable")
	}
	return v
}

func (self *Parser) parseOptionExpr() util.Option[Expr] {
	return self.parseOptionBinary(0)
}

func (self *Parser) parseOptionPrimary() util.Option[Expr] {
	switch self.nextTok.Kind {
	case token.INTEGER:
		return util.Some[Expr](self.parseInteger())
	case token.FLOAT:
		return util.Some[Expr](self.parseFloat())
	case token.TRUE, token.FALSE:
		return util.Some[Expr](self.parseBool())
	case token.IDENT:
		return util.Some[Expr](self.parseIdent())
	case token.LPA:
		return util.Some[Expr](self.parseUnit())
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

func (self *Parser) parseBool() *Boolean {
	var value token.Token
	if self.skipNextIs(token.TRUE) {
		value = self.curTok
	} else {
		value = self.expectNextIs(token.FALSE)
	}
	return &Boolean{Value: value}
}

func (self *Parser) parseUnit() *Unit {
	begin := self.expectNextIs(token.LPA).Position
	value := self.mustExpr(self.parseOptionExpr())
	end := self.expectNextIs(token.RPA).Position
	return &Unit{
		Begin: begin,
		Value: value,
		End:   end,
	}
}

func (self *Parser) parseOptionPrefixUnary() util.Option[Expr] {
	switch self.nextTok.Kind {
	case token.SUB:
		opera := self.expectNextIs(token.SUB)
		value := self.mustExpr(self.parseOptionPrefixUnary())
		return util.Some[Expr](&Unary{
			Opera: opera,
			Value: value,
		})
	default:
		return self.parseOptionPrimary()
	}
}

func (self *Parser) parseOptionSuffixUnary(front util.Option[Expr]) util.Option[Expr] {
	fv, ok := front.Value()
	if !ok {
		return front
	}

	switch self.nextTok.Kind {
	case token.LPA:
		self.expectNextIs(token.LPA)
		self.expectNextIs(token.RPA)
		front = util.Some[Expr](&Call{Func: fv})
	default:
		return front
	}

	return self.parseOptionSuffixUnary(front)
}

func (self *Parser) parseOptionTailUnary(front util.Option[Expr]) util.Option[Expr] {
	_, ok := front.Value()
	if !ok {
		return front
	}

	switch self.nextTok.Kind {
	default:
		return front
	}

	return self.parseOptionTailUnary(front)
}

func (self *Parser) parseOptionUnary() util.Option[Expr] {
	return self.parseOptionTailUnary(self.parseOptionSuffixUnary(self.parseOptionPrefixUnary()))
}

func (self *Parser) parseOptionBinary(priority uint8) util.Option[Expr] {
	left, ok := self.parseOptionUnary().Value()
	if !ok {
		return util.None[Expr]()
	}
	for {
		nextOp := self.nextTok
		if priority >= nextOp.Kind.Priority() {
			break
		}
		self.next()

		right := self.mustExpr(self.parseOptionBinary(nextOp.Kind.Priority()))
		left = &Binary{
			Left:  left,
			Opera: nextOp,
			Right: right,
		}
	}
	return util.Some(left)
}

func (self *Parser) parseIdent() *Ident {
	return &Ident{Name: self.expectNextIs(token.IDENT)}
}
