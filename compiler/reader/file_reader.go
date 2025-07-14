package reader

import (
	"io"
	"os"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
)

// NewReaderFromFile 从文件中新建读取器
func NewReaderFromFile(path stlos.FilePath) (io.Closer, Reader, error) {
	file, err := stlerror.ErrorWith(os.Open(string(path)))
	if err != nil {
		return nil, nil, err
	}
	self, err := NewReaderFromIO(path, file)
	return file, self, err
}
