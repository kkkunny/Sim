package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseParamDecl() *ast.ParamDecl {
	mut := p.ifSkip(token.KindEnum.Mut)
	name := p.expect(token.KindEnum.Ident)
	p.expect(token.KindEnum.Col)
	typ := p.parseType()
	return &ast.ParamDecl{Mut: mut, Name: name, Type: typ}
}
