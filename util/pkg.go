package util

import (
	"path/filepath"

	"github.com/kkkunny/Sim/config"
)

// GetBuildInPackagePath 获取buildin包的地址
func GetBuildInPackagePath() string {
	return filepath.Join(config.ROOT, "std", "buildin")
}
