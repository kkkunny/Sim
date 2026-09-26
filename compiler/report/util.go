package report

import (
	"errors"
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
	// io.Reader.Read 允许短读，必须用 ReadFull 否则诊断源码可能被截断
	n, err := io.ReadFull(r, data)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", err
	}
	return string(data[:n]), nil
}
