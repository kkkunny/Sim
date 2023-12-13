package function

import "github.com/kkkunny/Sim/mir"

// Pass 函数pass
type Pass interface {
	init(ir mir.Function)
	Run(ir *mir.Function)
}

// Run 运行pass
func Run(function *mir.Function, pass ...Pass){
	for _, p := range pass{
		p.init(*function)
		p.Run(function)
	}
}
