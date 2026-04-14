package lex

import (
	"bytes"
	"errors"
	"io"
	"strings"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/token"
)

type Reader interface {
	io.RuneReader
	io.Seeker
}

// Lexer 词法分析器
type Lexer struct {
	reader Reader

	offset   uint
	row, col uint

	cache bytes.Buffer
}

func New(r Reader) *Lexer {
	return &Lexer{
		reader: r,
		row:    1,
		col:    1,
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
			l.col = 1
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

func (l *Lexer) Position() token.Position {
	return token.Position{
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

func (l *Lexer) Scan() token.Token {
	l.skipWhite()
	l.cache.Reset()

	begin := l.Position()
	c := l.next()

	kind := token.KindEnum.Illegal
	switch {
	case c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
		kind = l.scanIdent(c)
	default:
		switch c {
		case 0:
			kind = token.KindEnum.Eof
		case '(':
			kind = token.KindEnum.Lpa
		case ')':
			kind = token.KindEnum.Rpa
		case '{':
			kind = token.KindEnum.Lbr
		case '}':
			kind = token.KindEnum.Rbr
		case ';', '\n':
			kind = token.KindEnum.Sem
		}
	}

	end := l.Position()
	return token.Token{
		Position:   token.MixPosition(begin, end),
		Kind:       kind,
		OriginText: l.cache.String(),
	}
}
