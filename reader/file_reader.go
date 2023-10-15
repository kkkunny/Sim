package reader

import (
	"os"

	stlerror "github.com/kkkunny/stl/error"
)

// NewReaderFromFile 从文件中新建读取器
func NewReaderFromFile(path string) (*os.File, Reader, stlerror.Error) {
	file, err := stlerror.ErrorWith(os.Open(path))
	if err != nil {
		return nil, nil, err
	}
	self, err := NewReaderFromIO(path, file)
	return file, self, err
}
