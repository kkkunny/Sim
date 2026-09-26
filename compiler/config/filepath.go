package config

import (
	"os"
	"path/filepath"

	stlerr "github.com/kkkunny/stl/error"
)

// SimRootEnvName 显式指定 Sim 根目录（包含 std/ 的目录）的环境变量名。
const SimRootEnvName = "SIM_ROOT"

var (
	WorkPath       = stlerr.MustWith(os.Getwd())
	SimRootPath    = WorkPath
	StdPkgPath     = filepath.Join(SimRootPath, "std")
	BuildinPkgPath = filepath.Join(StdPkgPath, "buildin")
)

func init() {
	// 根路径解析顺序：SIM_ROOT 环境变量 → 可执行文件旁的 std/ → 当前工作目录
	if root := os.Getenv(SimRootEnvName); root != "" {
		SetSimRoot(root)
	} else if root := executableSimRoot(); root != "" {
		SetSimRoot(root)
	}
}

// SetSimRoot 显式设置 Sim 根目录（包含 std/ 的目录），使编译与当前工作目录解耦。
func SetSimRoot(root string) {
	if root == "" {
		return
	}
	absPath := stlerr.MustWith(filepath.Abs(root))
	if info, err := os.Stat(filepath.Join(absPath, "std", "buildin")); err != nil || !info.IsDir() {
		panic(stlerr.Errorf("invalid Sim root %q: std/buildin not found", root))
	}
	SimRootPath = absPath
	StdPkgPath = filepath.Join(absPath, "std")
	BuildinPkgPath = filepath.Join(StdPkgPath, "buildin")
}

// executableSimRoot 返回可执行文件旁的 Sim 根目录（存在 std/buildin 时），
// 使编译产物（编译器二进制 + std/）可整体迁移、不依赖启动目录。
func executableSimRoot() string {
	exePath, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exePath)
	if info, err := os.Stat(filepath.Join(dir, "std", "buildin")); err != nil || !info.IsDir() {
		return ""
	}
	return dir
}

const (
	CacheDir = ".sim_cache"
)

const (
	SourceCodeFileExtName = ".sim"
)
