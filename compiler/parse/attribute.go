package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseAttribute() (attrs []ast.Attribute) {
	p.skip(token.KindEnum.Br)
	for p.ifSkip(token.KindEnum.At) {
		begin := p.curToken.Position
		name := p.expect(token.KindEnum.Ident).OriginText
		switch name {
		case "extern":
			attrs = append(attrs, p.parseExtern(begin))
		default:
			p.reporter.Fatalf(
				reader.MixPosition(begin, p.curToken.Position),
				report.Errors.UnknownAttribute,
				name,
			)
		}
		p.skip(token.KindEnum.Br)
	}
	return attrs
}

func (p *Parser) parseExtern(begin reader.Position) *ast.Extern {
	p.expect(token.KindEnum.Lpa)
	name := p.expect(token.KindEnum.Ident)
	end := p.expect(token.KindEnum.Rpa).Position
	return &ast.Extern{
		BeginPosition: begin,
		Name:          name,
		EndPosition:   end,
	}
}
