package local

import "github.com/kkkunny/Sim/compiler/hir"

type BlockEndType uint8

const (
	BlockEndTypeNone BlockEndType = iota
	BlockEndTypeLoopNext
	BlockEndTypeLoopBreak
	BlockEndTypeFuncRet
)

type blockEnd interface {
	hir.Local
	BlockEndType() BlockEndType
}
