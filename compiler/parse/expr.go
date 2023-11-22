package parse

import (
	"github.com/kkkunny/stl/container/pair"

	"github.com/kkkunny/Sim/ast"

	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/token"
	"github.com/kkkunny/Sim/util"
)

func (self *Parser) mustExpr(op util.Option[ast.Expr]) ast.Expr {
	v, ok := op.Value()
	if !ok {
		errors.ThrowIllegalExpression(self.nextTok.Position)
	}
	return v
}

func (self *Parser) parseOptionExpr(canStruct bool) util.Option[ast.Expr] {
	return self.parseOptionBinary(0, canStruct)
}

func (self *Parser) parseOptionPrimary(canStruct bool) util.Option[ast.Expr] {
	switch self.nextTok.Kind {
	case token.INTEGER:
		return util.Some[ast.Expr](self.parseInteger())
	case token.CHAR:
		return util.Some[ast.Expr](self.parseChar())
	case token.FLOAT:
		return util.Some[ast.Expr](self.parseFloat())
	case token.TRUE, token.FALSE:
		return util.Some[ast.Expr](self.parseBool())
	case token.IDENT:
		ident := self.parseIdent()
		if !canStruct || !self.nextIs(token.LBR) {
			return util.Some[ast.Expr](ident)
		}
		return util.Some[ast.Expr](self.parseStruct(&ast.IdentType{
			Pkg:  ident.Pkg,
			Name: ident.Name,
		}))
	case token.LPA:
		return util.Some[ast.Expr](self.parseTuple())
	case token.LBA:
		return util.Some[ast.Expr](self.parseArray())
	case token.STRING:
		return util.Some[ast.Expr](self.parseString())
	default:
		return util.None[ast.Expr]()
	}
}

func (self *Parser) parseInteger() *ast.Integer {
	return &ast.Integer{Value: self.expectNextIs(token.INTEGER)}
}

func (self *Parser) parseChar() *ast.Char {
	return &ast.Char{Value: self.expectNextIs(token.CHAR)}
}

func (self *Parser) parseString() *ast.String {
	return &ast.String{Value: self.expectNextIs(token.STRING)}
}

func (self *Parser) parseFloat() *ast.Float {
	return &ast.Float{Value: self.expectNextIs(token.FLOAT)}
}

func (self *Parser) parseBool() *ast.Boolean {
	var value token.Token
	if self.skipNextIs(token.TRUE) {
		value = self.curTok
	} else {
		value = self.expectNextIs(token.FALSE)
	}
	return &ast.Boolean{Value: value}
}

func (self *Parser) parseTuple() *ast.Tuple {
	begin := self.expectNextIs(token.LPA).Position
	elems := self.parseExprList(token.RPA)
	end := self.expectNextIs(token.RPA).Position
	return &ast.Tuple{
		Begin: begin,
		Elems: elems,
		End:   end,
	}
}

func (self *Parser) parseOptionPrefixUnary(canStruct bool) util.Option[ast.Expr] {
	switch self.nextTok.Kind {
	case token.SUB, token.NOT, token.AND, token.MUL:
		self.next()
		opera := self.curTok
		value := self.mustExpr(self.parseOptionPrefixUnary(canStruct))
		return util.Some[ast.Expr](&ast.Unary{
			Opera: opera,
			Value: value,
		})
	default:
		return self.parseOptionPrimary(canStruct)
	}
}

