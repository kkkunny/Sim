package util

import (
	"runtime"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/tuple"
	stlerr "github.com/kkkunny/stl/error"
)

func GetLLVMTarget() (llvm.Target, error) {
	switch tuple.Pack2(runtime.GOOS, runtime.GOARCH) {
	case tuple.Pack2("windows", "amd64"):
		err := stlerr.ErrorWrap(llvm.InitializeTargetInfo(llvm.X86))
		if err != nil {
			return llvm.Target{}, err
		}
		err = stlerr.ErrorWrap(llvm.InitializeTarget(llvm.X86))
		if err != nil {
			return llvm.Target{}, err
		}
		err = stlerr.ErrorWrap(llvm.InitializeTargetMC(llvm.X86))
		if err != nil {
			return llvm.Target{}, err
		}
		err = stlerr.ErrorWrap(llvm.InitializeTargetMC(llvm.X86))
		if err != nil {
			return llvm.Target{}, err
		}
		target, err := stlerr.ErrorWith(llvm.NewTargetFromTriple("x86_64-pc-windows-msvc"))
		if err != nil {
			return llvm.Target{}, err
		}
		err = stlerr.ErrorWrap(llvm.InitializeAsmParser(llvm.X86))
		return target, err
	default:
		target, err := stlerr.ErrorWith(llvm.NativeTarget())
		if err != nil {
			return llvm.Target{}, err
		}
		err = stlerr.ErrorWrap(llvm.InitializeNativeAsmPrinter())
		return target, err
	}
}
