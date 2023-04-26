package mir

import (
	"github.com/kkkunny/containers/list"
)

// Package 包
type Package struct {
	Globals *list.List[Global]
}

func NewPackage() *Package {
	return &Package{Globals: list.NewList[Global]()}
}
