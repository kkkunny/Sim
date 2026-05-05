package parse

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseExpr() ast.Expr {
	return p.parseBinaryExpr(0)
}

func (p *Parser) parseBinaryExpr(minPrec int) ast.Expr {
	left := p.parseUnaryExpr()

	for {
		prec := p.nextToken.Kind.Priority()
		if prec < minPrec {
			break
		}
		op := p.nextToken
		p.next()
		p.skip(token.KindEnum.Br)
		right := p.parseBinaryExpr(prec + 1)
		left = &ast.Binary{Op: op, Left: left, Right: right}
	}

	return left
}

func (p *Parser) parseUnaryExpr() ast.Expr {
	switch p.nextToken.Kind {
	case token.KindEnum.Not, token.KindEnum.Mul:
		p.next()
		op := p.curToken
		p.skip(token.KindEnum.Br)
		expr := p.parseUnaryExpr()
		return &ast.Unary{Op: op, Expr: expr}
	case token.KindEnum.And:
		begin := p.expect(token.KindEnum.And).Position
		mut := p.ifSkip(token.KindEnum.Mut)
		p.skip(token.KindEnum.Br)
		expr := p.parseUnaryExpr()
		return &ast.GetReference{
			BeginPosition: begin,
			Mut:           mut,
			Value:         expr,
		}
	default:
		return p.parseSuffixExpr(p.parsePrimaryExpr())
	}
}

func (p *Parser) parsePrimaryExpr() ast.Expr {
	switch p.nextToken.Kind {
	case token.KindEnum.Ident:
		name := p.expect(token.KindEnum.Ident)
		var pkg optional.Optional[token.Token]
		if p.ifSkip(token.KindEnum.Scope) {
			pkg = optional.Some(name)
			name = p.expect(token.KindEnum.Ident)
		}
		return &ast.IdentExpr{Pkg: pkg, Name: name}
	case token.KindEnum.Integer:
		value := p.expect(token.KindEnum.Integer)
		return &ast.Integer{Value: value}
	case token.KindEnum.Char:
		value := p.expect(token.KindEnum.Char)
		return &ast.Char{Value: value}
	case token.KindEnum.String:
		value := p.expect(token.KindEnum.String)
		return &ast.String{Value: value}
	case token.KindEnum.Lpa:
		begin := p.expect(token.KindEnum.Lpa).Position
		p.skip(token.KindEnum.Br)
		if p.ifSkip(token.KindEnum.Rpa) {
			// 空元组
			if p.nextToken.Kind != token.KindEnum.Arrow && p.nextToken.Kind != token.KindEnum.Lbr {
				return &ast.Tuple{
					BeginPosition: begin,
					EndPosition:   p.curToken.Position,
				}
			}

			// 无参函数
			var returnType optional.Optional[ast.Type]
			if p.ifSkip(token.KindEnum.Arrow) {
				returnType = optional.Some(p.parseType())
			}

			var body optional.Optional[*ast.Block]
			if p.nextToken.Kind == token.KindEnum.Lbr {
				body = optional.Some(p.parseBlock())
			}

			return &ast.Func{BeginPosition: begin, ReturnType: returnType, Body: body, EndPosition: p.curToken.Position}
		}

		var params []*ast.ParamDecl
		var expr ast.Expr
		if p.ifSkip(token.KindEnum.Mut) {
			// 有mut, 说明在定义形参，走函数表达式逻辑
			pn := p.expect(token.KindEnum.Ident)
			p.expect(token.KindEnum.Col)
			pt := p.parseType()
			params = append(params, &ast.ParamDecl{Mut: true, Name: pn, Type: pt})
		} else {
			expr = p.parseExpr()
			if identExpr, ok := expr.(*ast.IdentExpr); ok && identExpr.Pkg.IsNone() && p.nextToken.Kind == token.KindEnum.Col {
				p.expect(token.KindEnum.Col)
				pt := p.parseType()
				params = append(params, &ast.ParamDecl{Name: identExpr.Name, Type: pt})
			}
		}

		if len(params) > 0 {
			if p.ifSkip(token.KindEnum.Comma) {
				p.skip(token.KindEnum.Br)
				for p.nextToken.Kind == token.KindEnum.Ident {
					params = append(params, p.parseParamDecl())
					if !p.ifSkip(token.KindEnum.Comma) {
						break
					}
					p.skip(token.KindEnum.Br)
				}
			}
			p.expect(token.KindEnum.Rpa)

			var returnType optional.Optional[ast.Type]
			if p.ifSkip(token.KindEnum.Arrow) {
				returnType = optional.Some(p.parseType())
			}

			var body optional.Optional[*ast.Block]
			if p.nextToken.Kind == token.KindEnum.Lbr {
				body = optional.Some(p.parseBlock())
			}

			return &ast.Func{BeginPosition: begin, Params: params, ReturnType: returnType, Body: body, EndPosition: p.curToken.Position}
		}

		// 非空元组
		elems := []ast.Expr{expr}
		p.ifSkip(token.KindEnum.Comma)
		p.skip(token.KindEnum.Br)
		for p.nextToken.Kind != token.KindEnum.Rpa {
			elems = append(elems, p.parseExpr())
			if !p.ifSkip(token.KindEnum.Comma) {
				break
			}
			p.skip(token.KindEnum.Br)
		}
		end := p.expect(token.KindEnum.Rpa).Position
		return &ast.Tuple{BeginPosition: begin, Elems: elems, EndPosition: end}
	case token.KindEnum.Lba:
		begin := p.expect(token.KindEnum.Lba).Position
		p.skip(token.KindEnum.Br)
		var elems []ast.Expr
		for p.nextToken.Kind != token.KindEnum.Rba {
			elems = append(elems, p.parseExpr())
			if !p.ifSkip(token.KindEnum.Comma) {
				break
			}
			p.skip(token.KindEnum.Br)
		}
		end := p.expect(token.KindEnum.Rba).Position
		return &ast.Array{BeginPosition: begin, Elems: elems, EndPosition: end}
	case token.KindEnum.True, token.KindEnum.False:
		p.next()
		return &ast.Boolean{Value: p.curToken}
	default:
		p.reporter.Fatalf(
			p.nextToken.Position,
			report.Errors.UnexpectedToken,
			p.nextToken.Kind,
		)
		return nil
	}
}

