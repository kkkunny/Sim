package local

type BlockEndType uint8

const (
	BlockEndTypeNone BlockEndType = iota
	BlockEndTypeLoopNext
	BlockEndTypeLoopBreak
	BlockEndTypeFuncRet
)

type blockEnd interface {
	Local
	BlockEndType() BlockEndType
}
