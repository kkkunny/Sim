package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseParamDecl() *ast.ParamDecl {
	name := p.cur.OriginText
	p.expect(token.KindEnum.Ident)
	p.expect(token.KindEnum.Col)
	typ := p.parseType()
	return &ast.ParamDecl{Name: name, Type: typ}
}
