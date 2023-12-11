package pass

import "github.com/kkkunny/Sim/mir"

type Pass interface {
	Run(module *mir.Module)
}

// Run 运行pass
func Run(module *mir.Module, pass ...Pass){
	for _, p := range pass{
		p.Run(module)
	}
}
