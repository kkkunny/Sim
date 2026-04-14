package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseType() ast.Type {
	if p.cur.Kind == token.KindEnum.I32 {
		p.nextToken()
		return &ast.IdentType{Name: "i32"}
	}

	name := p.cur.OriginText
	p.expect(token.KindEnum.Ident)
	return &ast.IdentType{Name: name}
}
