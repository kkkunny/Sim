package reader

import (
	"strings"

	stlos "github.com/kkkunny/stl/os"
)

// NewReaderFromString 从字符串中新建读取器
func NewReaderFromString(path stlos.FilePath, s string) (Reader, error) {
	return NewReaderFromIO(path, strings.NewReader(s))
}
