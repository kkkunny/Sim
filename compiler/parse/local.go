package parse

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseLocal() ast.Local {
	switch p.nextToken.Kind {
	case token.KindEnum.Lbr:
		return p.parseBlock()
	case token.KindEnum.Return:
		return p.parseReturn()
	case token.KindEnum.Let:
		return p.parseLet()
	case token.KindEnum.If:
		return p.parseIf()
	default:
		return p.parseExpr()
	}
}

func (p *Parser) parseBlock() *ast.Block {
	begin := p.expect(token.KindEnum.Lbr).Position
	p.skip(token.KindEnum.Sem, token.KindEnum.Br)
	var stmts []ast.Local
	for p.nextToken.Kind != token.KindEnum.Rbr {
		stmts = append(stmts, p.parseLocal())
		if !p.ifSkip(token.KindEnum.Sem) && !p.ifSkip(token.KindEnum.Br) {
			break
		}
		p.skip(token.KindEnum.Sem, token.KindEnum.Br)
	}
	end := p.expect(token.KindEnum.Rbr).Position
	return &ast.Block{BeginPosition: begin, Stmts: stmts, EndPosition: end}
}

func (p *Parser) parseReturn() *ast.Return {
	p.expect(token.KindEnum.Return)

	var value optional.Optional[ast.Expr]
	if p.nextToken.Kind != token.KindEnum.Sem && p.nextToken.Kind != token.KindEnum.Rbr {
		value = optional.Some(p.parseExpr())
	}

	return &ast.Return{Value: value}
}

func (p *Parser) parseLet() *ast.Let {
	p.expect(token.KindEnum.Let)
	mut := p.ifSkip(token.KindEnum.Mut)
	name := p.expect(token.KindEnum.Ident)
	var t optional.Optional[ast.Type]
	if p.ifSkip(token.KindEnum.Col) {
		t = optional.Some(p.parseType())
	}
	var value optional.Optional[ast.Expr]
	if t.IsNone() || p.nextToken.Kind == token.KindEnum.Assign {
		p.expect(token.KindEnum.Assign)
		value = optional.Some(p.parseExpr())
	}
	return &ast.Let{Mut: mut, Name: name, Type: t, Value: value}
}

func (p *Parser) parseIf() *ast.If {
	p.expect(token.KindEnum.If)
	cond := p.parseExpr()
	body := p.parseBlock()
	var next optional.Optional[either.Either[*ast.If, *ast.Block]]
	if p.ifSkip(token.KindEnum.Else) {
		if p.nextToken.Kind == token.KindEnum.If {
			next = optional.Some(either.Left[*ast.If, *ast.Block](p.parseIf()))
		} else {
			next = optional.Some(either.Right[*ast.If, *ast.Block](p.parseBlock()))
		}
	}
	return &ast.If{Condition: cond, Body: body, Else: next}
}
