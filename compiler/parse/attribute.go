package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseAttribute() (attrs []ast.Attribute) {
	p.skip(token.KindEnum.Br)
	for p.ifSkip(token.KindEnum.At) {
		switch p.nextToken.Kind {
		case token.KindEnum.Extern:
			attrs = append(attrs, p.parseExtern())
		default:
			p.reporter.Fatalf(
				p.nextToken.Position,
				report.Errors.UnexpectedToken,
				p.nextToken.Kind,
			)
		}
		p.skip(token.KindEnum.Br)
	}
	return attrs
}

func (p *Parser) parseExtern() *ast.Extern {
	begin := p.curToken.Position
	p.expect(token.KindEnum.Extern)
	p.expect(token.KindEnum.Lpa)
	name := p.expect(token.KindEnum.Ident)
	end := p.expect(token.KindEnum.Rpa).Position
	return &ast.Extern{
		BeginPosition: begin,
		Name:          name,
		EndPosition:   end,
	}
}
