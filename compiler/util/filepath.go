package util

import (
	"runtime"

	stlerror "github.com/kkkunny/stl/error"
)

// GetFileName 获取当前文件名
func GetFileName(skip uint) (string, error) {
	_, filename, _, ok := runtime.Caller(int(skip) + 1)
	if !ok {
		return "", stlerror.Errorf("can not get the function name")
	}
	return filename, nil
}
