package reader

import (
	"io"

	stlos "github.com/kkkunny/stl/os"
)

// Reader 读取器
type Reader interface {
	io.RuneReader
	io.Seeker
	Path() stlos.FilePath
	Position() Position
	Offset() uint
}
