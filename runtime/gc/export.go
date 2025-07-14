package gc

import "unsafe"

var globalGC *GarbageCollector

// Init 初始化gc
func Init(stackBegin uintptr) (err error) {
	globalGC, err = newGarbageCollector(stackBegin)
	if err != nil {
		return err
	}
	return nil
}

// Alloca 分配内存
func Alloca(size uint, stackPos uintptr) (unsafe.Pointer, error) {
	return globalGC.allocate(size, stackPos)
}

// GC 立即进行gc
func GC(stackPos uintptr) error {
	return globalGC.gc(stackPos)
}
