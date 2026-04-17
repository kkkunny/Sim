package reader

import (
	"errors"
	"io"
	"unicode/utf8"

	stlerr "github.com/kkkunny/stl/error"
)

type fileReader struct {
	path string
	r    io.ReadSeeker
}

func NewFile(path string, reader io.ReadSeeker) Reader {
	return &fileReader{
		path: path,
		r:    reader,
	}
}

func (r *fileReader) Read(p []byte) (n int, err error) {
	return stlerr.ErrorWith(r.r.Read(p))
}

func (r *fileReader) ReadRune() (rune, int, error) {
	var data []byte
	for {
		var temp [1]byte
		_, err := r.Read(temp[:])
		if err != nil && !errors.Is(err, io.EOF) {
			return 0, 0, err
		} else if err != nil {
			break
		}
		data = append(data, temp[0])
		if utf8.Valid(data) {
			break
		}
	}
	if len(data) == 0 {
		return 0, 0, io.EOF
	}
	c, size := utf8.DecodeRune(data)
	if c == utf8.RuneError {
		return 0, 0, stlerr.Errorf("invalid UTF-8")
	}
	return c, size, nil
}

func (r *fileReader) Seek(offset int64, whence int) (int64, error) {
	return stlerr.ErrorWith(r.r.Seek(offset, whence))
}

func (r *fileReader) Path() string {
	return r.path
}
