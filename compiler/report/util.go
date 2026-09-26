package report

import (
	"io"

	"github.com/kkkunny/Sim/compiler/reader"
)

func ReadFromTo(r reader.Reader, from, to int64) (string, error) {
	if from < 0 {
		from = 0
	}
	if to < from {
		return "", nil
	}
	_, err := r.Seek(from, io.SeekStart)
	if err != nil {
		return "", err
	}
	data := make([]byte, to-from)
	n, err := r.Read(data)
	if err != nil {
		return "", err
	}
	return string(data[:n]), nil
}
