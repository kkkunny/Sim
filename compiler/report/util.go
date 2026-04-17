package report

import (
	"io"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/reader"
)

func ReadFromTo(r reader.Reader, from, to int64) (string, error) {
	_, err := r.Seek(from, io.SeekStart)
	if err != nil {
		return "", err
	}
	data := make([]byte, to-from)
	_, err = stlerr.ErrorWith(r.Read(data))
	if err != nil {
		return "", err
	}
	return string(data), nil
}
