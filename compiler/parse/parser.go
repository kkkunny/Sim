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
	cur      token.Token
	reporter *report.Reporter
}

func New(l *lex.Lexer, reporter *report.Reporter) *Parser {
	return &Parser{
		lexer:    l,
		reporter: reporter,
	}
}

func (p *Parser) next() {
	p.cur = p.lexer.Scan()
}

func (p *Parser) skip(skips ...token.Kind) {
	for stlslices.Contain(skips, p.cur.Kind) {
		p.next()
	}
}

func (p *Parser) IfSkip(k token.Kind, skips ...token.Kind) (token.Token, bool) {
	p.skip(skips...)
	if p.cur.Kind != k {
		return token.Token{}, false
	}
	tok := p.cur
	p.next()
	return tok, true
}

func (p *Parser) expect(k token.Kind, skip ...token.Kind) token.Token {
	tok, ok := p.IfSkip(k, skip...)
	if !ok {
		p.reporter.Fatalf(
			p.cur.Position,
			report.Errors.ExpectedToken,
			k, p.cur.Kind,
		)
	}
	return tok
}

func (p *Parser) Parse() *ast.Program {
	p.next()
	var globals []ast.Global
	for {
		p.skip(token.KindEnum.Sem, token.KindEnum.Br)
		globals = append(globals, p.parseGlobal())
		if p.cur.Kind == token.KindEnum.Eof {
			break
		}
		p.expect(token.KindEnum.Sem)
	}
	return &ast.Program{Globals: globals}
}
