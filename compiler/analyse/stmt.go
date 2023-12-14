package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/hir"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseStmt(node ast.Stmt) (hir.Stmt, hir.BlockEof) {
	switch stmtNode := node.(type) {
	case *ast.Return:
		ret := self.analyseReturn(stmtNode)
		return ret, hir.BlockEofReturn
	case *ast.Variable:
		return self.analyseLocalVariable(stmtNode), hir.BlockEofNone
	case *ast.Block:
		return self.analyseBlock(stmtNode, nil)
	case *ast.IfElse:
		return self.analyseIfElse(stmtNode)
	case ast.Expr:
		return self.analyseExpr(nil, stmtNode), hir.BlockEofNone
	case *ast.Loop:
		return self.analyseEndlessLoop(stmtNode)
	case *ast.Break:
		return self.analyseBreak(stmtNode), hir.BlockEofBreakLoop
	case *ast.Continue:
		return self.analyseContinue(stmtNode), hir.BlockEofNextLoop
	case *ast.For:
		return self.analyseFor(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseBlock(node *ast.Block, afterBlockCreate func(scope _LocalScope)) (*hir.Block, hir.BlockEof) {
	blockScope := _NewBlockScope(self.localScope)
	if afterBlockCreate != nil {
		afterBlockCreate(blockScope)
	}

	self.localScope = blockScope
	defer func() {
		self.localScope = self.localScope.GetParent().(_LocalScope)
	}()

	var jump hir.BlockEof
	stmts := linkedlist.NewLinkedList[hir.Stmt]()
	for iter := node.Stmts.Iterator(); iter.Next(); {
		stmt, stmtJump := self.analyseStmt(iter.Value())
		if b, ok := stmt.(*hir.Block); ok {
			for iter := b.Stmts.Iterator(); iter.Next(); {
				stmts.PushBack(iter.Value())
			}
		} else {
			stmts.PushBack(stmt)
		}
		jump = max(jump, stmtJump)
	}
	return &hir.Block{Stmts: stmts}, jump
}

func (self *Analyser) analyseReturn(node *ast.Return) *hir.Return {
	f := self.localScope.GetFunc()
	ft := f.GetFuncType()
	if v, ok := node.Value.Value(); ok {
		value := self.expectExpr(ft.Ret, v)
		return &hir.Return{
			Func:  f,
			Value: util.Some[hir.Expr](value),
		}
	} else {
		if !ft.Ret.Equal(hir.Empty) {
			errors.ThrowTypeMismatchError(node.Position(), ft.Ret, hir.Empty)
		}
		return &hir.Return{
			Func:  f,
			Value: util.None[hir.Expr](),
		}
	}
}

func (self *Analyser) analyseLocalVariable(node *ast.Variable) *hir.VarDef {
	v := &hir.VarDef{
		Mut:  node.Mutable,
		Name: node.Name.Source(),
	}
	if !self.localScope.SetValue(v.Name, v) {
		errors.ThrowIdentifierDuplicationError(node.Position(), node.Name)
	}

	if typeNode, ok := node.Type.Value(); ok{
		v.Type = self.analyseType(typeNode)
		if valueNode, ok := node.Value.Value(); ok{
			v.Value = self.expectExpr(v.Type, valueNode)
		}else{
			v.Value = self.getTypeDefaultValue(typeNode.Position(), v.Type)
		}
	}else{
		v.Value = self.analyseExpr(nil, node.Value.MustValue())
		v.Type = v.Value.GetType()
	}
	return v
}

func (self *Analyser) analyseIfElse(node *ast.IfElse) (*hir.IfElse, hir.BlockEof) {
	if condNode, ok := node.Cond.Value(); ok {
		cond := self.expectExpr(hir.Bool, condNode)
		body, jump := self.analyseBlock(node.Body, nil)

		var next util.Option[*hir.IfElse]
		if nextNode, ok := node.Next.Value(); ok {
			nextIf, nextJump := self.analyseIfElse(nextNode)
			next = util.Some(nextIf)
			jump = max(jump, nextJump)
		} else {
			jump = hir.BlockEofNone
		}

		return &hir.IfElse{
			Cond: util.Some(cond),
			Body: body,
			Next: next,
		}, jump
	} else {
		body, jump := self.analyseBlock(node.Body, nil)
		return &hir.IfElse{Body: body}, jump
	}
}

func (self *Analyser) analyseEndlessLoop(node *ast.Loop) (*hir.EndlessLoop, hir.BlockEof) {
	loop := &hir.EndlessLoop{}
	body, eof := self.analyseBlock(node.Body, func(scope _LocalScope) {
		scope.SetLoop(loop)
	})
	loop.Body = body

	if eof == hir.BlockEofNextLoop || eof == hir.BlockEofBreakLoop {
		eof = hir.BlockEofNone
	}
	return loop, eof
}

func (self *Analyser) analyseBreak(node *ast.Break) *hir.Break {
	loop := self.localScope.GetLoop()
	if loop == nil {
		errors.ThrowLoopControlError(node.Position())
	}
	return &hir.Break{Loop: loop}
}

func (self *Analyser) analyseContinue(node *ast.Continue) *hir.Continue {
	loop := self.localScope.GetLoop()
	if loop == nil {
		errors.ThrowLoopControlError(node.Position())
	}
	return &hir.Continue{Loop: loop}
}

func (self *Analyser) analyseFor(node *ast.For) (*hir.For, hir.BlockEof) {
	iterator := self.analyseExpr(nil, node.Iterator)
	iterType := iterator.GetType()
	if !stlbasic.Is[*hir.ArrayType](iterType) {
		errors.ThrowNotArrayError(node.Iterator.Position(), iterType)
	}

	et := iterType.(*hir.ArrayType).Elem
	loop := &hir.For{
		Iterator: iterator,
		Cursor: &hir.VarDef{
			Mut:   node.CursorMut,
			Type:  et,
			Name:  node.Cursor.Source(),
			Value: &hir.Default{Type: et},
		},
	}
	body, eof := self.analyseBlock(node.Body, func(scope _LocalScope) {
		if !scope.SetValue(loop.Cursor.Name, loop.Cursor) {
			errors.ThrowIdentifierDuplicationError(node.Cursor.Position, node.Cursor)
		}
		scope.SetLoop(loop)
	})
	loop.Body = body

	if eof == hir.BlockEofNextLoop || eof == hir.BlockEofBreakLoop {
		eof = hir.BlockEofNone
	}
	return loop, eof
}
