package parse

import (
	"github.com/kkkunny/Sim/src/compiler/lex"
	"github.com/kkkunny/Sim/src/compiler/utils"
	"github.com/kkkunny/stl/list"
	stlos "github.com/kkkunny/stl/os"
	"github.com/kkkunny/stl/queue"
)

// Ast 抽象语法树
type Ast interface {
	Position() utils.Position // 获取位置
}

// Package 文件
type Package struct {
	Path  stlos.Path
	Files []*File
}

func NewPackage(path stlos.Path, file ...*File) *Package {
	return &Package{
		Path:  path,
		Files: file,
	}
}

// File 文件
type File struct {
	Path    stlos.Path
	Globals *list.SingleLinkedList[Global]
}

func NewFile(path stlos.Path) *File {
	return &File{
		Path:    path,
		Globals: list.NewSingleLinkedList[Global](),
	}
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

// NameOrNilAndType 名字（或空）和类型
type NameOrNilAndType struct {
	Name *lex.Token
	Type Type
}

func NewNameOrNilAndType(name *lex.Token, t Type) *NameOrNilAndType {
	return &NameOrNilAndType{
		Name: name,
		Type: t,
	}
}

func (self NameOrNilAndType) Position() utils.Position {
	if self.Name != nil {
		return utils.MixPosition(self.Name.Pos, self.Type.Position())
	} else {
		return self.Type.Position()
	}
}

// ****************************************************************

// Parser 语法分析器
type Parser struct {
	lexer     *lex.Lexer              // 词法分析器
	curTok    lex.Token               // 当前token
	nextTok   lex.Token               // 待分析token
	tokenPool *queue.Queue[lex.Token] // token缓存池
}

// NewParser 新建语法分析器
func NewParser(lexer *lex.Lexer) *Parser {
	parser := &Parser{
		lexer:     lexer,
		tokenPool: queue.NewQueue[lex.Token](),
	}
	parser.next()
	return parser
}

// 从词法分析器或token池中获取一个token
func (self *Parser) scanToken() lex.Token {
	if !self.tokenPool.Empty() {
		return self.tokenPool.Pop()
	} else {
		return self.lexer.Scan()
	}
}

// 读取下一个token
func (self *Parser) next() {
	self.curTok = self.nextTok
	token := self.scanToken()
	for token.Kind == lex.COMMENT {
		token = self.scanToken()
	}
	self.nextTok = token
}

// 下一个token是否是
func (self *Parser) nextIs(kind lex.TokenKind) bool {
	return self.nextTok.Kind == kind
}

// 如果下一个token是，则跳过
func (self *Parser) skipNextIs(kind lex.TokenKind) bool {
	if self.nextIs(kind) {
		self.next()
		return true
	}
	return false
}

// 如果下一个token是，则跳过，若不是，则报错
func (self *Parser) expectNextIs(kind lex.TokenKind) lex.Token {
	if !self.skipNextIs(kind) {
		self.throwErrorf(self.nextTok.Pos, "expect token `%s`", kind)
	}
	return self.curTok
}

// 跳过分隔符
func (self *Parser) skipSem() {
	for self.nextIs(lex.SEM) {
		self.next()
	}
}

// 至少有一个分隔符，跳过多余的分隔符
func (self *Parser) skipSemAtLeastOne() {
	self.expectNextIs(lex.SEM)
	self.skipSem()
}

// 退回一个token到token池
func (self *Parser) backToTokenPool(tok lex.Token) {
	self.tokenPool.Push(tok)
}

// 退回一个token到nextToken
func (self *Parser) backToNextToken(tok lex.Token) {
	self.backToTokenPool(self.nextTok)
	self.nextTok = tok
}

// Parse 语法分析
func (self *Parser) Parse() (file *File, err utils.Error) {
	defer func() {
		if ea := recover(); ea != nil {
			if e, ok := ea.(utils.Error); ok {
				err = e
			}
			panic(ea)
		}
	}()
	return self.parseFile(), nil
}

// 抛出异常
func (self *Parser) throwErrorf(pos utils.Position, f string, a ...any) {
	panic(utils.Errorf(pos, f, a...))
}

// 文件
func (self *Parser) parseFile() *File {
	file := NewFile(self.lexer.GetFilepath())

	for self.skipSem(); !self.nextIs(lex.EOF); self.skipSem() {
		file.Globals.Add(self.parseGlobal())
		if !self.nextIs(lex.EOF) {
			self.expectNextIs(lex.SEM)
		}
	}

	return file
}

// token列表
func (self *Parser) parseTokenList(sep lex.TokenKind) (toks []lex.Token) {
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
func (self *Parser) parseTokenListAtLeastOne(sep lex.TokenKind) (toks []lex.Token) {
	for {
		toks = append(toks, self.expectNextIs(lex.IDENT))
		if !self.skipNextIs(sep) {
			break
		}
	}
	return toks
}

// 名字和类型
func (self *Parser) parseNameAndType(ft bool) *NameAndType {
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

// 名字（或空）和类型列表
func (self *Parser) parseNameOrNilAndTypeList(sep lex.TokenKind, skipSem bool) (toks []*NameOrNilAndType) {
	for {
		typ := self.parseTypeOrNil()
		if typ == nil {
			break
		}
		var name *lex.Token
		if ident, ok := typ.(*TypeIdent); ok && ident.Pkg == nil && self.skipNextIs(lex.COL) {
			name = &ident.Name
			typ = self.parseType()
		}
		toks = append(toks, NewNameOrNilAndType(name, typ))
		if !self.skipNextIs(sep) {
			break
		}
		if skipSem {
			self.skipSem()
		}
	}
	return toks
}
