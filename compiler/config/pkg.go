package config

import (
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
)

// ROOT 语言根目录
var ROOT = func() stlos.FilePath {
	workdir := stlerror.MustWith(stlos.GetWorkDirectory())
	if workdir.Base() == "compiler"{
		workdir = workdir.Dir()
	}
	return workdir
}()
