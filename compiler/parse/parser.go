package parse

import (
	"fmt"

	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/token"
)

type Parser struct {
	lexer *lex.Lexer
	cur   token.Token
}

func New(l *lex.Lexer) *Parser {
	return &Parser{lexer: l}
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
		panic(fmt.Sprintf("expected %s but got %s at %s", k.String(), p.cur.Kind.String(), p.cur.Position.String()))
	}
	return tok
}

func (p *Parser) Parse() *ast.Program {
	p.next()
	var funcs []*ast.FuncDecl
	for p.cur.Kind != token.KindEnum.Eof {
		if p.cur.Kind == token.KindEnum.Let {
			funcs = append(funcs, p.parseFuncDecl())
		} else {
			p.next()
		}
	}
	return &ast.Program{Functions: funcs}
}
