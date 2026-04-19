package config

import (
	"os"
	"path/filepath"

	stlerr "github.com/kkkunny/stl/error"
)

var (
	WorkPath    = stlerr.MustWith(os.Getwd())
	SimRootPath = WorkPath
	IncludePath = filepath.Join(WorkPath, "include")
)
