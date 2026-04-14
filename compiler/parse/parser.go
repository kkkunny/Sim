package parse

import (
	"fmt"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/token"
)

type Parser struct {
	lexer *lex.Lexer
	cur   token.Token
}

func New(l *lex.Lexer) *Parser {
	p := &Parser{lexer: l}
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.cur = p.lexer.Scan()
}

func (p *Parser) expect(k token.Kind) {
	if p.cur.Kind != k {
		panic(fmt.Sprintf("expected %s but got %s at %s", k.String(), p.cur.Kind.String(), p.cur.Position.String()))
	}
	p.nextToken()
}

func (p *Parser) Parse() ast.Ast {
	return p.parseFuncDecl()
}

func (p *Parser) parseFuncDecl() *ast.FuncDecl {
	p.expect(token.KindEnum.Func)
	name := p.cur.OriginText
	p.expect(token.KindEnum.Ident)
	p.expect(token.KindEnum.Lpa)
	p.expect(token.KindEnum.Rpa)
	p.expect(token.KindEnum.Lbr)

	for p.cur.Kind != token.KindEnum.Rbr && p.cur.Kind != token.KindEnum.Eof {
		p.nextToken()
	}

	p.expect(token.KindEnum.Rbr)
	return &ast.FuncDecl{Name: name}
}
