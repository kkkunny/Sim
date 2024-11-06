package local

// Break 跳出当前循环
type Break struct{}

func NewBreak() *Break {
	return &Break{}
}

func (self *Break) local() {
	return
}

func (self *Break) BlockEndType() BlockEndType {
	return BlockEndTypeLoopBreak
}
