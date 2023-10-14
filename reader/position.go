package reader

import (
	"fmt"
	"io"
	"strings"

	stlerror "github.com/kkkunny/stl/error"
)

// Position 位置
type Position struct {
	Reader                             Reader // 读取器
	BeginOffset, EndOffset             uint
	BeginRow, BeginCol, EndRow, EndCol uint
}

// MixPosition 混合两个位置
func MixPosition(begin, end Position) Position {
	if begin.Reader != end.Reader {
		panic("unreachable")
	}
	return Position{
		Reader:      begin.Reader,
		BeginOffset: begin.BeginOffset,
		EndOffset:   end.EndOffset,
		BeginRow:    begin.BeginRow,
		BeginCol:    begin.BeginCol,
		EndRow:      end.EndRow,
		EndCol:      end.EndCol,
	}
}

func (self Position) String() string {
	return fmt.Sprintf("%s[%d:%d]", self.Reader.Path(), self.BeginOffset, self.EndOffset)
}

func (self Position) Text() string {
	offset := stlerror.MustWith(self.Reader.Seek(0, io.SeekCurrent))
	defer self.Reader.Seek(offset, io.SeekStart)

	stlerror.MustWith(self.Reader.Seek(int64(self.BeginOffset), io.SeekStart))

	var buf strings.Builder
	for i := 0; i < int(self.EndOffset-self.BeginOffset); i++ {
		r, _ := stlerror.MustWith2(self.Reader.ReadRune())
		buf.WriteRune(r)
	}
	return buf.String()
}
