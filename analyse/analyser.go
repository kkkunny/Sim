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
}

func New(parser *parse.Parser) *Analyser {
	return &Analyser{parser: parser}
}

// Analyse 分析语义
func (self *Analyser) Analyse() linkedlist.LinkedList[Global] {
	meanNodes := linkedlist.NewLinkedList[Global]()
	self.parser.Parse().Iterator().Foreach(func(v ast.Global) bool {
		meanNodes.PushBack(self.analyseGlobal(v))
		return true
	})
	return meanNodes
}
