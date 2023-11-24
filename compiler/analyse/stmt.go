package analyse

import (
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/mean"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseStmt(node ast.Stmt) (mean.Stmt, mean.BlockEof) {
	switch stmtNode := node.(type) {
	case *ast.Return:
		ret := self.analyseReturn(stmtNode)
		return ret, mean.BlockEofReturn
	case *ast.Variable:
		return self.analyseLocalVariable(stmtNode), mean.BlockEofNone
	case *ast.Block:
		return self.analyseBlock(stmtNode, nil)
	case *ast.IfElse:
		return self.analyseIfElse(stmtNode)
	case ast.Expr:
		return self.analyseExpr(nil, stmtNode), mean.BlockEofNone
	case *ast.Loop:
		return self.analyseEndlessLoop(stmtNode)
	case *ast.Break:
		return self.analyseBreak(stmtNode), mean.BlockEofBreakLoop
	case *ast.Continue:
		return self.analyseContinue(stmtNode), mean.BlockEofNextLoop
	case *ast.For:
		return self.analyseFor(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseBlock(node *ast.Block, afterBlockCreate func(scope _LocalScope)) (*mean.Block, mean.BlockEof) {
	blockScope := _NewBlockScope(self.localScope)
	if afterBlockCreate != nil {
		afterBlockCreate(blockScope)
	}

	self.localScope = blockScope
	defer func() {
		self.localScope = self.localScope.GetParent().(_LocalScope)
	}()

	var jump mean.BlockEof
	stmts := linkedlist.NewLinkedList[mean.Stmt]()
	for iter := node.Stmts.Iterator(); iter.Next(); {
		stmt, stmtJump := self.analyseStmt(iter.Value())
		if b, ok := stmt.(*mean.Block); ok {
			for iter := b.Stmts.Iterator(); iter.Next(); {
				stmts.PushBack(iter.Value())
			}
		} else {
			stmts.PushBack(stmt)
		}
		jump = max(jump, stmtJump)
	}
	return &mean.Block{Stmts: stmts}, jump
}

func (self *Analyser) analyseReturn(node *ast.Return) *mean.Return {
	f := self.localScope.GetFunc()
	if v, ok := node.Value.Value(); ok {
		value := self.expectExpr(f.Ret, v)
		return &mean.Return{
			Func:  f,
			Value: util.Some[mean.Expr](value),
		}
	} else {
		if !f.Ret.Equal(mean.Empty) {
			errors.ThrowTypeMismatchError(node.Position(), f.Ret, mean.Empty)
		}
		return &mean.Return{
			Func:  f,
			Value: util.None[mean.Expr](),
		}
	}
}

func (self *Analyser) analyseLocalVariable(node *ast.Variable) *mean.Variable {
	v := &mean.Variable{
		Mut:  node.Mutable,
		Name: node.Name.Source(),
	}
	if !self.localScope.SetValue(v.Name, v) {
		errors.ThrowIdentifierDuplicationError(node.Position(), node.Name)
	}

	v.Type = self.analyseType(node.Type)
	v.Value = self.expectExpr(v.Type, node.Value)
	return v
}

func (self *Analyser) analyseIfElse(node *ast.IfElse) (*mean.IfElse, mean.BlockEof) {
	if condNode, ok := node.Cond.Value(); ok {
		cond := self.expectExpr(mean.Bool, condNode)
		body, jump := self.analyseBlock(node.Body, nil)

		var next util.Option[*mean.IfElse]
		if nextNode, ok := node.Next.Value(); ok {
			nextIf, nextJump := self.analyseIfElse(nextNode)
			next = util.Some(nextIf)
			jump = max(jump, nextJump)
		} else {
			jump = mean.BlockEofNone
		}

		return &mean.IfElse{
			Cond: util.Some(cond),
			Body: body,
			Next: next,
		}, jump
	} else {
		body, jump := self.analyseBlock(node.Body, nil)
		return &mean.IfElse{Body: body}, jump
	}
}

func (self *Analyser) analyseEndlessLoop(node *ast.Loop) (*mean.EndlessLoop, mean.BlockEof) {
	loop := &mean.EndlessLoop{}
	body, eof := self.analyseBlock(node.Body, func(scope _LocalScope) {
		scope.SetLoop(loop)
	})
	loop.Body = body

	if eof == mean.BlockEofNextLoop || eof == mean.BlockEofBreakLoop {
		eof = mean.BlockEofNone
	}
	return loop, eof
}

func (self *Analyser) analyseBreak(node *ast.Break) *mean.Break {
	loop := self.localScope.GetLoop()
	if loop == nil {
		errors.ThrowLoopControlError(node.Position())
	}
	return &mean.Break{Loop: loop}
}

func (self *Analyser) analyseContinue(node *ast.Continue) *mean.Continue {
	loop := self.localScope.GetLoop()
	if loop == nil {
		errors.ThrowLoopControlError(node.Position())
	}
	return &mean.Continue{Loop: loop}
}

func (self *Analyser) analyseFor(node *ast.For) (*mean.For, mean.BlockEof) {
	iterator := self.analyseExpr(nil, node.Iterator)
	iterType := iterator.GetType()
	if !mean.TypeIs[*mean.ArrayType](iterType) {
		errors.ThrowNotArrayError(node.Iterator.Position(), iterType)
	}

	et := iterType.(*mean.ArrayType).Elem
	loop := &mean.For{
		Iterator: iterator,
		Cursor: &mean.Variable{
			Mut:   node.CursorMut,
			Type:  et,
			Name:  node.Cursor.Source(),
			Value: &mean.Zero{Type: et},
		},
	}
	body, eof := self.analyseBlock(node.Body, func(scope _LocalScope) {
		if !scope.SetValue(loop.Cursor.Name, loop.Cursor) {
			errors.ThrowIdentifierDuplicationError(node.Cursor.Position, node.Cursor)
		}
		scope.SetLoop(loop)
	})
	loop.Body = body

	if eof == mean.BlockEofNextLoop || eof == mean.BlockEofBreakLoop {
		eof = mean.BlockEofNone
	}
	return loop, eof
}
