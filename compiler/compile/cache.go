package compile

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/util"
)

// 后端标识：避免与旧 C 后端产物混用（缓存目录内文件名）
const (
	backendMarkerFile = ".backend"
	backendMarker     = "llvm-" + config.ABIVersion
)

type cacheLock struct {
	f      *os.File
	locker util.FileLock
}

func newCacheLock(pkgDir string) (*cacheLock, error) {
	cachePath := filepath.Join(pkgDir, config.CacheDir, ".lock")
	f, err := stlerr.ErrorWith(os.OpenFile(cachePath, os.O_CREATE|os.O_RDWR, 0644))
	if err != nil {
		return nil, err
	}
	return &cacheLock{
		f:      f,
		locker: util.NewFileLock(f),
	}, nil
}

func (l *cacheLock) Lock() error {
	return l.locker.Lock()
}

func (l *cacheLock) Close() error {
	// 必须先解锁再关闭 fd：对已关闭的 fd 执行 flock(LOCK_UN) 会返回 EBADF，
	// 错误还会被 defer 吞掉。
	unlockErr := stlerr.ErrorWrap(l.locker.Unlock())
	closeErr := stlerr.ErrorWrap(l.f.Close())
	return errors.Join(unlockErr, closeErr)
}

func writeBackendMarker(cacheDir string) error {
	return stlerr.ErrorWrap(os.WriteFile(filepath.Join(cacheDir, backendMarkerFile), []byte(backendMarker), 0644))
}

// isCacheValid 判断包的编译缓存是否有效（基于 mtime + 后端标识）。
// 依赖包的产物也必须不新于本包产物：依赖重编（Topo 序遍历中先于本包发生）会使其 .o
// mtime 变新，从而级联判定本包缓存失效；否则依赖的接口/布局变化会被旧 .o 静默沿用。
func isCacheValid(pkg *globals.Package) (bool, error) {
	cacheDir := filepath.Join(pkg.Path, config.CacheDir)
	if _, err := os.Stat(cacheDir); errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	objPath := filepath.Join(cacheDir, pkg.Name+".o")
	objInfo, err := os.Stat(objPath)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	objMtime := objInfo.ModTime()

	marker, err := os.ReadFile(filepath.Join(cacheDir, backendMarkerFile))
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if string(marker) != backendMarker {
		return false, nil
	}

	entries, err := stlerr.ErrorWith(os.ReadDir(pkg.Path))
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != config.SourceCodeFileExtName {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return false, err
		}
		if info.ModTime().After(objMtime) {
			return false, nil
		}
	}

	// 依赖产物必须不新于本包产物（依赖重编会级联失效本包缓存）
	for _, dep := range pkg.Dependencies {
		depObjPath := filepath.Join(dep.Path, config.CacheDir, dep.Name+".o")
		depInfo, err := os.Stat(depObjPath)
		if errors.Is(err, fs.ErrNotExist) {
			return false, nil
		} else if err != nil {
			return false, err
		}
		if depInfo.ModTime().After(objMtime) {
			return false, nil
		}
	}
	return true, nil
}
