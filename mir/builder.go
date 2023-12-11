package mir

// Builder 构造器
type Builder struct {
	ctx *Context
	cur *Block
}

func (self *Context) NewBuilder()*Builder{
	return &Builder{ctx: self}
}

func (self *Builder) Current()*Block{
	return self.cur
}

func (self *Builder) MoveTo(block *Block){
	self.cur = block
}
