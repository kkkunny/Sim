package local

import (
	"github.com/kkkunny/stl/list"
)

// Break 跳出当前循环
type Break struct {
	pos *list.Element[Local]
}

func NewBreak() *Continue {
	return &Continue{}
}

func (self *Break) setPosition(pos *list.Element[Local]) {
	self.pos = pos
}

func (self *Break) position() (*list.Element[Local], bool) {
	return self.pos, self.pos != nil
}

func (self *Break) BlockEndType() BlockEndType {
	return BlockEndTypeLoopBreak
}
