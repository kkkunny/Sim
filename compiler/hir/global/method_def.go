package global

// MethodDef 方法定义
type MethodDef struct {
	pkgGlobalAttr
	FuncDef
	from *CustomTypeDef
}

func (self *MethodDef) From() *CustomTypeDef {
	return self.from
}
