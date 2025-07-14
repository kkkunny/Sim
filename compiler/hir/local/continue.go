package local

// Continue 跳过当前循环
type Continue struct {
	loop Loop
}

func NewContinue(loop Loop) *Continue {
	return &Continue{loop: loop}
}

func (self *Continue) Local() {
	return
}

func (self *Continue) BlockEndType() BlockEndType {
	return BlockEndTypeLoopNext
}

func (self *Continue) Loop() Loop {
	return self.loop
}
