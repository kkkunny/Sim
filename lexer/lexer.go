package lexer

import (
	"errors"
	"io"

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

func (self *Lexer) Scan() (Token, stlerror.Error) {
	begin := self.reader.Position()
	ch, err := self.next()
	if err != nil {
		return Token{}, err
	}

	var kind Kind
	switch ch {
	case 0:
		kind = Eof
	default:
		kind = Illegal
	}

	end := self.reader.Position()
	return Token{
		Position: reader.MixPosition(begin, end),
		Kind:     kind,
	}, nil
}
