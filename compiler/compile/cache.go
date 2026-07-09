package compile

import (
	"os"
	"path/filepath"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/config"
	"github.com/kkkunny/Sim/compiler/hir/globals"
	"github.com/kkkunny/Sim/compiler/util"
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
	err := stlerr.ErrorWrap(l.f.Close())
	if err != nil {
		return err
	}
	return l.locker.Unlock()
}

// isCacheValid 判断包的编译缓存是否有效（基于 mtime）
func isCacheValid(pkg *globals.Package) (bool, error) {
	cacheDir := filepath.Join(pkg.Path, config.CacheDir)
	if _, err := os.Stat(cacheDir); err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	objPath := filepath.Join(cacheDir, pkg.Name+".o")
	hdrPath := filepath.Join(cacheDir, pkg.Name+".h")

	objInfo, err := stlerr.ErrorWith(os.Stat(objPath))
	if err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	objMtime := objInfo.ModTime()
	hdrInfo, err := stlerr.ErrorWith(os.Stat(hdrPath))
	if err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	hdrMtime := hdrInfo.ModTime()

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
		if info.ModTime().After(objMtime) || info.ModTime().After(hdrMtime) {
			return false, nil
		}
	}
	return true, nil
}
