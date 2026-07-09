package config

import (
	"os"
	"path/filepath"

	stlerr "github.com/kkkunny/stl/error"
)

var (
	WorkPath    = stlerr.MustWith(os.Getwd())
	SimRootPath = WorkPath

	StdPkgPath     = filepath.Join(SimRootPath, "std")
	BuildinPkgPath = filepath.Join(StdPkgPath, "buildin")
	CIncludePath   = filepath.Join(SimRootPath, "include")
)

const (
	CacheDir = ".sim_cache"
)

const (
	SourceCodeFileExtName = ".sim"
)
