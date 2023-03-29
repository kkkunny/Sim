package mir

import (
	"strings"

	"github.com/bahlo/generic-list-go"
)

// Package 包
type Package struct {
	Globals *list.List[Global]
}

func NewPackage() *Package {
	return &Package{Globals: list.New[Global]()}
}

func (self Package) String() string {
	var buf strings.Builder
	for cursor := self.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		buf.WriteString(cursor.Value.String())
		if cursor.Next() != nil {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}
