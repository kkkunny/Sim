package reader

import "io"

// Reader 读取器
type Reader interface {
	io.RuneReader
	io.Seeker
	Path() string
	Position() Position
}
