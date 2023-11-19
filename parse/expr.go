package parse

import (
	"github.com/kkkunny/stl/container/pair"

	. "github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) mustExpr(op util.Option[Expr]) Expr {
	v, ok := op.Value()
	if !ok {
		errors.ThrowIllegalExpression(self.nextTok.Position)
	}
	return v
}

func (self *Parser) parseOptionExpr(canStruct bool) util.Option[Expr] {
	return self.parseOptionBinary(0, canStruct)
}

func (self *Parser) parseOptionPrimary(canStruct bool) util.Option[Expr] {
	switch self.nextTok.Kind {
	case token.INTEGER:
		return util.Some[Expr](self.parseInteger())
	case token.CHAR:
		return util.Some[Expr](self.parseChar())
	case token.FLOAT:
		return util.Some[Expr](self.parseFloat())
	case token.TRUE, token.FALSE:
		return util.Some[Expr](self.parseBool())
	case token.IDENT:
		ident := self.parseIdent()
		if !canStruct || !self.nextIs(token.LBR) {
			return util.Some[Expr](ident)
		}
		return util.Some[Expr](self.parseStruct(&IdentType{
			Pkg:  ident.Pkg,
			Name: ident.Name,
		}))
	case token.LPA:
		return util.Some[Expr](self.parseTuple())
	case token.LBA:
		return util.Some[Expr](self.parseArray())
	case token.STRING:
		return util.Some[Expr](self.parseString())
	default:
		return util.None[Expr]()
	}
}

func (self *Parser) parseInteger() *Integer {
	return &Integer{Value: self.expectNextIs(token.INTEGER)}
}

func (self *Parser) parseChar() *Char {
	return &Char{Value: self.expectNextIs(token.CHAR)}
}

func (self *Parser) parseString() *String {
	return &String{Value: self.expectNextIs(token.STRING)}
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
	elems := self.parseExprList(token.RPA)
	end := self.expectNextIs(token.RPA).Position
	return &Tuple{
		Begin: begin,
		Elems: elems,
		End:   end,
	}
}

func (self *Parser) parseOptionPrefixUnary(canStruct bool) util.Option[Expr] {
	switch self.nextTok.Kind {
	case token.SUB, token.NOT:
		self.next()
		opera := self.curTok
		value := self.mustExpr(self.parseOptionPrefixUnary(canStruct))
		return util.Some[Expr](&Unary{
			Opera: opera,
			Value: value,
		})
	default:
		return self.parseOptionPrimary(canStruct)
	}
}

func (self *Parser) parseOptionSuffixUnary(front util.Option[Expr], canStruct bool) util.Option[Expr] {
	fv, ok := front.Value()
	if !ok {
		return front
	}

	switch self.nextTok.Kind {
	case token.LPA:
		self.expectNextIs(token.LPA)
		args := self.parseExprList(token.RPA)
		end := self.expectNextIs(token.RPA).Position
		front = util.Some[Expr](&Call{
			Func: fv,
			Args: args,
			End:  end,
		})
	case token.LBA:
		self.expectNextIs(token.LBA)
		index := self.mustExpr(self.parseOptionExpr(canStruct))
		self.expectNextIs(token.RBA)
		front = util.Some[Expr](&Index{
			From:  fv,
			Index: index,
		})
	case token.DOT:
		self.expectNextIs(token.DOT)
		if self.skipNextIs(token.INTEGER) {
			index := self.curTok
			front = util.Some[Expr](&Extract{
				From:  fv,
				Index: index,
			})
		} else {
			index := self.expectNextIs(token.IDENT)
			front = util.Some[Expr](&Field{
				From:  fv,
				Index: index,
			})
		}
	default:
		return front
	}

	return self.parseOptionSuffixUnary(front, canStruct)
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
	case token.IS:
		self.expectNextIs(token.IS)
		t := self.parseType()
		front = util.Some[Expr](&Judgment{
			Value: fv,
			Type:  t,
		})
	default:
		return front
	}

	return self.parseOptionTailUnary(front)
}

func (self *Parser) parseOptionUnary(canStruct bool) util.Option[Expr] {
	return self.parseOptionTailUnary(self.parseOptionSuffixUnary(self.parseOptionPrefixUnary(canStruct), canStruct))
}

func (self *Parser) parseOptionBinary(priority uint8, canStruct bool) util.Option[Expr] {
	left, ok := self.parseOptionUnary(canStruct).Value()
	if !ok {
		return util.None[Expr]()
	}
	for {
		nextOp := self.nextTok
		if priority >= nextOp.Kind.Priority() {
			break
		}
		self.next()

		right := self.mustExpr(self.parseOptionBinary(nextOp.Kind.Priority(), canStruct))
		left = &Binary{
			Left:  left,
			Opera: nextOp,
			Right: right,
		}
	}
	return util.Some(left)
}

func (self *Parser) parseIdent() *Ident {
	pkg := util.None[token.Token]()
	var name token.Token
	pkgOrName := self.expectNextIs(token.IDENT)
	if self.skipNextIs(token.SCOPE) {
		pkg = util.Some(pkgOrName)
		name = self.expectNextIs(token.IDENT)
	} else {
		name = pkgOrName
	}
	return &Ident{
		Pkg:  pkg,
		Name: name,
	}
}

func (self *Parser) parseArray() *Array {
	at := self.parseArrayType()
	self.expectNextIs(token.LBR)
	elems := self.parseExprList(token.RBR)
	end := self.expectNextIs(token.RBR).Position
	return &Array{
		Type:  at,
		Elems: elems,
		End:   end,
	}
}

func (self *Parser) parseExprList(end token.Kind) (res []Expr) {
	return loopParseWithUtil(self, token.COM, end, func() Expr {
		return self.mustExpr(self.parseOptionExpr(true))
	})
}

func (self *Parser) parseStruct(st *IdentType) *Struct {
	self.expectNextIs(token.LBR)
	fields := loopParseWithUtil(self, token.COM, token.RBR, func() pair.Pair[token.Token, Expr] {
		fn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		fv := self.mustExpr(self.parseOptionExpr(true))
		return pair.NewPair(fn, fv)
	})
	end := self.expectNextIs(token.RBR).Position
	return &Struct{
		Type:   st,
		Fields: fields,
		End:    end,
	}
}
