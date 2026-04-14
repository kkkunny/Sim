package parse

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseFuncDecl() *ast.FuncDecl {
	p.expect(token.KindEnum.Let)
	name := p.expect(token.KindEnum.Ident)
	p.expect(token.KindEnum.Assign)
	p.expect(token.KindEnum.Lpa)

	var params []*ast.ParamDecl
	if p.cur.Kind != token.KindEnum.Rpa {
		params = append(params, p.parseParamDecl())
		for p.cur.Kind == token.KindEnum.Comma {
			p.next()
			params = append(params, p.parseParamDecl())
		}
	}
	p.expect(token.KindEnum.Rpa)

	var returnType optional.Optional[ast.Type]
	if _, ok := p.IfSkip(token.KindEnum.Arrow); ok {
		returnType = optional.Some(p.parseType())
	}

	var body optional.Optional[*ast.Block]
	p.skip(token.KindEnum.Sem)
	if p.cur.Kind == token.KindEnum.Lbr {
		body = optional.Some(p.parseBlock())
	}

	return &ast.FuncDecl{Name: name, Params: params, ReturnType: returnType, Body: body}
}
