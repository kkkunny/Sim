package local

// Break 跳出当前循环
type Break struct {
	loop Loop
}

func NewBreak(loop Loop) *Break {
	return &Break{loop: loop}
}

func (self *Break) local() {
	return
}

func (self *Break) BlockEndType() BlockEndType {
	return BlockEndTypeLoopBreak
}

func (self *Break) Loop() Loop {
	return self.loop
}
