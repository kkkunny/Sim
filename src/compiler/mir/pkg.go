package mir

import (
	"strings"

	"github.com/kkkunny/containers/list"
)

// Package 包
type Package struct {
	Globals *list.List[Global]
}

func NewPackage() *Package {
	return &Package{Globals: list.NewList[Global]()}
}

func (self Package) String() string {
	var buf strings.Builder
	for cursor := self.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		buf.WriteString(cursor.Value().String())
		if cursor.Next() != nil {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}
