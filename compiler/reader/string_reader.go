package reader

import (
	"strings"
)

// NewReaderFromString 从字符串中新建读取器
func NewReaderFromString(path string, s string) (Reader, error) {
	return NewReaderFromIO(path, strings.NewReader(s))
}
