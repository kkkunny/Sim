package parse

import (
	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/slices"

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
	case token.IDENT:
		ident := (*ast.IdentExpr)(self.parseIdent())
		if canStruct && self.nextIs(token.LBR) {
			return util.Some[ast.Expr](self.parseStruct((*ast.IdentType)(ident)))
		} else {
			return util.Some[ast.Expr](ident)
		}
	case token.LPA:
		return util.Some(self.parseTupleOrLambda())
	case token.LBA:
		return util.Some[ast.Expr](self.parseArray())
	case token.STRING:
		return util.Some[ast.Expr](self.parseString())
	case token.SELF:
		st := &ast.SelfType{Token: self.expectNextIs(token.SELF)}
		return util.Some[ast.Expr](self.parseStruct(st))
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

func (self *Parser) parseTupleOrLambda() ast.Expr {
	begin := self.expectNextIs(token.LPA).Position
	first := true
	var isLambda bool
	elems := loopParseWithUtil(self, token.COM, token.RPA, func() any {
		if first {
			first = false

			var mut util.Option[token.Token]
			if self.skipNextIs(token.MUT) {
				isLambda, mut = true, util.Some(self.curTok)
			}
			firstParam := self.mustExpr(self.parseOptionExpr(true))
			firstIdent, ok := firstParam.(*ast.IdentExpr)
			if ok && firstIdent.Pkg.IsNone() && self.skipNextIs(token.COL) {
				isLambda = true
			}

			if isLambda {
				pt := self.parseType()
				return ast.Param{
					Mutable: mut,
					Name:    util.Some(firstIdent.Name),
					Type:    pt,
				}
			} else {
				return firstParam
			}
		} else if isLambda {
			return self.parseParam()
		} else {
			return self.mustExpr(self.parseOptionExpr(true))
		}
	})
	end := self.expectNextIs(token.RPA).Position
	if len(elems) == 0 && self.nextIs(token.ARROW) {
		isLambda = true
	}

	if !isLambda {
		return &ast.Tuple{
			Begin: begin,
			Elems: stlslices.Map(elems, func(_ int, e any) ast.Expr {
				return e.(ast.Expr)
			}),
			End: end,
		}
	}

	self.expectNextIs(token.ARROW)
	ret := self.parseType()
	body := self.parseBlock()
	return &ast.Lambda{
		Begin: begin,
		Params: stlslices.Map(elems, func(_ int, e any) ast.Param {
			return e.(ast.Param)
		}),
		Ret:  ret,
		Body: body,
	}
}

func (self *Parser) parseOptionPrefixUnary(canStruct bool) util.Option[ast.Expr] {
	switch self.nextTok.Kind {
	case token.SUB, token.NOT, token.AND, token.MUL:
		self.next()
		opera := self.curTok
		if self.curTok.Is(token.AND) && self.skipNextIs(token.MUT) {
			opera.Kind = token.AND_WITH_MUT
		}
		value := self.mustExpr(self.parseOptionUnary(canStruct))
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
			front = util.Some[ast.Expr](&ast.Extract{
				From:  fv,
				Index: self.curTok,
			})
		} else {
			front = util.Some[ast.Expr](&ast.Dot{
				From:  fv,
				Index: self.expectNextIs(token.IDENT),
			})
		}
	case token.NOT:
		front = util.Some[ast.Expr](&ast.CheckNull{
			Value: fv,
			End:   self.expectNextIs(token.NOT).Position,
		})
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
	return self.parseOptionSuffixUnary(self.parseOptionPrefixUnary(canStruct), canStruct)
}

func (self *Parser) parseOptionUnaryWithTail(canStruct bool) util.Option[ast.Expr] {
	return self.parseOptionTailUnary(self.parseOptionUnary(canStruct))
}

func (self *Parser) parseOptionBinary(priority uint8, canStruct bool) util.Option[ast.Expr] {
	left, ok := self.parseOptionUnaryWithTail(canStruct).Value()
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

func (self *Parser) parseArray() *ast.Array {
	begin := self.expectNextIs(token.LBA).Position
	elems := self.parseExprList(token.RBA)
	end := self.expectNextIs(token.RBA).Position
	return &ast.Array{
		Begin: begin,
		Elems: elems,
		End:   end,
	}
}

func (self *Parser) parseStruct(st ast.Type) *ast.Struct {
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
