package util

import "github.com/kkkunny/go-llvm"

// InitLLVM 初始化LLVM
func InitLLVM() {
	_ = Logger.Infof(0, "LLVM VERSION: %s", llvm.Version)
}
