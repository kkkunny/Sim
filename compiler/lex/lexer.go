package lex

import (
	"bytes"
	"errors"
	"io"
	"strings"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/reader"
	"github.com/kkkunny/Sim/compiler/token"
)

// Lexer 词法分析器
type Lexer struct {
	reader reader.Reader

	offset   int64
	row, col int64

	cache bytes.Buffer
}

func New(r reader.Reader) *Lexer {
	return &Lexer{
		reader: r,
		offset: -1,
		row:    1,
	}
}

// 下一个字符
func (l *Lexer) next() rune {
	c, _, err := stlerror.ErrorWith2(l.reader.ReadRune())
	if err != nil && !errors.Is(err, io.EOF) {
		panic(err)
	}
	if err == nil {
		l.cache.WriteRune(c)
		l.offset++
		l.col++
		if c == '\n' {
			l.row++
			l.col = 0
		}
	}
	return c
}

// 提前获取下一个字符
func (l *Lexer) peek(skip ...uint) rune {
	offset := 1
	for _, i := range skip {
		offset += int(i)
	}

	var c rune
	var s int
	for i := 0; i < offset; i++ {
		cc, size, err := stlerror.ErrorWith2(l.reader.ReadRune())
		if err != nil && !errors.Is(err, io.EOF) {
			panic(err)
		}
		c = cc
		s += size
	}

	stlerror.MustWith(l.reader.Seek(-int64(s), io.SeekCurrent))

	return c
}

func (l *Lexer) Position() reader.Position {
	return reader.Position{
		Reader:      l.reader,
		BeginOffset: l.offset,
		EndOffset:   l.offset,
		BeginRow:    l.row,
		BeginCol:    l.col,
		EndRow:      l.row,
		EndCol:      l.col,
	}
}

// 跳过空白
func (l *Lexer) skipWhite() {
	for c := l.peek(); c == ' ' || c == '\r' || c == '\t'; c = l.peek() {
		l.next()
	}
}

// 扫描标识符
func (l *Lexer) scanIdent(c rune) token.Kind {
	var buf strings.Builder
	buf.WriteRune(c)
	for c = l.peek(); c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9'); c = l.peek() {
		buf.WriteRune(l.next())
	}
	return token.Lookup(buf.String())
}

func (l *Lexer) scanInteger(c rune) token.Kind {
	var buf strings.Builder
	buf.WriteRune(c)
	for c = l.peek(); c >= '0' && c <= '9'; c = l.peek() {
		buf.WriteRune(l.next())
	}
	return token.KindEnum.Integer
}

func (l *Lexer) scanString() token.Kind {
	for c := l.peek(); c != '"'; c = l.peek() {
		if c == '\\' {
			l.next()
		}
		l.next()
	}
	l.next()
	return token.KindEnum.String
}

func (l *Lexer) Scan() token.Token {
	l.skipWhite()
	l.cache.Reset()

	c := l.next()
	begin := l.Position()

	var kind token.Kind
	switch {
	case c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
		kind = l.scanIdent(c)
	case c >= '0' && c <= '9':
		kind = l.scanInteger(c)
	case c == '"':
		kind = l.scanString()
	default:
		switch c {
		case 0:
			kind = token.KindEnum.Eof
		case '=':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.Eq
			default:
				kind = token.KindEnum.Assign
			}
		case '(':
			kind = token.KindEnum.Lpa
		case ')':
			kind = token.KindEnum.Rpa
		case '[':
			kind = token.KindEnum.Lba
		case ']':
			kind = token.KindEnum.Rba
		case '{':
			kind = token.KindEnum.Lbr
		case '}':
			kind = token.KindEnum.Rbr
		case ',':
			kind = token.KindEnum.Comma
		case ':':
			switch l.peek() {
			case ':':
				l.next()
				kind = token.KindEnum.Scope
			default:
				kind = token.KindEnum.Col
			}
		case ';':
			kind = token.KindEnum.Sem
		case '\n':
			kind = token.KindEnum.Br
		case '-':
			switch l.peek() {
			case '>':
				l.next()
				kind = token.KindEnum.Arrow
			case '=':
				l.next()
				kind = token.KindEnum.SubAssign
			default:
				kind = token.KindEnum.Sub
			}
		case '+':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.AddAssign
			default:
				kind = token.KindEnum.Add
			}
		case '*':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.MulAssign
			default:
				kind = token.KindEnum.Mul
			}
		case '/':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.QuoAssign
			default:
				kind = token.KindEnum.Quo
			}
		case '%':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.RemAssign
			default:
				kind = token.KindEnum.Rem
			}
		case '&':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.AndAssign
			case '&':
				l.next()
				kind = token.KindEnum.LogicAnd
			default:
				kind = token.KindEnum.And
			}
		case '|':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.OrAssign
			case '|':
				l.next()
				kind = token.KindEnum.LogicOr
			default:
				kind = token.KindEnum.Or
			}
		case '^':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.XorAssign
			default:
				kind = token.KindEnum.Xor
			}
		case '!':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.Neq
			default:
				kind = token.KindEnum.Not
			}
		case '?':
			kind = token.KindEnum.Question
		case '@':
			kind = token.KindEnum.At
		case '<':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.Lte
			case '<':
				l.next()
				switch l.peek() {
				case '=':
					l.next()
					kind = token.KindEnum.ShlAssign
				default:
					kind = token.KindEnum.Shl
				}
			default:
				kind = token.KindEnum.Lt
			}
		case '>':
			switch l.peek() {
			case '=':
				l.next()
				kind = token.KindEnum.Gte
			case '>':
				l.next()
				switch l.peek() {
				case '=':
					l.next()
					kind = token.KindEnum.ShrAssign
				default:
					kind = token.KindEnum.Shr
				}
			default:
				kind = token.KindEnum.Gt
			}
		}
	}

	end := l.Position()
	return token.Token{
		Position:   reader.MixPosition(begin, end),
		Kind:       kind,
		OriginText: l.cache.String(),
	}
}
