package token

import (
	"fmt"
)

// Position 位置
type Position struct {
	BeginOffset, EndOffset             uint
	BeginRow, BeginCol, EndRow, EndCol uint
}

// MixPosition 混合两个位置
func MixPosition(begin, end Position) Position {
	return Position{
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
