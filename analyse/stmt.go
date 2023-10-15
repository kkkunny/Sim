package analyse

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseStmt(node ast.Stmt) Stmt {
	switch stmtNode := node.(type) {
	case *ast.Return:
		return self.analyseReturn(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseBlock(node *ast.Block) *Block {
	stmts := linkedlist.NewLinkedList[Stmt]()
	for iter := node.Stmts.Iterator(); iter.Next(); {
		stmts.PushBack(self.analyseStmt(iter.Value()))
	}
	return &Block{Stmts: stmts}
}

func (self *Analyser) analyseReturn(node *ast.Return) *Return {
	value := util.None[Expr]()
	if v, ok := node.Value.Value(); ok {
		value = util.Some[Expr](self.analyseExpr(v))
	}
	return &Return{Value: value}
}
