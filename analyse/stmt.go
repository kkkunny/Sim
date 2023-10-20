package analyse

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseStmt(node ast.Stmt) (Stmt, bool) {
	switch stmtNode := node.(type) {
	case *ast.Return:
		return self.analyseReturn(stmtNode), true
	case ast.Expr:
		return self.analyseExpr(nil, stmtNode), false
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseBlock(node *ast.Block) (*Block, bool) {
	self.localScope = _NewBlockScope(self.localScope)
	defer func() {
		self.localScope = self.localScope.GetParent().(_LocalScope)
	}()

	var end bool
	stmts := linkedlist.NewLinkedList[Stmt]()
	for iter := node.Stmts.Iterator(); iter.Next(); {
		stmt, stmtEnd := self.analyseStmt(iter.Value())
		end = end || stmtEnd
		stmts.PushBack(stmt)
	}
	return &Block{Stmts: stmts}, end
}

func (self *Analyser) analyseReturn(node *ast.Return) *Return {
	expectRetType := self.localScope.GetRetType()
	if v, ok := node.Value.Value(); ok {
		value := self.analyseExpr(expectRetType, v)
		if !value.GetType().Equal(expectRetType) {
			// TODO: 编译时异常：类型不相等
			panic("unreachable")
		}
		return &Return{Value: util.Some[Expr](value)}
	} else {
		if !expectRetType.Equal(Empty) {
			// TODO: 编译时异常：类型不相等
			panic("unreachable")
		}
		return &Return{Value: util.None[Expr]()}
	}
}
