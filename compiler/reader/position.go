package reader

import (
	"fmt"
)

// Position 位置
type Position struct {
	Reader                             Reader
	BeginOffset, EndOffset             int64
	BeginRow, BeginCol, EndRow, EndCol int64
}

// MixPosition 混合两个位置
func MixPosition(begin, end Position) Position {
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
	return fmt.Sprintf("%d:%d", self.BeginRow, self.BeginCol)
}
