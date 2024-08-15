package global

import (
	"fmt"

	stlbasic "github.com/kkkunny/stl/basic"

	"github.com/kkkunny/Sim/compiler/hir/types"
)

// TypeDef 类型定义
type TypeDef struct {
	pkgGlobalAttr
	name   string
	target types.Type
}

func (self *TypeDef) String() string {
	return stlbasic.Ternary(self.pkg.Equal(self.pkg.module.BuildinPkg), self.name, fmt.Sprintf("%s::%s", self.pkg.String(), self.name))
}

func (self *TypeDef) Equal(dst types.Type) bool {
	t, ok := dst.(*TypeDef)
	return ok && self.pkg.Equal(t.pkg) && self.name == t.name
}

func (self *TypeDef) Name() string {
	return self.name
}

func (self *TypeDef) Target() types.Type {
	return self.target
}
