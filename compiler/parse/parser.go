package parse

import (
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/lex"
	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/report"
	"github.com/kkkunny/Sim/compiler/token"
)

// parseAbort 当前语句/声明解析失败（诊断已记录），panic 后由恢复点跳过该构造并继续。
type parseAbort struct{}

type Parser struct {
	lexer    *lex.Lexer
	reporter *report.Reporter

	curToken, nextToken token.Token
	lexFailed           bool
}

func New(l *lex.Lexer, reporter *report.Reporter) *Parser {
	return &Parser{
		lexer:    l,
		reporter: reporter,
	}
}

func (p *Parser) next() {
	p.curToken, p.nextToken = p.nextToken, p.scan()
}

// scan 调用词法器；词法错误转为诊断并中止当前构造。
func (p *Parser) scan() token.Token {
	if p.lexFailed {
		return token.Token{Kind: token.KindEnum.Eof, Position: p.curToken.Position}
	}
	defer func() {
		if r := recover(); r != nil {
			if lexErr, ok := r.(*lex.Error); ok {
				p.lexFailed = true
				p.reporter.Errorf(lexErr.Pos, report.Errors.InvalidSource, lexErr)
				panic(parseAbort{})
			}
			panic(r)
		}
	}()
	return p.lexer.Scan()
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
		p.errorAt(
			p.nextToken.Position,
			report.Errors.ExpectedToken,
			k, p.nextToken.Kind,
		)
	}
	return p.curToken
}

// errorAt 记录诊断并中止当前语句/声明；词法错误已报告时不再级联报告。
func (p *Parser) errorAt(pos reader.Position, err report.ErrorType, args ...any) {
	if !p.lexFailed {
		p.reporter.Errorf(pos, err, args...)
	}
	panic(parseAbort{})
}

// Parse 解析整个文件；单个全局声明失败不会影响其余声明的解析。
func (p *Parser) Parse() *ast.File {
	p.next()
	var globals []ast.Global
	for {
		p.skip(token.KindEnum.Br)
		if p.nextToken.Kind == token.KindEnum.Eof {
			break
		}
		before := p.nextToken.Position.BeginOffset
		if !p.parseGlobalSafe(&globals) {
			p.syncGlobal()
			// 恢复后仍停在原处则强制前进，保证解析推进
			if p.nextToken.Kind != token.KindEnum.Eof && p.nextToken.Position.BeginOffset == before {
				p.next()
			}
		}
	}
	return &ast.File{Globals: globals}
}

// parseGlobalSafe 解析单个全局声明，失败时丢弃该声明并返回 false。
func (p *Parser) parseGlobalSafe(globals *[]ast.Global) (ok bool) {
	ok = true
	func() {
		defer func() {
			if r := recover(); r != nil {
				if _, aborted := r.(parseAbort); !aborted {
					panic(r)
				}
				ok = false
			}
		}()
		attrs := p.parseAttribute()
		*globals = append(*globals, p.parseGlobal(attrs))
	}()
	return ok
}

// syncGlobal 跳到下一个可能的全局声明起始处（或文件末尾）。
func (p *Parser) syncGlobal() {
	for p.nextToken.Kind != token.KindEnum.Eof {
		switch p.nextToken.Kind {
		case token.KindEnum.At, token.KindEnum.Pub, token.KindEnum.Import, token.KindEnum.Let,
			token.KindEnum.Type, token.KindEnum.Br:
			return
		}
		p.next()
	}
}

// syncLocal 跳到语句边界（分号）或块结尾。
func (p *Parser) syncLocal() {
	for {
		switch p.nextToken.Kind {
		case token.KindEnum.Sem, token.KindEnum.Rbr, token.KindEnum.Eof:
			return
		}
		p.next()
	}
}
