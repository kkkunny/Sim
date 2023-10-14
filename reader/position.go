package reader

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
