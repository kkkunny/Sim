package parse

import (
	"github.com/kkkunny/stl/container/optional"
	"github.com/kkkunny/stl/container/tuple"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseLocal() ast.Local {
	switch p.cur.Kind {
	case token.KindEnum.Lbr:
		return p.parseBlock()
	case token.KindEnum.Return:
		return p.parseReturn()
	case token.KindEnum.Let:
		return p.parseLet()
	default:
		return p.parseExpr()
	}
}

func (p *Parser) parseBlock() *ast.Block {
	p.expect(token.KindEnum.Lbr)
	p.skip(token.KindEnum.Sem, token.KindEnum.Br)
	var stmts []ast.Local
	for p.cur.Kind != token.KindEnum.Rbr {
		stmts = append(stmts, p.parseLocal())
		if !tuple.Pack2(p.IfSkip(token.KindEnum.Sem)).E2() && !tuple.Pack2(p.IfSkip(token.KindEnum.Br)).E2() {
			break
		}
		p.skip(token.KindEnum.Sem, token.KindEnum.Br)
	}
	p.expect(token.KindEnum.Rbr)
	return &ast.Block{Stmts: stmts}
}

func (p *Parser) parseReturn() *ast.Return {
	p.expect(token.KindEnum.Return)

	var value optional.Optional[ast.Expr]
	if p.cur.Kind != token.KindEnum.Sem && p.cur.Kind != token.KindEnum.Rbr {
		value = optional.Some(p.parseExpr())
	}

	return &ast.Return{
		Value: value,
	}
}

func (p *Parser) parseLet() *ast.Let {
	p.expect(token.KindEnum.Let)
	name := p.expect(token.KindEnum.Ident)
	p.expect(token.KindEnum.Assign)
	p.skip(token.KindEnum.Br)
	value := p.parseExpr()
	return &ast.Let{Name: name, Value: value}
}
