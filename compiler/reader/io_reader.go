package reader

import (
	"errors"
	"io"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
)

// IO读取器
type _IOReader struct {
	path    stlos.FilePath
	reader  io.ReadSeeker
	rowLens []uint
}

// NewReaderFromIO 从io中新建读取器
func NewReaderFromIO(path stlos.FilePath, reader io.ReadSeeker) (Reader, stlerror.Error) {
	var rowlens []uint
	var rowLen uint
	for {
		tmp := make([]byte, 1)
		_, e := reader.Read(tmp)
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

	stlerror.MustWith(reader.Seek(0, io.SeekStart))

	return &_IOReader{
		path:    path,
		reader:  reader,
		rowLens: rowlens,
	}, nil
}

func (self _IOReader) Path() stlos.FilePath {
	return self.path
}

func (self *_IOReader) ReadByte() (byte, error) {
	tmp := make([]byte, 1)
	_, err := self.reader.Read(tmp)
	if err != nil {
		return 0, err
	}
	return tmp[0], nil
}

func (self *_IOReader) Seek(offset int64, whence int) (int64, error) {
	return self.reader.Seek(offset, whence)
}

func (self *_IOReader) Offset() uint {
	return uint(stlerror.MustWith(self.Seek(0, io.SeekCurrent)))
}

func (self *_IOReader) Position() Position {
	offset := self.Offset()

	cursor := offset
	var curRowOffset, curColOffset uint
	for rowOffset, rowLen := range self.rowLens {
		if int64(cursor)-int64(rowLen) < 0 {
			curRowOffset = uint(rowOffset)
			curColOffset = cursor
			break
		} else {
			cursor -= rowLen
		}
	}

	return Position{
		Reader:      self,
		BeginOffset: offset,
		EndOffset:   offset,
		BeginRow:    curRowOffset + 1,
		BeginCol:    curColOffset + 1,
		EndRow:      curRowOffset + 1,
		EndCol:      curColOffset + 1,
	}
}
