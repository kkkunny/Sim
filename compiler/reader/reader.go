package reader

import "io"

type Reader interface {
	io.Reader
	io.RuneReader
	io.Seeker
	Path() string
}
