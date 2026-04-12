package reader

import (
	"io"
)

// Reader 读取器
type Reader interface {
	io.ByteReader
	io.Seeker
	Path() string
	Position() Position
	Offset() uint
}
