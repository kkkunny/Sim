package analyse

import (
	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseType(node ast.Type) Type {
	switch typeNode := node.(type) {
	case *ast.IdentType:
		return self.analyseIdentType(typeNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseOptionType(node util.Option[ast.Type]) Type {
	t, ok := node.Value()
	if !ok {
		return Empty
	}
	return self.analyseType(t)
}

func (self *Analyser) analyseIdentType(node *ast.IdentType) Type {
	switch node.Name.Source() {
	case "isize":
		return Isize
	default:
		// TODO: 编译时异常：未知的类型
		panic("unreachable")
	}
}
