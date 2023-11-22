package reader

import (
	"strings"

	stlerror "github.com/kkkunny/stl/error"
)

// NewReaderFromString 从字符串中新建读取器
func NewReaderFromString(path string, s string) (Reader, stlerror.Error) {
	return NewReaderFromIO(path, strings.NewReader(s))
}
