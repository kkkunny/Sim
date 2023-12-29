package c

import (
	"github.com/kkkunny/Sim/mir"
)

func getTarget(target mir.Target)(string, string){
	switch target.Name() {
	case mir.PackTargetName(mir.ArchX8664, mir.OSWindows):
		return "windows", "amd64"
	case mir.PackTargetName(mir.ArchX8664, mir.OSLinux):
		return "linux", "amd64"
	default:
		panic("unreachable")
	}
}
