package config

import (
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
)

// ROOT 语言根目录
var ROOT stlos.FilePath = stlerror.MustWith(stlos.GetWorkDirectory()).Dir()
