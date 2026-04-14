package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseExpr() ast.Expr {
	return &ast.IdentExpr{
		Name: p.expect(token.KindEnum.Ident),
	}
}
