package analyse

import (
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

func New(parser *parse.Parser) *Analyser {
	return &Analyser{
		parser:   parser,
		pkgScope: _NewPkgScope(),
	}
}

// Analyse 分析语义
func (self *Analyser) Analyse() linkedlist.LinkedList[Global] {
	meanNodes := linkedlist.NewLinkedList[Global]()
	iter := self.parser.Parse().Iterator()
	iter.Foreach(func(v ast.Global) bool {
		if st, ok := v.(*ast.StructDef); ok {
			self.declTypeDef(st)
		}
		return true
	})
	iter.Reset()
	iter.Foreach(func(v ast.Global) bool {
		self.analyseGlobalDecl(v)
		return true
	})
	iter.Reset()
	iter.Foreach(func(v ast.Global) bool {
		meanNodes.PushBack(self.analyseGlobalDef(v))
		return true
	})
	return meanNodes
}
