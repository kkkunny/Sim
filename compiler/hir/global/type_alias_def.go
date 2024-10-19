package global

import (
	"fmt"

	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir/types"
)

// TypeAliasDef 类型别名定义
type TypeAliasDef struct {
	pkgGlobalAttr
	name   string
	target types.Type
}

func (self *TypeAliasDef) String() string {
	return stlval.Ternary(self.pkg.Equal(self.pkg.module.BuildinPkg), self.name, fmt.Sprintf("%s::%s", self.pkg.String(), self.name))
}

func (self *TypeAliasDef) Equal(dst types.Type) bool {
	t, ok := dst.(*TypeAliasDef)
	if ok {
		return self.pkg.Equal(t.pkg) && self.name == t.name
	} else {
		return self.target.Equal(dst)
	}
}

func (self *TypeAliasDef) Name() string {
	return self.name
}

func (self *TypeAliasDef) Target() types.Type {
	return self.target
}
