package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/linkedlist"

	"github.com/kkkunny/Sim/ast"
	errors "github.com/kkkunny/Sim/error"
	. "github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
)

func (self *Analyser) analyseStmt(node ast.Stmt) (Stmt, JumpOut) {
	switch stmtNode := node.(type) {
	case *ast.Return:
		ret := self.analyseReturn(stmtNode)
		return ret, JumpOutReturn
	case *ast.Variable:
		return self.analyseLocalVariable(stmtNode), JumpOutNone
	case *ast.Block:
		return self.analyseBlock(stmtNode, nil)
	case *ast.IfElse:
		return self.analyseIfElse(stmtNode)
	case ast.Expr:
		return self.analyseExpr(nil, stmtNode), JumpOutNone
	case *ast.Loop:
		return self.analyseLoop(stmtNode)
	case *ast.Break:
		return self.analyseBreak(stmtNode), JumpOutLoop
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseBlock(node *ast.Block, loop *Loop) (*Block, JumpOut) {
	self.localScope = _NewBlockScope(self.localScope)
	self.localScope.SetLoop(loop)
	defer func() {
		self.localScope = self.localScope.GetParent().(_LocalScope)
	}()

	var jump JumpOut
	stmts := linkedlist.NewLinkedList[Stmt]()
	for iter := node.Stmts.Iterator(); iter.Next(); {
		stmt, stmtJump := self.analyseStmt(iter.Value())
		if b, ok := stmt.(*Block); ok {
			for iter := b.Stmts.Iterator(); iter.Next(); {
				stmts.PushBack(iter.Value())
			}
		} else {
			stmts.PushBack(stmt)
		}
		jump = max(jump, stmtJump)
	}
	return &Block{Stmts: stmts}, jump
}

func (self *Analyser) analyseReturn(node *ast.Return) *Return {
	expectRetType := self.localScope.GetRetType()
	if v, ok := node.Value.Value(); ok {
		value := self.expectExpr(expectRetType, v)
		return &Return{Value: util.Some[Expr](value)}
	} else {
		if !expectRetType.Equal(Empty) {
			errors.ThrowTypeMismatchError(node.Position(), expectRetType, Empty)
		}
		return &Return{Value: util.None[Expr]()}
	}
}

func (self *Analyser) analyseLocalVariable(node *ast.Variable) *Variable {
	v := &Variable{
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

func (self *Analyser) analyseIfElse(node *ast.IfElse) (*IfElse, JumpOut) {
	if condNode, ok := node.Cond.Value(); ok {
		cond := self.expectExpr(Bool, condNode)
		body, jump := self.analyseBlock(node.Body, nil)

		var next util.Option[*IfElse]
		if nextNode, ok := node.Next.Value(); ok {
			nextIf, nextJump := self.analyseIfElse(nextNode)
			next = util.Some(nextIf)
			jump = max(jump, nextJump)
		} else {
			jump = JumpOutNone
		}

		return &IfElse{
			Cond: util.Some(cond),
			Body: body,
			Next: next,
		}, jump
	} else {
		body, jump := self.analyseBlock(node.Body, nil)
		return &IfElse{Body: body}, jump
	}
}

func (self *Analyser) analyseLoop(node *ast.Loop) (*Loop, JumpOut) {
	loop := &Loop{}
	body, jump := self.analyseBlock(node.Body, loop)
	loop.Body = body
	return loop, stlbasic.Ternary(jump == JumpOutLoop, JumpOutNone, jump)
}

func (self *Analyser) analyseBreak(node *ast.Break) *Break {
	loop := self.localScope.GetLoop()
	if loop == nil {
		errors.ThrowLoopControlError(node.Position())
	}
	return &Break{Loop: loop}
}
