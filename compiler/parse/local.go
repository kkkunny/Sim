package parse

import (
	"fmt"

	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseLocal() ast.Local {
	switch p.cur.Kind {
	case token.KindEnum.Lbr:
		return p.parseBlock()
	case token.KindEnum.Return:
		return p.parseReturn()
	default:
		panic(fmt.Sprintf("invalid local: %s", p.cur.Kind))
	}
}

func (p *Parser) parseBlock() *ast.Block {
	p.expect(token.KindEnum.Lbr, token.KindEnum.Sem)
	p.skip(token.KindEnum.Sem)
	var stmts []ast.Local
	for p.cur.Kind != token.KindEnum.Rbr {
		stmts = append(stmts, p.parseLocal())
		if _, ok := p.IfSkip(token.KindEnum.Sem); ok {
			break
		}
		p.skip(token.KindEnum.Sem)
	}
	p.expect(token.KindEnum.Rbr, token.KindEnum.Sem)
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
