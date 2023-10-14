package reader

import (
	"errors"
	"io"
	"os"
	"unicode/utf8"

	stlerror "github.com/kkkunny/stl/error"
)

// 文件读取器
type _FileReader struct {
	path    string
	file    *os.File
	rowLens []uint
}

// NewReaderFromFile 从文件中新建读取器
func NewReaderFromFile(path string) (Reader, stlerror.Error) {
	file, err := stlerror.ErrorWith(os.Open(path))
	if err != nil {
		return nil, err
	}

	var rowlens []uint
	var rowLen uint
	for {
		tmp := make([]byte, 1)
		_, e := file.Read(tmp)
		if e != nil && !errors.Is(e, io.EOF) {
			return nil, stlerror.ErrorWrap(e)
		}

		if e != nil {
			rowlens = append(rowlens, rowLen)
			break
		} else if tmp[0] == '\n' {
			rowLen++
			rowlens = append(rowlens, rowLen)
			rowLen = 0
		} else {
			rowLen++
		}
	}

	_, err = stlerror.ErrorWith(file.Seek(0, io.SeekStart))
	if err != nil {
		return nil, err
	}

	return &_FileReader{
		path:    path,
		file:    file,
		rowLens: rowlens,
	}, nil
}

func (self _FileReader) Path() string {
	return self.path
}

func (self *_FileReader) ReadRune() (rune, int, error) {
	var buf []byte
	for {
		tmp := make([]byte, 1)
		_, err := self.file.Read(tmp)
		if err != nil {
			return utf8.RuneError, len(buf), err
		}
		buf = append(buf, tmp...)

		if utf8.FullRune(buf) {
			r, size := utf8.DecodeRune(buf)
			return r, size, nil
		}
	}
}

func (self *_FileReader) Seek(offset int64, whence int) (int64, error) {
	return self.file.Seek(offset, whence)
}

func (self *_FileReader) Position() Position {
	offset := stlerror.MustWith(self.file.Seek(0, io.SeekCurrent))

	cursor := offset
	var curRowOffset, curColOffset uint
	for rowOffset, rowLen := range self.rowLens {
		if cursor-int64(rowLen) < 0 {
			curRowOffset = uint(rowOffset)
			curColOffset = uint(cursor)
			break
		} else {
			cursor -= int64(rowLen)
		}
	}

	return Position{
		Reader:      self,
		BeginOffset: uint(offset),
		EndOffset:   uint(offset),
		BeginRow:    curRowOffset + 1,
		BeginCol:    curColOffset + 1,
		EndRow:      curRowOffset + 1,
		EndCol:      curColOffset + 1,
	}
}
