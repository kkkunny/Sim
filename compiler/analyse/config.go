package analyse

import "github.com/kkkunny/Sim/ast"

type analyseConfig struct{
	PtrBits uint  // 指针大小

	// traits
	DefaultTrait *ast.Trait
}

func (self *Analyser) getConfig()*analyseConfig{
	if self.parent == nil{
		return self.config
	}
	return self.parent.getConfig()
}
