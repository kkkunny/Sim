package parse

import (
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	"github.com/kkkunny/stl/container/tuple"

	"github.com/kkkunny/Sim/compiler/ast"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
)

func (self *Parser) mustExpr(op optional.Optional[ast.Expr]) ast.Expr {
	v, ok := op.Value()
	if !ok {
		errors.ThrowIllegalExpression(self.nextTok.Position)
	}
	return v
}

func (self *Parser) parseOptionExpr(canStruct bool) optional.Optional[ast.Expr] {
	return self.parseOptionBinary(0, canStruct)
}

func (self *Parser) parseOptionPrimary(canStruct bool) optional.Optional[ast.Expr] {
	switch self.nextTok.Kind {
	case token.INTEGER:
		return optional.Some[ast.Expr](self.parseInteger())
	case token.CHAR:
		return optional.Some[ast.Expr](self.parseChar())
	case token.FLOAT:
		return optional.Some[ast.Expr](self.parseFloat())
	case token.IDENT:
		ident := (*ast.IdentExpr)(self.parseIdent())
		if canStruct && self.nextIs(token.LBR) {
			return optional.Some[ast.Expr](self.parseStruct((*ast.IdentType)(ident)))
		} else {
			return optional.Some[ast.Expr](ident)
		}
	case token.LPA:
		return optional.Some(self.parseTupleOrLambda())
	case token.LBA:
		return optional.Some[ast.Expr](self.parseArray())
	case token.STRING:
		return optional.Some[ast.Expr](self.parseString())
	case token.SELF:
		st := &ast.SelfType{Token: self.expectNextIs(token.SELF)}
		return optional.Some[ast.Expr](self.parseStruct(st))
	default:
		return optional.None[ast.Expr]()
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

			var mut optional.Optional[token.Token]
			if self.skipNextIs(token.MUT) {
				isLambda, mut = true, optional.Some(self.curTok)
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
					Name:    optional.Some(firstIdent.Name),
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

func (self *Parser) parseOptionPrefixUnary(canStruct bool) optional.Optional[ast.Expr] {
	switch self.nextTok.Kind {
	case token.SUB, token.NOT, token.AND, token.MUL:
		self.next()
		opera := self.curTok
		if self.curTok.Is(token.AND) && self.skipNextIs(token.MUT) {
			opera.Kind = token.AND_WITH_MUT
		}
		value := self.mustExpr(self.parseOptionUnary(canStruct))
		return optional.Some[ast.Expr](&ast.Unary{
			Opera: opera,
			Value: value,
		})
	default:
		return self.parseOptionPrimary(canStruct)
	}
}

func (self *Parser) parseOptionSuffixUnary(front optional.Optional[ast.Expr], canStruct bool) optional.Optional[ast.Expr] {
	fv, ok := front.Value()
	if !ok {
		return front
	}

	switch self.nextTok.Kind {
	case token.LPA:
		self.expectNextIs(token.LPA)
		args := self.parseExprList(token.RPA)
		end := self.expectNextIs(token.RPA).Position
		front = optional.Some[ast.Expr](&ast.Call{
			Func: fv,
			Args: args,
			End:  end,
		})
	case token.LBA:
		self.expectNextIs(token.LBA)
		index := self.mustExpr(self.parseOptionExpr(canStruct))
		self.expectNextIs(token.RBA)
		front = optional.Some[ast.Expr](&ast.Index{
			From:  fv,
			Index: index,
		})
	case token.DOT:
		self.expectNextIs(token.DOT)
		if self.skipNextIs(token.INTEGER) {
			front = optional.Some[ast.Expr](&ast.Extract{
				From:  fv,
				Index: self.curTok,
			})
		} else {
			front = optional.Some[ast.Expr](&ast.Dot{
				From:  fv,
				Index: self.expectNextIs(token.IDENT),
			})
		}
	case token.NOT:
		front = optional.Some[ast.Expr](&ast.CheckNull{
			Value: fv,
			End:   self.expectNextIs(token.NOT).Position,
		})
	default:
		return front
	}

	return self.parseOptionSuffixUnary(front, canStruct)
}

func (self *Parser) parseOptionTailUnary(front optional.Optional[ast.Expr]) optional.Optional[ast.Expr] {
	fv, ok := front.Value()
	if !ok {
		return front
	}

	switch self.nextTok.Kind {
	case token.AS:
		self.expectNextIs(token.AS)
		t := self.parseType()
		front = optional.Some[ast.Expr](&ast.Covert{
			Value: fv,
			Type:  t,
		})
	case token.IS:
		self.expectNextIs(token.IS)
		t := self.parseType()
		front = optional.Some[ast.Expr](&ast.Judgment{
			Value: fv,
			Type:  t,
		})
	default:
		return front
	}

	return self.parseOptionTailUnary(front)
}

func (self *Parser) parseOptionUnary(canStruct bool) optional.Optional[ast.Expr] {
	return self.parseOptionSuffixUnary(self.parseOptionPrefixUnary(canStruct), canStruct)
}

func (self *Parser) parseOptionUnaryWithTail(canStruct bool) optional.Optional[ast.Expr] {
	return self.parseOptionTailUnary(self.parseOptionUnary(canStruct))
}

func (self *Parser) parseOptionBinary(priority uint8, canStruct bool) optional.Optional[ast.Expr] {
	left, ok := self.parseOptionUnaryWithTail(canStruct).Value()
	if !ok {
		return optional.None[ast.Expr]()
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
	return optional.Some[ast.Expr](left)
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
	fields := loopParseWithUtil(self, token.COM, token.RBR, func() tuple.Tuple2[token.Token, ast.Expr] {
		fn := self.expectNextIs(token.IDENT)
		self.expectNextIs(token.COL)
		fv := self.mustExpr(self.parseOptionExpr(true))
		return tuple.Pack2(fn, fv)
	})
	end := self.expectNextIs(token.RBR).Position
	return &ast.Struct{
		Type:   st,
		Fields: fields,
		End:    end,
	}
}
