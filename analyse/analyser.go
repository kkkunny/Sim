package analyse

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/iterator"
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/parse"
)

// Analyser 语义分析器
type Analyser struct {
	parser *parse.Parser

	pkgScope   *_PkgScope
	localScope _LocalScope
}

func New(parser *parse.Parser, target *llvm.Target) *Analyser {
	Isize.Bits = target.PointerSize()
	Usize.Bits = target.PointerSize()
	return &Analyser{
		parser:   parser,
		pkgScope: _NewPkgScope(),
	}
}

// Analyse 分析语义
func (self *Analyser) Analyse() linkedlist.LinkedList[Global] {
	meanNodes := linkedlist.NewLinkedList[Global]()
	astNodes := self.parser.Parse()
	iterator.Foreach(astNodes, func(v ast.Global) bool {
		if st, ok := v.(*ast.StructDef); ok {
			self.declTypeDef(st)
		}
		return true
	})
	iterator.Foreach(astNodes, func(v ast.Global) bool {
		self.analyseGlobalDecl(v)
		return true
	})
	iterator.Foreach(astNodes, func(v ast.Global) bool {
		meanNodes.PushBack(self.analyseGlobalDef(v))
		return true
	})
	return meanNodes
}
