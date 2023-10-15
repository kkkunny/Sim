package lex

import (
	"errors"
	"io"
	"strings"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/reader"
	. "github.com/kkkunny/Sim/token"
)

// Lexer 词法分析器
type Lexer struct {
	reader reader.Reader
}

func New(r reader.Reader) *Lexer {
	return &Lexer{reader: r}
}

// 下一个字符
func (self *Lexer) next() rune {
	r, _, err := stlerror.ErrorWith2(self.reader.ReadRune())
	if err != nil && errors.Is(err, io.EOF) {
		return 0
	} else if err != nil {
		panic(err)
	}
	return r
}

// 提前获取下一个字符
func (self Lexer) peek() rune {
	offset := stlerror.MustWith(self.reader.Seek(0, io.SeekCurrent))
	defer self.reader.Seek(offset, io.SeekStart)
	return self.next()
}

// 跳过空白
func (self *Lexer) skipWhite() {
	for ch := self.peek(); ch == ' ' || ch == '\r'; ch = self.peek() {
		_ = self.next()
	}
}

// 扫描标识符
func (self *Lexer) scanIdent(ch rune) Kind {
	var buf strings.Builder
	buf.WriteRune(ch)
	for ch = self.peek(); ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9'); ch = self.peek() {
		buf.WriteRune(self.next())
	}
	return Lookup(buf.String())
}

// 扫描整数
func (self *Lexer) scanInteger(ch rune) Kind {
	for ch = self.peek(); ch >= '0' && ch <= '9'; ch = self.peek() {
		self.next()
	}
	return INTEGER
}

func (self *Lexer) Scan() Token {
	self.skipWhite()

	begin := self.reader.Position()
	ch := self.next()

	var kind Kind
	switch {
	case ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z'):
		kind = self.scanIdent(ch)
	case ch >= '0' && ch <= '9':
		kind = self.scanInteger(ch)
	default:
		switch ch {
		case 0:
			kind = EOF
		case '(':
			kind = LPA
		case ')':
			kind = RPA
		case '{':
			kind = LBR
		case '}':
			kind = RBR
		case ';', '\n':
			kind = SEM
		default:
			kind = ILLEGAL
		}
	}

	end := self.reader.Position()
	return Token{
		Position: reader.MixPosition(begin, end),
		Kind:     kind,
	}
}
