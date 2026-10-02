package compile

import (
	"os"
	"path/filepath"

	stlerr "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/compiler/config"
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
	err := l.locker.Unlock()
	if err != nil {
		return err
	}
	return stlerr.ErrorWrap(l.f.Close())
}

// isCacheValid 判断包的编译缓存是否有效（基于 mtime）
func isCacheValid(pkgDir, pkgName string) (bool, error) {
	cacheDir := filepath.Join(pkgDir, config.CacheDir)
	if _, err := os.Stat(cacheDir); err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, stlerr.ErrorWrap(err)
	}

	backend, err := os.ReadFile(filepath.Join(cacheDir, ".backend"))
	if err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, stlerr.ErrorWrap(err)
	}
	if string(backend) != config.BackendVersion {
		return false, nil
	}

	objPath := filepath.Join(cacheDir, pkgName+".o")
	objInfo, err := os.Stat(objPath)
	if err != nil && os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, stlerr.ErrorWrap(err)
	}
	objMtime := objInfo.ModTime()

	entries, err := stlerr.ErrorWith(os.ReadDir(pkgDir))
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
	return true, nil
}

// writeCacheBackend 写入后端版本标记
func writeCacheBackend(cacheDir string) error {
	path := filepath.Join(cacheDir, ".backend")
	return stlerr.ErrorWrap(os.WriteFile(path, []byte(config.BackendVersion), 0644))
}
