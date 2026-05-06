package parse

import (
	"github.com/kkkunny/stl/container/either"
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseLocal() ast.Local {
	switch p.nextToken.Kind {
	case token.KindEnum.Lbr:
		return p.parseBlock()
	case token.KindEnum.Return:
		return p.parseReturn()
	case token.KindEnum.Let:
		return p.parseLet(nil, false)
	case token.KindEnum.If:
		return p.parseIf()
	case token.KindEnum.For:
		return p.parseFor()
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

func (p *Parser) parseLet(attrs []ast.Attribute, pub bool) *ast.Let {
	p.expect(token.KindEnum.Let)
	mut := p.ifSkip(token.KindEnum.Mut)
	name := p.expect(token.KindEnum.Ident)

	if name.OriginText == "main" {
		if mut {
			p.reporter.Fatalf(
				name.Position,
				report.Errors.MustImmutable,
			)
		} else if len(attrs) > 0 {
			p.reporter.Fatalf(
				attrs[0].Position(),
				report.Errors.InvalidAttribute,
				attrs[0].AttrName(), "main",
			)
		}
	}

	var bind optional.Optional[ast.Type]
	if p.ifSkip(token.KindEnum.Or) {
		bind = optional.Some(p.parseType())
	}

	var t optional.Optional[ast.Type]
	if p.ifSkip(token.KindEnum.Col) {
		t = optional.Some(p.parseType())
	}
	var value optional.Optional[ast.Expr]
	if t.IsNone() || p.nextToken.Kind == token.KindEnum.Assign {
		p.expect(token.KindEnum.Assign)
		value = optional.Some(p.parseExpr())
	}
	return &ast.Let{
		Attributes: attrs,
		Public:     pub,
		Mut:        mut,
		Name:       name,
		Type:       t,
		Value:      value,

		Bind: bind,
	}
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

func (p *Parser) parseFor() ast.Local {
	p.expect(token.KindEnum.For)
	if p.ifSkip(token.KindEnum.Mut) {
		varname := p.expect(token.KindEnum.Ident)
		p.expect(token.KindEnum.In)
		rangeValue := p.parseExpr()
		body := p.parseBlock()
		return &ast.For{
			Mut:      true,
			Variable: varname,
			Range:    rangeValue,
			Body:     body,
		}
	}
	var cond optional.Optional[ast.Expr]
	if p.nextToken.Kind != token.KindEnum.Lbr {
		cond = optional.Some(p.parseExpr())
		if ident, ok := cond.MustValue().(*ast.IdentExpr); ok && ident.Pkg.IsNone() && p.ifSkip(token.KindEnum.In) {
			rangeValue := p.parseExpr()
			body := p.parseBlock()
			return &ast.For{
				Mut:      false,
				Variable: ident.Name,
				Range:    rangeValue,
				Body:     body,
			}
		}
	}
	body := p.parseBlock()
	return &ast.While{Condition: cond, Body: body}
}
