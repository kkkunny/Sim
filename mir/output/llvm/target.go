package llvm

import (
	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/mir"
)

func getTarget(target mir.Target)*llvm.Target{
	llvm.InitializeAllTargetInfos()
	llvm.InitializeAllTargets()
	llvm.InitializeAllTargetMCs()
	switch target.Name() {
	case mir.PackTargetName(mir.ArchX8664, mir.OSWindows):
		return stlerror.MustWith(llvm.NewTargetFromTriple("x86_64-w64-windows-gnu", "generic", ""))
	default:
		panic("unreachable")
	}
}
