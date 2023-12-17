package reader

import (
	"strings"

	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
)

// NewReaderFromString 从字符串中新建读取器
func NewReaderFromString(path stlos.FilePath, s string) (Reader, stlerror.Error) {
	return NewReaderFromIO(path, strings.NewReader(s))
}
