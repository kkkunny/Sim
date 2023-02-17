package hir

import (
	"path/filepath"
	"strings"

	"github.com/kkkunny/stl/list"
	stlos "github.com/kkkunny/stl/os"
)

// Package 包
type Package struct {
	Globals *list.SingleLinkedList[Global]
}

// PkgPath 包地址
type PkgPath struct {
	Path stlos.Path // 地址
}

// NewPkgPath 新建包地址
func NewPkgPath(path stlos.Path) PkgPath {
	return PkgPath{
		Path: path,
	}
}

func (self PkgPath) Equal(dst PkgPath) bool {
	return self.Path == dst.Path
}

func (self PkgPath) String() string {
	dirs := strings.Split(self.Path.String(), string([]rune{filepath.Separator}))
	return strings.Join(dirs, ".")
}
