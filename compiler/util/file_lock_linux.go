package util

import (
	"os"

	stlerr "github.com/kkkunny/stl/error"
	"golang.org/x/sys/unix"
)

type fileLock struct {
	f *os.File
}

func NewFileLock(f *os.File) FileLock {
	return &fileLock{f: f}
}

// Lock 获取排他锁（阻塞）
func (fl *fileLock) Lock() error {
	return stlerr.ErrorWrap(unix.Flock(int(fl.f.Fd()), unix.LOCK_EX))
}

func (fl *fileLock) Unlock() error {
	return stlerr.ErrorWrap(unix.Flock(int(fl.f.Fd()), unix.LOCK_UN))
}
