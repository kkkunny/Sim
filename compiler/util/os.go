package util

import (
	"os/exec"

	stlerr "github.com/kkkunny/stl/error"
)

// LookupCCompiler 查找c语言编译器
func LookupCCompiler() (string, error) {
	path, err := stlerr.ErrorWith(exec.LookPath("clang"))
	if err == nil {
		return path, nil
	}
	return stlerr.ErrorWith(exec.LookPath("gcc"))
}
