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
		prec := p.cur.Kind.Priority()
		if prec < minPrec {
			break
		}
		op := p.cur
		p.next()
		p.skip(token.KindEnum.Br)
		right := p.parseBinaryExpr(prec + 1)
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}

	return left
}

func (p *Parser) parseUnaryExpr() ast.Expr {
	switch p.cur.Kind {
	case token.KindEnum.Not:
		op := p.expect(token.KindEnum.Not)
		p.skip(token.KindEnum.Br)
		expr := p.parseUnaryExpr()
		return &ast.UnaryExpr{Op: op, Expr: expr}
	default:
		return p.parsePrimaryExpr()
	}
}

func (p *Parser) parsePrimaryExpr() ast.Expr {
	switch p.cur.Kind {
	case token.KindEnum.Ident:
		name := p.expect(token.KindEnum.Ident)
		return &ast.IdentExpr{Name: name}
	case token.KindEnum.Integer:
		value := p.expect(token.KindEnum.Integer)
		return &ast.IntegerExpr{Value: value}
	case token.KindEnum.Lpa:
		begin := p.expect(token.KindEnum.Lpa).Position
		p.skip(token.KindEnum.Br)
		expr := p.parseExpr()
		if identExpr, ok := expr.(*ast.IdentExpr); ok && p.cur.Kind == token.KindEnum.Col {
			p.expect(token.KindEnum.Col)
			typ := p.parseType()
			params := []*ast.ParamDecl{{Name: identExpr.Name, Type: typ}}
			if _, ok = p.IfSkip(token.KindEnum.Comma); ok {
				p.skip(token.KindEnum.Br)
				for {
					if p.cur.Kind != token.KindEnum.Ident {
						break
					}
					params = append(params, p.parseParamDecl())
					if _, ok = p.IfSkip(token.KindEnum.Comma); !ok {
						break
					}
					p.skip(token.KindEnum.Br)
				}
			}
			end := p.expect(token.KindEnum.Rpa).Position

			var returnType optional.Optional[ast.Type]
			if _, ok := p.IfSkip(token.KindEnum.Arrow); ok {
				returnType = optional.Some(p.parseType())
				end = returnType.MustValue().Position()
			}

			var body optional.Optional[*ast.Block]
			if p.cur.Kind == token.KindEnum.Lbr {
				body = optional.Some(p.parseBlock())
				end = body.MustValue().Position()
			}

			return &ast.FuncExpr{BeginPosition: begin, Params: params, ReturnType: returnType, Body: body, EndPosition: end}
		}
		p.expect(token.KindEnum.Rpa)
		return expr
	default:
		p.reporter.Fatalf(
			p.cur.Position,
			report.Errors.UnexpectedToken,
			p.cur.Kind,
		)
		return nil
	}
}
