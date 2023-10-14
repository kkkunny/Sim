package lexer

import (
	"errors"
	"io"
	"unicode"

	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/reader"
	. "github.com/kkkunny/Sim/token"
)

// Lexer 词法分析器
type Lexer struct {
	reader reader.Reader // 读取器
}

func New(r reader.Reader) *Lexer {
	return &Lexer{reader: r}
}

// 下一个字符
func (self *Lexer) next() (rune, stlerror.Error) {
	r, _, err := stlerror.ErrorWith2(self.reader.ReadRune())
	if err != nil && errors.Is(err, io.EOF) {
		return 0, nil
	}
	return r, err
}

// 提前获取下一个字符
func (self Lexer) peek() (rune, stlerror.Error) {
	offset := stlerror.MustWith(self.reader.Seek(0, io.SeekCurrent))
	defer self.reader.Seek(offset, io.SeekStart)
	return self.next()
}

// 扫描标识符
func (self *Lexer) scanIdent() (Kind, stlerror.Error) {
	for ch, err := self.peek(); ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch); ch, err = self.peek() {
		if err != nil {
			return ILLEGAL, err
		}
		_, err = self.next()
		if err != nil {
			return ILLEGAL, err
		}
	}
	return IDENT, nil
}

// 跳过空白
func (self *Lexer) skipWhite() stlerror.Error {
	for ch, err := self.peek(); ch == ' '; ch, err = self.peek() {
		if err != nil {
			return err
		}
		_, err = self.next()
		if err != nil {
			return err
		}
	}
	return nil
}

func (self *Lexer) Scan() (Token, stlerror.Error) {
	if err := self.skipWhite(); err != nil {
		return Token{}, err
	}

	begin := self.reader.Position()
	ch, err := self.next()
	if err != nil {
		return Token{}, err
	}

	var kind Kind
	switch {
	case ch == '_' || unicode.IsLetter(ch):
		kind, err = self.scanIdent()
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
		default:
			kind = ILLEGAL
		}
	}

	if err != nil {
		return Token{}, err
	}

	end := self.reader.Position()
	return Token{
		Position: reader.MixPosition(begin, end),
		Kind:     kind,
	}, nil
}
