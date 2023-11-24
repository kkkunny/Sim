package lex

import (
	"errors"
	"io"
	"strings"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/token"

	"github.com/kkkunny/Sim/reader"
	"github.com/kkkunny/Sim/util"
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
func (self Lexer) peek(skip ...uint) rune {
	defer self.reader.Seek(int64(self.reader.Offset()), io.SeekStart)

	offset := 1
	if len(skip) > 0 {
		offset += int(skip[0])
	}

	var ch rune
	for i := 0; i < offset; i++ {
		ch = self.next()
	}
	return ch
}

// 跳过空白
func (self *Lexer) skipWhite() {
	for ch := self.peek(); ch == ' ' || ch == '\r' || ch == '\t'; ch = self.peek() {
		_ = self.next()
	}
}

// 扫描标识符
func (self *Lexer) scanIdent(ch rune) token.Kind {
	var buf strings.Builder
	buf.WriteRune(ch)
	for ch = self.peek(); ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9'); ch = self.peek() {
		buf.WriteRune(self.next())
	}
	return token.Lookup(buf.String())
}

// 扫描数字
func (self *Lexer) scanNumber(ch rune) token.Kind {
	var point bool
	for ch = self.peek(); ch == '.' || (ch >= '0' && ch <= '9'); ch = self.peek() {
		if ch == '.' {
			if ch2 := self.peek(1); !(ch2 >= '0' && ch2 <= '9') || point {
				break
			} else {
				point = true
			}
		}
		self.next()
	}
	if point {
		return token.FLOAT
	} else {
		return token.INTEGER
	}
}

// 扫描字符
func (self *Lexer) scanChar(ch rune) token.Kind {
	var buf strings.Builder
	prevChar := ch
	for ch = self.peek(); ch != '\'' || prevChar == '\\'; ch, prevChar = self.peek(), ch {
		if ch == 0 {
			return token.ILLEGAL
		}
		self.next()
		buf.WriteRune(ch)
	}
	self.next()

	s := util.ParseEscapeCharacter(buf.String(), `\'`, `'`)
	if len(s) != 1 {
		return token.ILLEGAL
	}
	return token.CHAR
}

// 扫描字符串
func (self *Lexer) scanString(ch rune) token.Kind {
	prevChar := ch
	for ch = self.peek(); ch != '"' || prevChar == '\\'; ch, prevChar = self.peek(), ch {
		if ch == 0 {
			return token.ILLEGAL
		}
		self.next()
	}
	self.next()
	return token.STRING
}

func (self *Lexer) Scan() token.Token {
	self.skipWhite()

	begin := self.reader.Position()
	ch := self.next()

	var kind token.Kind
	switch {
	case ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z'):
		kind = self.scanIdent(ch)
	case ch >= '0' && ch <= '9':
		kind = self.scanNumber(ch)
	case ch == '\'':
		kind = self.scanChar(ch)
	case ch == '"':
		kind = self.scanString(ch)
	default:
		switch ch {
		case 0:
			kind = token.EOF
		case '=':
			kind = token.ASS
			if nextCh := self.peek(); nextCh == '=' {
				self.next()
				kind = token.EQ
			}
		case '&':
			kind = token.AND
			if nextCh := self.peek(); nextCh == '&' {
				self.next()
				kind = token.LAND
			}
		case '|':
			kind = token.OR
			if nextCh := self.peek(); nextCh == '|' {
				self.next()
				kind = token.LOR
			}
		case '^':
			kind = token.XOR
		case '!':
			kind = token.NOT
			if nextCh := self.peek(); nextCh == '=' {
				self.next()
				kind = token.NE
			}
		case '+':
			kind = token.ADD
		case '-':
			kind = token.SUB
		case '*':
			kind = token.MUL
		case '/':
			kind = token.DIV
		case '%':
			kind = token.REM
		case '<':
			kind = token.LT
			if nextCh := self.peek(); nextCh == '=' {
				self.next()
				kind = token.LE
			} else if nextCh == '<' {
				self.next()
				kind = token.SHL
			}
		case '>':
			kind = token.GT
			if nextCh := self.peek(); nextCh == '=' {
				self.next()
				kind = token.GE
			} else if nextCh == '>' {
				self.next()
				kind = token.SHR
			}
		case '(':
			kind = token.LPA
		case ')':
			kind = token.RPA
		case '[':
			kind = token.LBA
		case ']':
			kind = token.RBA
		case '{':
			kind = token.LBR
		case '}':
			kind = token.RBR
		case ';', '\n':
			kind = token.SEM
		case ',':
			kind = token.COM
		case '.':
			kind = token.DOT
		case ':':
			kind = token.COL
			if nextCh := self.peek(); nextCh == ':' {
				self.next()
				kind = token.SCOPE
			}
		case '@':
			kind = token.AT
		default:
			kind = token.ILLEGAL
		}
	}

	end := self.reader.Position()
	return token.Token{
		Position: reader.MixPosition(begin, end),
		Kind:     kind,
	}
}
