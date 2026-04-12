package config

import (
	"os"
	"path/filepath"

	stlerror "github.com/kkkunny/stl/error"
)

var (
	// RootPath 语言根目录
	RootPath = func() string {
		workdir := stlerror.MustWith(os.Getwd())
		if filepath.Base(workdir) == "compiler" {
			workdir = filepath.Dir(workdir)
		}
		return stlerror.MustWith(filepath.Abs(workdir))
	}()
	// OfficialPkgPath 官方包目录
	OfficialPkgPath = RootPath
	// StdPkgPath 标准库目录
	StdPkgPath = filepath.Join(OfficialPkgPath, "std")
	// BuildInPkgPath buildin库目录
	BuildInPkgPath = filepath.Join(StdPkgPath, "buildin")
)
