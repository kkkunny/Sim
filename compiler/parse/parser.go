package parse

import (
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/queue"

	"github.com/kkkunny/Sim/compiler/ast"
	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/lex"

	errors "github.com/kkkunny/Sim/compiler/error"

	"github.com/kkkunny/Sim/compiler/token"
)

// Parser 语法分析器
type Parser struct {
	lexer           *lex.Lexer
	curTok, nextTok token.Token
	tokenCaches     queue.Queue[token.Token]
}

func New(lexer *lex.Lexer) *Parser {
	self := &Parser{
		lexer:       lexer,
		tokenCaches: queue.New[token.Token](),
	}
	self.next()
	return self
}

// 获取下一个token
func (self *Parser) next() {
	self.curTok = self.nextTok
	for {
		if !self.tokenCaches.Empty() {
			self.nextTok = self.tokenCaches.Pop()
		} else {
			self.nextTok = self.lexer.Scan()
		}
		if !self.nextIs(token.COMMENT) {
			break
		}
	}
}

// 下一个token是否是
func (self Parser) nextIs(k token.Kind) bool {
	return self.nextTok.Is(k)
}

// 下一个token是否属于
func (self Parser) nextIn(k ...token.Kind) bool {
	for _, kk := range k {
		if self.nextIs(kk) {
			return true
		}
	}
	return false
}

// 如果下一个token是则跳过
func (self *Parser) skipNextIs(k token.Kind) bool {
	if self.nextIs(k) {
		self.next()
		return true
	}
	return false
}

// 期待下一个token是
func (self *Parser) expectNextIs(k token.Kind) token.Token {
	if self.skipNextIs(k) {
		return self.curTok
	}
	// TODO: 规范化
	if k == token.GT && self.nextIs(token.SHR) {
		left, right := token.Token{
			Position: reader.Position{
				Reader:      self.nextTok.Position.Reader,
				BeginOffset: self.nextTok.Position.BeginOffset,
				EndOffset:   self.nextTok.Position.EndOffset - 1,
				BeginRow:    self.nextTok.Position.BeginRow,
				BeginCol:    self.nextTok.Position.BeginCol,
				EndRow:      self.nextTok.Position.EndRow,
				EndCol:      self.nextTok.Position.EndCol - 1,
			},
			Kind: token.GT,
		}, token.Token{
			Position: reader.Position{
				Reader:      self.nextTok.Position.Reader,
				BeginOffset: self.nextTok.Position.BeginOffset + 1,
				EndOffset:   self.nextTok.Position.EndOffset,
				BeginRow:    self.nextTok.Position.BeginRow,
				BeginCol:    self.nextTok.Position.BeginCol + 1,
				EndRow:      self.nextTok.Position.EndRow,
				EndCol:      self.nextTok.Position.EndCol,
			},
			Kind: token.GT,
		}
		self.nextTok = left
		self.tokenCaches.Push(right)
		self.next()
		return self.curTok
	} else if k == token.AND && self.nextIs(token.LAND) {
		left, right := token.Token{
			Position: reader.Position{
				Reader:      self.nextTok.Position.Reader,
				BeginOffset: self.nextTok.Position.BeginOffset,
				EndOffset:   self.nextTok.Position.EndOffset - 1,
				BeginRow:    self.nextTok.Position.BeginRow,
				BeginCol:    self.nextTok.Position.BeginCol,
				EndRow:      self.nextTok.Position.EndRow,
				EndCol:      self.nextTok.Position.EndCol - 1,
			},
			Kind: token.AND,
		}, token.Token{
			Position: reader.Position{
				Reader:      self.nextTok.Position.Reader,
				BeginOffset: self.nextTok.Position.BeginOffset + 1,
				EndOffset:   self.nextTok.Position.EndOffset,
				BeginRow:    self.nextTok.Position.BeginRow,
				BeginCol:    self.nextTok.Position.BeginCol + 1,
				EndRow:      self.nextTok.Position.EndRow,
				EndCol:      self.nextTok.Position.EndCol,
			},
			Kind: token.AND,
		}
		self.nextTok = left
		self.tokenCaches.Push(right)
		self.next()
		return self.curTok
	}
	errors.ThrowNotExpectToken(self.nextTok.Position, k, self.nextTok)
	return token.Token{}
}

// 跳过分隔符
func (self *Parser) skipSEM() {
	for self.skipNextIs(token.SEM) {
		continue
	}
}

// Parse 语法分析
func (self *Parser) Parse() linkedlist.LinkedList[ast.Global] {
	globals := linkedlist.NewLinkedList[ast.Global]()
	for {
		self.skipSEM()
		if self.nextIs(token.EOF) {
			break
		}

		globals.PushBack(self.parseGlobal())

		if !self.nextIs(token.EOF) {
			self.expectNextIs(token.SEM)
		}
	}
	return globals
}
