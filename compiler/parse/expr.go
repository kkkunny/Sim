package parse

import (
	"fmt"

	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseExpr() ast.Expr {
	return p.parseBinaryExpr(0)
}

func (p *Parser) parseBinaryExpr(minPrec int) ast.Expr {
	left := p.parsePrimaryExpr()

	for {
		prec := p.cur.Kind.Priority()
		if prec < minPrec {
			break
		}
		op := p.cur
		p.next()
		right := p.parseBinaryExpr(prec + 1)
		left = &ast.BinaryExpr{Op: op, Left: left, Right: right}
	}

	return left
}

func (p *Parser) parsePrimaryExpr() ast.Expr {
	switch p.cur.Kind {
	case token.KindEnum.Not:
		op := p.expect(token.KindEnum.Not)
		expr := p.parsePrimaryExpr()
		return &ast.UnaryExpr{Op: op, Expr: expr}
	case token.KindEnum.Ident:
		return &ast.IdentExpr{Name: p.expect(token.KindEnum.Ident)}
	case token.KindEnum.Integer:
		return &ast.IntegerExpr{Value: p.expect(token.KindEnum.Integer)}
	case token.KindEnum.Lpa:
		p.expect(token.KindEnum.Lpa)
		expr := p.parseExpr()
		if identExpr, ok := expr.(*ast.IdentExpr); ok && p.cur.Kind == token.KindEnum.Col {
			p.expect(token.KindEnum.Col)
			typ := p.parseType()
			params := []*ast.ParamDecl{{Name: identExpr.Name, Type: typ}}
			for p.cur.Kind == token.KindEnum.Comma {
				p.next()
				params = append(params, p.parseParamDecl())
			}
			p.expect(token.KindEnum.Rpa)

			var returnType optional.Optional[ast.Type]
			if _, ok := p.IfSkip(token.KindEnum.Arrow); ok {
				returnType = optional.Some(p.parseType())
			}

			var body optional.Optional[*ast.Block]
			p.skip(token.KindEnum.Sem)
			if p.cur.Kind == token.KindEnum.Lbr {
				body = optional.Some(p.parseBlock())
			}

			return &ast.FuncExpr{Params: params, ReturnType: returnType, Body: body}
		}
		p.expect(token.KindEnum.Rpa)
		return expr
	default:
		panic(fmt.Sprintf("unexpected token `%s` in expression", p.cur.Kind))
	}
}
