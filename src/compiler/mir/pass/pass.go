package pass

import "github.com/kkkunny/Sim/src/compiler/mir"

type Pass interface {
	walk(pkg *mir.Package)
}

type multiPasses struct {
	passes []Pass
}

func multi(pass ...Pass) *multiPasses {
	return &multiPasses{passes: pass}
}
func (self *multiPasses) walk(pkg *mir.Package) {
	for _, p := range self.passes {
		p.walk(pkg)
	}
}

var (
	UCE = newUnreachableCodeElimination() // 不可达代码消除
	DVE = newDeadVariablesElimination()   // 死变量消除
	DCE = multi(UCE, DVE)                 // 死代码消除（UCE+DVE）
	IE  = newInlineExpansion()            // 内联展开

	MUST = multi(IE, DCE) // 必须的pass
)

// WalkPass 执行pass
func WalkPass(pkg *mir.Package, pass ...Pass) *mir.Package {
	for _, p := range pass {
		p.walk(pkg)
	}
	return pkg
}
