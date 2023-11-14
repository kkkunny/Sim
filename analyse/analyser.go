package analyse

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/iterator"
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
)

// Analyser 语义分析器
type Analyser struct {
	asts linkedlist.LinkedList[ast.Global]

	pkgScope   *_PkgScope
	localScope _LocalScope
}

func New(asts linkedlist.LinkedList[ast.Global], target *llvm.Target) *Analyser {
	Isize.Bits = target.PointerSize()
	Usize.Bits = target.PointerSize()
	return &Analyser{
		asts:     asts,
		pkgScope: _NewPkgScope(),
	}
}

// Analyse 分析语义
func (self *Analyser) Analyse() linkedlist.LinkedList[Global] {
	iterator.Foreach(self.asts, func(v ast.Global) bool {
		if st, ok := v.(*ast.StructDef); ok {
			self.declTypeDef(st)
		}
		return true
	})
	iterator.Foreach(self.asts, func(v ast.Global) bool {
		self.analyseGlobalDecl(v)
		return true
	})
	meanNodes := linkedlist.NewLinkedList[Global]()
	iterator.Foreach(self.asts, func(v ast.Global) bool {
		meanNodes.PushBack(self.analyseGlobalDef(v))
		return true
	})
	return meanNodes
}
