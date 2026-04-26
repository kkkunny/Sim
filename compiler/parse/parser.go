package parse

import (
	"os"
	"path/filepath"

	stlslices "github.com/kkkunny/stl/container/slices"
	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/reader"
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

func (p *Parser) Parse() *ast.Program {
	p.next()
	var globals []ast.Global
	for {
		p.skip(token.KindEnum.Sem, token.KindEnum.Br)
		if p.nextToken.Kind == token.KindEnum.Eof {
			break
		}
		globals = append(globals, p.parseGlobal())
	}
	return &ast.Program{Globals: globals}
}

func parseDir(dirPath string, reporter *report.Reporter) (*ast.Program, error) {
	entries, err := stlerr.ErrorWith(os.ReadDir(dirPath))
	if err != nil {
		return nil, err
	}
	var globals []ast.Global
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sim" {
			continue
		}
		filePath := filepath.Join(dirPath, entry.Name())
		fileProgram, err := parseFile(filePath, reporter)
		if err != nil {
			return nil, err
		}
		globals = append(globals, fileProgram.Globals...)
	}
	return &ast.Program{Globals: globals}, nil
}

func parseFile(filePath string, reporter *report.Reporter) (*ast.Program, error) {
	file, err := stlerr.ErrorWith(os.Open(filePath))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	lexer := lex.New(reader.NewFile(filePath, file))
	parser := New(lexer, reporter)
	return parser.Parse(), nil
}

func Parse(path string, reporter *report.Reporter) (*ast.Program, error) {
	info, err := stlerr.ErrorWith(os.Stat(path))
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return parseDir(path, reporter)
	} else {
		return parseFile(path, reporter)
	}
}
