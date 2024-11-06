package local

// Continue 跳过当前循环
type Continue struct{}

func NewContinue() *Continue {
	return &Continue{}
}

func (self *Continue) local() {
	return
}

func (self *Continue) BlockEndType() BlockEndType {
	return BlockEndTypeLoopNext
}
