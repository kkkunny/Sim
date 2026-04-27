package parse

import (
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

type Parser struct {
	lexer    *lex.Lexer
	reporter *report.Reporter

	curToken, nextToken token.Token
}

func New(l *lex.Lexer, reporter *report.Reporter) *Parser {
	return &Parser{
		lexer:    l,
		reporter: reporter,
	}
}

func (p *Parser) next() {
	p.curToken, p.nextToken = p.nextToken, p.lexer.Scan()
}

func (p *Parser) skip(skips ...token.Kind) {
	for stlslices.Contain(skips, p.nextToken.Kind) {
		p.next()
	}
}

func (p *Parser) ifSkip(k token.Kind, skips ...token.Kind) bool {
	p.skip(skips...)
	if p.nextToken.Kind != k {
		return false
	}
	p.next()
	return true
}

func (p *Parser) expect(k token.Kind, skip ...token.Kind) token.Token {
	ok := p.ifSkip(k, skip...)
	if !ok {
		p.reporter.Fatalf(
			p.nextToken.Position,
			report.Errors.ExpectedToken,
			k, p.nextToken.Kind,
		)
	}
	return p.curToken
}

func (p *Parser) Parse() *ast.File {
	p.next()
	var globals []ast.Global
	for {
		p.skip(token.KindEnum.Sem, token.KindEnum.Br)
		if p.nextToken.Kind == token.KindEnum.Eof {
			break
		}
		globals = append(globals, p.parseGlobal())
	}
	return &ast.File{Globals: globals}
}
