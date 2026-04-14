package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseType() ast.Type {
	if tok, ok := p.IfSkip(token.KindEnum.I32); ok {
		return &ast.IdentType{Name: tok}
	}
	name := p.expect(token.KindEnum.Ident)
	return &ast.IdentType{Name: name}
}
