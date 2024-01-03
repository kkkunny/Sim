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
func (self *Lexer) next() byte {
	b, err := stlerror.ErrorWith(self.reader.ReadByte())
	if err != nil && errors.Is(err, io.EOF) {
		return 0
	} else if err != nil {
		panic(err)
	}
	return b
}

// 提前获取下一个字符
func (self Lexer) peek(skip ...uint) byte {
	defer self.reader.Seek(int64(self.reader.Offset()), io.SeekStart)

	offset := 1
	if len(skip) > 0 {
		offset += int(skip[0])
	}

	var b byte
	for i := 0; i < offset; i++ {
		b = self.next()
	}
	return b
}

// 跳过空白
func (self *Lexer) skipWhite() {
	for b := self.peek(); b == ' ' || b == '\r' || b == '\t'; b = self.peek() {
		_ = self.next()
	}
}

// 扫描标识符
func (self *Lexer) scanIdent(b byte) token.Kind {
	var buf strings.Builder
	buf.WriteByte(b)
	for b = self.peek(); b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9'); b = self.peek() {
		buf.WriteByte(self.next())
	}
	return token.Lookup(buf.String())
}

// 扫描数字
func (self *Lexer) scanNumber(b byte) token.Kind {
	var point bool
	for b = self.peek(); b == '.' || (b >= '0' && b <= '9'); b = self.peek() {
		if b == '.' {
			if b2 := self.peek(1); !(b2 >= '0' && b2 <= '9') || point {
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
func (self *Lexer) scanChar(b byte) token.Kind {
	var buf strings.Builder
	prevChar := b
	for b = self.peek(); b != '\'' || prevChar == '\\'; b, prevChar = self.peek(), b {
		if b == 0 {
			return token.ILLEGAL
		}
		self.next()
		buf.WriteByte(b)
	}
	self.next()

	s := util.ParseEscapeCharacter(buf.String(), `\'`, `'`)
	if len(s) != 1 {
		return token.ILLEGAL
	}
	return token.CHAR
}

// 扫描字符串
func (self *Lexer) scanString(b byte) token.Kind {
	prevChar := b
	for b = self.peek(); b != '"' || prevChar == '\\'; b, prevChar = self.peek(), b {
		if b == 0 {
			return token.ILLEGAL
		}
		self.next()
	}
	self.next()
	return token.STRING
}

// 扫描数字
func (self *Lexer) scanComment() token.Kind {
	for b := self.peek(); b != '\n' && b != 0; b = self.peek() {
		self.next()
	}
	return token.COMMENT
}

func (self *Lexer) Scan() token.Token {
	self.skipWhite()

	begin := self.reader.Position()
	b := self.next()

	var kind token.Kind
	switch {
	case b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z'):
		kind = self.scanIdent(b)
	case b >= '0' && b <= '9':
		kind = self.scanNumber(b)
	case b == '\'':
		kind = self.scanChar(b)
	case b == '"':
		kind = self.scanString(b)
	default:
		switch b {
		case 0:
			kind = token.EOF
		case '=':
			kind = token.ASS
			if nb := self.peek(); nb == '=' {
				self.next()
				kind = token.EQ
			}
		case '&':
			kind = token.AND
			if nb := self.peek(); nb == '&' {
				self.next()
				kind = token.LAND
			}
		case '|':
			kind = token.OR
			if nb := self.peek(); nb == '|' {
				self.next()
				kind = token.LOR
			}
		case '^':
			kind = token.XOR
		case '!':
			kind = token.NOT
			if nb := self.peek(); nb == '=' {
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
			if nb := self.peek(); nb == '/' {
				self.next()
				kind = self.scanComment()
			}
		case '%':
			kind = token.REM
		case '<':
			kind = token.LT
			if nb := self.peek(); nb == '=' {
				self.next()
				kind = token.LE
			} else if nb == '<' {
				self.next()
				kind = token.SHL
			}
		case '>':
			kind = token.GT
			if nb := self.peek(); nb == '=' {
				self.next()
				kind = token.GE
			} else if nb == '>' {
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
			if nb := self.peek(); nb == ':' {
				self.next()
				kind = token.SCOPE
			}
		case '@':
			kind = token.AT
		case '?':
			kind = token.QUE
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
