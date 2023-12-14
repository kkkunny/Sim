package function

import (
	"github.com/kkkunny/Sim/mir"
)

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

var (
	// DeadCodeElimination 死码消除
	DeadCodeElimination = newUnionPass(UnreachableCodeElimination, DeadVariablesElimination, CodeBlockMerge)
)

func newUnionPass(pass ...Pass)Pass{
	return &_UnionPass{passes: pass}
}

type _UnionPass struct {
	passes []Pass
}

func (self *_UnionPass) init(_ mir.Function){}

func (self *_UnionPass) Run(ir *mir.Function){
	for _, pass := range self.passes{
		pass.init(*ir)
		pass.Run(ir)
	}
}
