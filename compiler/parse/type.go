package parse

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseType() ast.Type {
	switch p.nextToken.Kind {
	case token.KindEnum.Lpa:
		begin := p.expect(token.KindEnum.Lpa).Position
		p.skip(token.KindEnum.Br)
		var params []ast.Type
		for p.nextToken.Kind != token.KindEnum.Rpa {
			params = append(params, p.parseType())
			if !p.ifSkip(token.KindEnum.Comma) {
				break
			}
			p.skip(token.KindEnum.Br)
		}
		p.expect(token.KindEnum.Rpa)

		var returnType optional.Optional[ast.Type]
		if p.ifSkip(token.KindEnum.Arrow) {
			returnType = optional.Some(p.parseType())
		}

		return &ast.FuncType{BeginPosition: begin, Params: params, ReturnType: returnType, EndPosition: p.curToken.Position}
	default:
		name := p.expect(token.KindEnum.Ident)
		return &ast.IdentType{Name: name}
	}
}
