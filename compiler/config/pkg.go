package config

import (
	stlerror "github.com/kkkunny/stl/error"
	stlos "github.com/kkkunny/stl/os"
)

var (
	// RootPath 语言根目录
	RootPath = func() stlos.FilePath {
		workdir := stlerror.MustWith(stlos.GetWorkDirectory())
		if workdir.Base() == "compiler" {
			workdir = workdir.Dir()
		}
		return stlerror.MustWith(workdir.Abs())
	}()
	// OfficialPkgPath 官方包目录
	OfficialPkgPath = RootPath
	// StdPkgPath 标准库目录
	StdPkgPath = OfficialPkgPath.Join("std")
	// BuildInPkgPath buildin库目录
	BuildInPkgPath = StdPkgPath.Join("buildin")
)
