package local

import (
	"github.com/kkkunny/stl/list"
)

// Continue 跳过当前循环
type Continue struct {
	pos *list.Element[Local]
}

func NewContinue() *Continue {
	return &Continue{}
}

func (self *Continue) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *Continue) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
}

func (self *Continue) BlockEndType() BlockEndType {
	return BlockEndTypeLoopNext
}
