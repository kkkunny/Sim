package mir

// Context 上下文
type Context struct {
	target Target
}

func NewContext(target Target)*Context{
	return &Context{target: target}
}

func (self Context) Target() Target {
	return self.target
}
