package mir

// Context 上下文
type Context struct {
	target Target
}

func (self Context) Target()Target{
	return self.target
}
