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
		return util.Some[Expr](self.parseTuple())
	case token.LBA:
		return util.Some[Expr](self.parseArray())
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

func (self *Parser) parseTuple() *Tuple {
	begin := self.expectNextIs(token.LPA).Position
	var elems []Expr
	for self.skipSEM(); !self.nextIs(token.RPA); self.skipSEM() {
		elems = append(elems, self.mustExpr(self.parseOptionExpr()))
		if !self.skipNextIs(token.COM) {
			break
		}
	}
	end := self.expectNextIs(token.RPA).Position
	return &Tuple{
		Begin: begin,
		Elems: elems,
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
	case token.LBA:
		self.expectNextIs(token.LBA)
		index := self.mustExpr(self.parseOptionExpr())
		self.expectNextIs(token.RBA)
		front = util.Some[Expr](&Index{
			From:  fv,
			Index: index,
		})
	case token.DOT:
		self.expectNextIs(token.DOT)
		index := self.expectNextIs(token.INTEGER)
		front = util.Some[Expr](&Extract{
			From:  fv,
			Index: index,
		})
	default:
		return front
	}

	return self.parseOptionSuffixUnary(front)
}

func (self *Parser) parseOptionTailUnary(front util.Option[Expr]) util.Option[Expr] {
	fv, ok := front.Value()
	if !ok {
		return front
	}

	switch self.nextTok.Kind {
	case token.AS:
		self.expectNextIs(token.AS)
		t := self.parseType()
		front = util.Some[Expr](&Covert{
			Value: fv,
			Type:  t,
		})
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

func (self *Parser) parseArray() *Array {
	at := self.parseArrayType()
	self.expectNextIs(token.LBR)
	var elems []Expr
	for self.skipSEM(); !self.nextIs(token.RBR); self.skipSEM() {
		elems = append(elems, self.mustExpr(self.parseOptionExpr()))
		if !self.skipNextIs(token.COM) {
			break
		}
	}
	end := self.expectNextIs(token.RBR).Position
	return &Array{
		Type:  at,
		Elems: elems,
		End:   end,
	}
}
