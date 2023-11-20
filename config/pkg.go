package config

import (
	"os"

	stlerror "github.com/kkkunny/stl/error"
)

// ROOT 语言根目录
var ROOT string = func() string {
	// FIXME: 从环境变量或程序目录获取语言根目录
	return stlerror.MustWith(os.Getwd())
}()
