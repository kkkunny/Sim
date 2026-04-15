package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseType() ast.Type {
	name := p.expect(token.KindEnum.Ident)
	return &ast.IdentType{Name: name}
}
