package util

import (
	"os/exec"

	stlerr "github.com/kkkunny/stl/error"
)

// LookupLinker 查找链接器（clang/gcc用作链接驱动）
func LookupLinker() (string, error) {
	path, err := stlerr.ErrorWith(exec.LookPath("clang"))
	if err == nil {
		return path, nil
	}
	return stlerr.ErrorWith(exec.LookPath("gcc"))
}

// AlignUp 向上对齐
func AlignUp(size, align uint64) uint64 {
	if align == 0 {
		return size
	}
	return (size + align - 1) / align * align
}
