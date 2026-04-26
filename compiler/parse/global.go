package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseGlobal() ast.Global {
	switch p.nextToken.Kind {
	case token.KindEnum.Let:
		return p.parseLet()
	case token.KindEnum.Type:
		return p.parseTypeDef()
	default:
		p.reporter.Fatalf(
			p.nextToken.Position,
			report.Errors.UnexpectedToken,
			p.nextToken.Kind,
		)
		return nil
	}
}

func (p *Parser) parseTypeDef() *ast.TypeDef {
	p.expect(token.KindEnum.Type)
	p.skip(token.KindEnum.Br)
	name := p.expect(token.KindEnum.Ident)
	typ := p.parseType()
	return &ast.TypeDef{Name: name, Type: typ}
}
