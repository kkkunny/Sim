package parse

import (
	"fmt"

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
		op := p.cur
		p.next()
		expr := p.parsePrimaryExpr()
		return &ast.UnaryExpr{Op: op, Expr: expr}
	case token.KindEnum.Ident:
		expr := &ast.IdentExpr{Name: p.cur}
		p.next()
		return expr
	case token.KindEnum.Integer:
		expr := &ast.IntegerExpr{Value: p.cur}
		p.next()
		return expr
	case token.KindEnum.Lpa:
		p.next()
		expr := p.parseExpr()
		p.expect(token.KindEnum.Rpa)
		return expr
	default:
		panic(fmt.Sprintf("unexpected token `%s` in expression", p.cur.Kind))
	}
}
