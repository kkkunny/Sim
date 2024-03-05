package hir

import (
	"fmt"

	stlbasic "github.com/kkkunny/stl/basic"
)

func GetGlobalName(pkg Package, name string) string {
	return stlbasic.Ternary(pkg.Equal(BuildInPackage), name, fmt.Sprintf("%s::%s", pkg, name))
}
