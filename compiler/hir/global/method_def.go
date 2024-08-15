package global

// MethodDef 方法定义
type MethodDef struct {
	pkgGlobalAttr
	FuncDef
	from *TypeDef
}

func (self *MethodDef) From() *TypeDef {
	return self.from
}
