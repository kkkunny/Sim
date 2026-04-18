package parse

import (
	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

func (p *Parser) parseGlobal() ast.Global {
	switch p.nextToken.Kind {
	case token.KindEnum.Let:
		return p.parseGlobalLet()
	default:
		p.reporter.Fatalf(
			p.nextToken.Position,
			report.Errors.UnexpectedToken,
			p.nextToken.Kind,
		)
		return nil
	}
}

func (p *Parser) parseGlobalLet() ast.Global {
	p.expect(token.KindEnum.Let)
	name := p.expect(token.KindEnum.Ident)
	p.expect(token.KindEnum.Assign)
	value := p.parseExpr()
	return &ast.Let{Name: name, Value: value}
}
