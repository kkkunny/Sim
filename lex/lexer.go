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
func (self *Lexer) scanIdent(ch rune) Kind {
	var buf strings.Builder
	buf.WriteRune(ch)
	for ch = self.peek(); ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9'); ch = self.peek() {
		buf.WriteRune(self.next())
	}
	return Lookup(buf.String())
}

// 扫描整数
func (self *Lexer) scanNumber(ch rune) Kind {
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
		return FLOAT
	} else {
		return INTEGER
	}
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
		kind = self.scanNumber(ch)
	default:
		switch ch {
		case 0:
			kind = EOF
		case '=':
			kind = ASS
			if nextCh := self.peek(); nextCh == '=' {
				self.next()
				kind = EQ
			}
		case '&':
			kind = AND
		case '|':
			kind = OR
		case '^':
			kind = XOR
		case '!':
			kind = NOT
			if nextCh := self.peek(); nextCh == '=' {
				self.next()
				kind = NE
			}
		case '+':
			kind = ADD
		case '-':
			kind = SUB
		case '*':
			kind = MUL
		case '/':
			kind = DIV
		case '%':
			kind = REM
		case '<':
			kind = LT
			if nextCh := self.peek(); nextCh == '=' {
				self.next()
				kind = LE
			}
		case '>':
			kind = GT
			if nextCh := self.peek(); nextCh == '=' {
				self.next()
				kind = GE
			}
		case '(':
			kind = LPA
		case ')':
			kind = RPA
		case '[':
			kind = LBA
		case ']':
			kind = RBA
		case '{':
			kind = LBR
		case '}':
			kind = RBR
		case ';', '\n':
			kind = SEM
		case ',':
			kind = COM
		case '.':
			kind = DOT
		case ':':
			kind = COL
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