func (self *Parser) parseOptionSuffixUnary(front util.Option[ast.Expr], canStruct bool) util.Option[ast.Expr] {
	fv, ok := front.Value()
	if !ok {
		return front
	}

	switch self.nextTok.Kind {
	case token.LPA:
		self.expectNextIs(token.LPA)
		args := self.parseExprList(token.RPA)
		end := self.expectNextIs(token.RPA).Position
		front = util.Some[ast.Expr](&ast.Call{
			Func: fv,
			Args: args,
			End:  end,
		})
	case token.LBA:
		self.expectNextIs(token.LBA)
		index := self.mustExpr(self.parseOptionExpr(canStruct))
		self.expectNextIs(token.RBA)
		front = util.Some[ast.Expr](&ast.Index{
			From:  fv,
			Index: index,
		})
	case token.DOT:
		self.expectNextIs(token.DOT)
		if self.skipNextIs(token.INTEGER) {
			index := self.curTok
			front = util.Some[ast.Expr](&ast.Extract{
				From:  fv,
				Index: index,
			})
		} else {
			index := self.expectNextIs(token.IDENT)
			front = util.Some[ast.Expr](&ast.Field{
				From:  fv,
				Index: index,
			})
		}
	default:
		return front
	}

	return self.parseOptionSuffixUnary(front, canStruct)
}

func (self *Parser) parseOptionTailUnary(front util.Option[ast.Expr]) util.Option[ast.Expr] {
	fv, ok := front.Value()
	if !ok {
		return front
	}

	switch self.nextTok.Kind {
	case token.AS:
		self.expectNextIs(token.AS)
		t := self.parseType()
		front = util.Some[ast.Expr](&ast.Covert{
			Value: fv,
			Type:  t,
		})
	case token.IS:
		self.expectNextIs(token.IS)
		t := self.parseType()
		front = util.Some[ast.Expr](&ast.Judgment{
			Value: fv,
			Type:  t,
		})
	default:
		return front
	}

	return self.parseOptionTailUnary(front)
}

func (self *Parser) parseOptionUnary(canStruct bool) util.Option[ast.Expr] {
	return self.parseOptionTailUnary(self.parseOptionSuffixUnary(self.parseOptionPrefixUnary(canStruct), canStruct))
}

func (self *Parser) parseOptionBinary(priority uint8, canStruct bool) util.Option[ast.Expr] {
	left, ok := self.parseOptionUnary(canStruct).Value()
	if !ok {
		return util.None[ast.Expr]()
	}
	for {
		nextOp := self.nextTok
		if priority >= nextOp.Kind.Priority() {
			break
		}
		self.next()

		right := self.mustExpr(self.parseOptionBinary(nextOp.Kind.Priority(), canStruct))
		left = &ast.Binary{
			Left:  left,
			Opera: nextOp,
			Right: right,
		}
	}
	return util.Some(left)
}

func (self *Parser) parseIdent() *ast.Ident {
	pkg := util.None[token.Token]()
	var name token.Token
	pkgOrName := self.expectNextIs(token.IDENT)
	if self.skipNextIs(token.SCOPE) {
		pkg = util.Some(pkgOrName)
		name = self.expectNextIs(token.IDENT)
	} else {
		name = pkgOrName
	}
	return &ast.Ident{
		Pkg:  pkg,
		Name: name,
	}
}

func (self *Parser) parseArray() *ast.Array {
	at := self.parseArrayType()
	self.expectNextIs(token.LBR)
	elems := self.parseExprList(token.RBR)
	end := self.expectNextIs(token.RBR).Position
	return &ast.Array{
		Type:  at,
		Elems: elems,
		End:   end,
	}
}

func (self *Parser) parseExprList(end token.Kind) (res []ast.Expr) {
	return loopParseWithUtil(self, token.COM, end, func() ast.Expr {
		return self.mustExpr(self.parseOptionExpr(true))
	})
}

func (self *Parser) parseStruct(st *ast.IdentType) *ast.Struct {
	self.expectNextIs(token.LBR)
	fields := loopParseWithUtil(self, token.COM, token.RBR, func() pair.Pair[token.Token, ast.Expr] {
		fn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		fv := self.mustExpr(self.parseOptionExpr(true))
		return pair.NewPair(fn, fv)
	})
	end := self.expectNextIs(token.RBR).Position
	return &ast.Struct{
		Type:   st,
		Fields: fields,
		End:    end,
	}
}
