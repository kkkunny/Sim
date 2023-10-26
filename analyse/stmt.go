package analyse

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/ast"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseStmt(node ast.Stmt) (Stmt, SkipOut) {
	switch stmtNode := node.(type) {
	case *ast.Return:
		ret := self.analyseReturn(stmtNode)
		return ret, ret
	case *ast.Variable:
		return self.analyseVariable(stmtNode), nil
	case *ast.Block:
		return self.analyseBlock(stmtNode)
	case *ast.If:
		return self.analyseIf(stmtNode), nil
	case ast.Expr:
		return self.analyseExpr(nil, stmtNode), nil
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseBlock(node *ast.Block) (*Block, SkipOut) {
	self.localScope = _NewBlockScope(self.localScope)
	defer func() {
		self.localScope = self.localScope.GetParent().(_LocalScope)
	}()

	var end SkipOut
	stmts := linkedlist.NewLinkedList[Stmt]()
	for iter := node.Stmts.Iterator(); iter.Next(); {
		stmt, out := self.analyseStmt(iter.Value())
		if b, ok := stmt.(*Block); ok {
			for iter := b.Stmts.Iterator(); iter.Next(); {
				stmts.PushBack(iter.Value())
			}
		} else {
			stmts.PushBack(stmt)
		}
		if out != nil {
			switch out.(type) {
			case *Return:
				end = out
			default:
				panic("unreachable")
			}
			break
		}
	}
	return &Block{Stmts: stmts}, end
}

func (self *Analyser) analyseReturn(node *ast.Return) *Return {
	expectRetType := self.localScope.GetRetType()
	if v, ok := node.Value.Value(); ok {
		value := self.expectExpr(expectRetType, v)
		return &Return{Value: util.Some[Expr](value)}
	} else {
		if !expectRetType.Equal(Empty) {
			// TODO: 编译时异常：类型不相等
			panic("unreachable")
		}
		return &Return{Value: util.None[Expr]()}
	}
}

func (self *Analyser) analyseVariable(node *ast.Variable) *Variable {
	v := &Variable{Name: node.Name.Source()}
	if !self.localScope.SetValue(v.Name, v) {
		// TODO: 编译时异常：变量名冲突
		panic("unreachable")
	}

	v.Type = self.analyseType(node.Type)
	v.Value = self.expectExpr(v.Type, node.Value)
	return v
}

func (self *Analyser) analyseIf(node *ast.If) *If {
	cond := self.expectExpr(Bool, node.Cond)
	body, _ := self.analyseBlock(node.Body)
	return &If{
		Cond: cond,
		Body: body,
	}
}
