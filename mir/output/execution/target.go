package execution

import (
	"sync"

	"github.com/kkkunny/go-llvm"
	stlerror "github.com/kkkunny/stl/error"

	"github.com/kkkunny/Sim/mir"
)

var targetInitFunc = map[mir.Arch]func(){
	mir.ArchX8664: sync.OnceFunc(func() {
		stlerror.Must(llvm.InitializeAsmParser(llvm.X86))
		stlerror.Must(llvm.InitializeAsmPrinter(llvm.X86))
	}),
}

func initTarget(target mir.Target){
	if initFn, ok := targetInitFunc[target.Arch()]; ok{
		initFn()
	}
}
