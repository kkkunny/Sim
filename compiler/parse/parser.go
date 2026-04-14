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

func (p *Parser) Parse() *ast.Program {
	var funcs []*ast.FuncDecl
	for p.cur.Kind != token.KindEnum.Eof {
		if p.cur.Kind == token.KindEnum.Func {
			funcs = append(funcs, p.parseFuncDecl())
		} else {
			p.nextToken()
		}
	}
	return &ast.Program{Functions: funcs}
}
