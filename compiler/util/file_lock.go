package util

type FileLock interface {
	Lock() error
	Unlock() error
}
