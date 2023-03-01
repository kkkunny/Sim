package parse

import (
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/utils"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/queue"
)

// Ast 抽象语法树
type Ast interface {
	Position() utils.Position // 获取位置
}

// NameAndType 名字和类型
type NameAndType struct {
	Name lex.Token
	Type Type
}

func NewNameAndType(name lex.Token, t Type) *NameAndType {
	return &NameAndType{
		Name: name,
		Type: t,
	}
}

func (self NameAndType) Position() utils.Position {
	return utils.MixPosition(self.Name.Pos, self.Type.Position())
}

// ****************************************************************

// 所有导入的包，包括main包
var pkgs = make(map[stlos.Path]*Package)

// 语法分析器
type parser struct {
	pkg       *Package
	lexer     *lex.Lexer              // 词法分析器
	curTok    lex.Token               // 当前token
	nextTok   lex.Token               // 待分析token
	tokenPool *queue.Queue[lex.Token] // token缓存池
}

// 新建语法分析器
func newParser(pkg *Package, lexer *lex.Lexer) *parser {
	parser := &parser{
		pkg:       pkg,
		lexer:     lexer,
		tokenPool: queue.NewQueue[lex.Token](),
	}
	parser.next()
	return parser
}

// 从词法分析器或token池中获取一个token
func (self *parser) scanToken() lex.Token {
	if !self.tokenPool.Empty() {
		return self.tokenPool.Pop()
	} else {
		return self.lexer.Scan()
	}
}

// 读取下一个token
func (self *parser) next() {
	self.curTok = self.nextTok
	token := self.scanToken()
	for token.Kind == lex.COMMENT {
		token = self.scanToken()
	}
	self.nextTok = token
}

// 下一个token是否是
func (self *parser) nextIs(kind lex.TokenKind) bool {
	return self.nextTok.Kind == kind
}

// 如果下一个token是，则跳过
func (self *parser) skipNextIs(kind lex.TokenKind) bool {
	if self.nextIs(kind) {
		self.next()
		return true
	}
	return false
}

// 如果下一个token是，则跳过，若不是，则报错
func (self *parser) expectNextIs(kind lex.TokenKind) lex.Token {
	if !self.skipNextIs(kind) {
		self.throwErrorf(self.nextTok.Pos, "expect token `%s`", kind)
	}
	return self.curTok
}

// 跳过分隔符
func (self *parser) skipSem() {
	for self.nextIs(lex.SEM) {
		self.next()
	}
}

// 至少有一个分隔符，跳过多余的分隔符
func (self *parser) skipSemAtLeastOne() {
	self.expectNextIs(lex.SEM)
	self.skipSem()
}

// 退回一个token到token池
func (self *parser) backToTokenPool(tok lex.Token) {
	self.tokenPool.Push(tok)
}

// 退回一个token到nextToken
func (self *parser) backToNextToken(tok lex.Token) {
	self.backToTokenPool(self.nextTok)
	self.nextTok = tok
}

// 语法分析
func (self *parser) parse() (err utils.Error) {
	defer func() {
		if ea := recover(); ea != nil {
			if e, ok := ea.(utils.Error); ok {
				err = e
			}
			panic(ea)
		}
	}()
	self.parseFile()
	return nil
}

// 抛出异常
func (self *parser) throwErrorf(pos utils.Position, f string, a ...any) {
	panic(utils.Errorf(pos, f, a...))
}

// 文件
func (self *parser) parseFile() {
	for self.skipSem(); !self.nextIs(lex.EOF); {
		if global := self.parseGlobal(); global != nil {
			self.pkg.Globals.Add(global)
		}
		if !self.nextIs(lex.EOF) {
			self.skipSemAtLeastOne()
		}
	}
}

// token列表
func (self *parser) parseTokenList(sep lex.TokenKind) (toks []lex.Token) {
	for {
		if len(toks) == 0 {
			if !self.skipNextIs(lex.IDENT) {
				break
			}
			toks = append(toks, self.curTok)
		} else {
			toks = append(toks, self.expectNextIs(lex.IDENT))
		}
		if !self.skipNextIs(sep) {
			break
		}
	}
	return toks
}

// token列表（至少一个）
func (self *parser) parseTokenListAtLeastOne(sep lex.TokenKind) (toks []lex.Token) {
	for {
		toks = append(toks, self.expectNextIs(lex.IDENT))
		if !self.skipNextIs(sep) {
			break
		}
	}
	return toks
}

// 名字和类型
func (self *parser) parseNameAndType(ft bool) *NameAndType {
	name := self.expectNextIs(lex.IDENT)
	self.expectNextIs(lex.COL)
	var typ Type
	if ft {
		typ = self.parseTypeFunc()
	} else {
		typ = self.parseType()
	}
	return NewNameAndType(name, typ)
}
