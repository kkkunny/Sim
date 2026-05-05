package parse

import (
	"github.com/kkkunny/stl/container/optional"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseGlobal(attrs []ast.Attribute) ast.Global {
	pub := p.ifSkip(token.KindEnum.Pub)
	switch p.nextToken.Kind {
	case token.KindEnum.Import:
		return p.parseImport()
	case token.KindEnum.Let:
		return p.parseLet(attrs, pub)
	case token.KindEnum.Type:
		return p.parseTypeDef(pub)
	default:
		p.reporter.Fatalf(
			p.nextToken.Position,
			report.Errors.UnexpectedToken,
			p.nextToken.Kind,
		)
		return nil
	}
}

func (p *Parser) parseImport() *ast.Import {
	p.expect(token.KindEnum.Import)
	var pkgs []token.Token
	for {
		pkgs = append(pkgs, p.expect(token.KindEnum.Ident))
		if !p.ifSkip(token.KindEnum.Scope) {
			break
		}
	}
	var alias optional.Optional[token.Token]
	if p.ifSkip(token.KindEnum.As) {
		if !p.ifSkip(token.KindEnum.Dot) {
			p.expect(token.KindEnum.Ident)
		}
		alias = optional.Some(p.curToken)
	}
	return &ast.Import{Alias: alias, Pkgs: pkgs}
}

func (p *Parser) parseTypeDef(pub bool) *ast.TypeDef {
	p.expect(token.KindEnum.Type)
	p.skip(token.KindEnum.Br)
	name := p.expect(token.KindEnum.Ident)
	typ := p.parseType()
	return &ast.TypeDef{
		Public: pub,
		Name:   name,
		Type:   typ,
	}
}
