package utils

import (
	"os"
	"path/filepath"
	"unsafe"

	stlos "github.com/kkkunny/stl/os"
	"golang.org/x/exp/constraints"
)

// PtrByte 指针大小
var PtrByte = uint(unsafe.Sizeof(uintptr(0)))

// AlignByte 对齐
var AlignByte = 4

// AlignTo 对齐
func AlignTo[T constraints.Integer | constraints.Float](n, align T) T {
	return (n + align - 1) / align * align
}

var StdPath = func() stlos.Path {
	exe, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return stlos.Path(filepath.Dir(exe))
}()
