package module

import "github.com/kkkunny/Sim/mir"

// Pass 模块pass
type Pass interface {
	init(module mir.Module)
	Run(module *mir.Module)
}

// Run 运行pass
func Run(module *mir.Module, pass ...Pass){
	for _, p := range pass{
		p.init(*module)
		p.Run(module)
	}
}