func (p *Parser) parseSuffixExpr(prev ast.Expr) (expr ast.Expr) {
	for {
		switch p.nextToken.Kind {
		case token.KindEnum.Lpa:
			p.expect(token.KindEnum.Lpa)
			p.skip(token.KindEnum.Br)
			var args []ast.Expr
			for p.nextToken.Kind != token.KindEnum.Rpa {
				args = append(args, p.parseExpr())
				if !p.ifSkip(token.KindEnum.Comma) {
					break
				}
				p.skip(token.KindEnum.Br)
			}
			end := p.expect(token.KindEnum.Rpa).Position
			prev = &ast.Call{
				Func:        prev,
				Args:        args,
				EndPosition: end,
			}
		case token.KindEnum.Lba:
			p.expect(token.KindEnum.Lba)
			p.skip(token.KindEnum.Br)
			index := p.parseExpr()
			end := p.expect(token.KindEnum.Rba).Position
			prev = &ast.Index{
				From:        prev,
				Index:       index,
				EndPosition: end,
			}
		case token.KindEnum.As:
			p.expect(token.KindEnum.As)
			t := p.parseType()
			prev = &ast.As{
				Left:  prev,
				Right: t,
			}
		case token.KindEnum.Question:
			p.expect(token.KindEnum.Question)
			p.skip(token.KindEnum.Br)
			trueExpr := p.parseExpr()
			p.expect(token.KindEnum.Col)
			p.skip(token.KindEnum.Br)
			falseExpr := p.parseExpr()
			prev = &ast.Ternary{Condition: prev, TrueExpr: trueExpr, FalseExpr: falseExpr}
		default:
			return prev
		}
	}
}
