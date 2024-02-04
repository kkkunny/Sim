package analyse

import (
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/slices"

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
	case *ast.SingleVariableDef:
		return self.analyseSingleLocalVariable(stmtNode), hir.BlockEofNone
	case *ast.MultipleVariableDef:
		return self.analyseLocalMultiVariable(stmtNode), hir.BlockEofNone
	case *ast.Block:
		return self.analyseBlock(stmtNode, nil)
	case *ast.IfElse:
		return self.analyseIfElse(stmtNode)
	case ast.Expr:
		return self.analyseExpr(nil, stmtNode), hir.BlockEofNone
	case *ast.While:
		return self.analyseWhile(stmtNode)
	case *ast.Break:
		return self.analyseBreak(stmtNode), hir.BlockEofBreakLoop
	case *ast.Continue:
		return self.analyseContinue(stmtNode), hir.BlockEofNextLoop
	case *ast.For:
		return self.analyseFor(stmtNode)
	case *ast.Match:
		return self.analyseMatch(stmtNode)
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

func (self *Analyser) analyseSingleLocalVariable(node *ast.SingleVariableDef) *hir.LocalVarDef {
	v := &hir.LocalVarDef{
		VarDecl: hir.VarDecl{
			Mut:  node.Var.Mutable,
			Name: node.Var.Name.Source(),
		},
	}
	if !self.localScope.SetValue(v.Name, v) {
		errors.ThrowIdentifierDuplicationError(node.Var.Name.Position, node.Var.Name)
	}

	if typeNode, ok := node.Var.Type.Value(); ok {
		v.Type = self.analyseType(typeNode)
		if valueNode, ok := node.Value.Value(); ok {
			v.Value = self.expectExpr(v.Type, valueNode)
		} else {
			v.Value = self.getTypeDefaultValue(typeNode.Position(), v.Type)
		}
	} else {
		v.Value = self.analyseExpr(nil, node.Value.MustValue())
		v.Type = v.Value.GetType()
	}
	return v
}

func (self *Analyser) analyseLocalMultiVariable(node *ast.MultipleVariableDef) *hir.MultiLocalVarDef {
	allHaveType := true
	vars := stlslices.Map(node.Vars, func(_ int, e ast.VarDef) *hir.LocalVarDef {
		v := &hir.LocalVarDef{
			VarDecl: hir.VarDecl{
				Mut:  e.Mutable,
				Name: e.Name.Source(),
			},
		}
		if typeNode, ok := e.Type.Value(); ok {
			v.Type = self.analyseType(typeNode)
		} else {
			allHaveType = false
		}
		if !self.localScope.SetValue(v.Name, v) {
			errors.ThrowIdentifierDuplicationError(e.Name.Position, e.Name)
		}
		return v
	})
	varTypes := stlslices.Map(vars, func(_ int, e *hir.LocalVarDef) hir.Type {
		return e.GetType()
	})

	var value hir.Expr
	if valueNode, ok := node.Value.Value(); ok && allHaveType {
		value = self.expectExpr(&hir.TupleType{Elems: varTypes}, valueNode)
	} else if ok {
		value = self.analyseExpr(&hir.TupleType{Elems: varTypes}, valueNode)
		vt, ok := value.GetType().(*hir.TupleType)
		if !ok {
			errors.ThrowExpectTupleError(valueNode.Position(), value.GetType())
		} else if len(vt.Elems) != len(vars) {
			errors.ThrowParameterNumberNotMatchError(valueNode.Position(), uint(len(vars)), uint(len(vt.Elems)))
		}
		for i, v := range vars {
			v.Type = vt.Elems[i]
		}
	} else {
		elems := stlslices.Map(vars, func(i int, e *hir.LocalVarDef) hir.Expr {
			return self.getTypeDefaultValue(node.Vars[i].Type.MustValue().Position(), e.GetType())
		})
		value = &hir.Tuple{Elems: elems}
	}

	for _, v := range vars {
		v.Value = &hir.Default{Type: v.Type}
	}
	return &hir.MultiLocalVarDef{
		Vars:  vars,
		Value: value,
	}
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

func (self *Analyser) analyseWhile(node *ast.While) (*hir.While, hir.BlockEof) {
	cond := self.expectExpr(hir.Bool, node.Cond)
	loop := &hir.While{Cond: cond}
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
	if !hir.IsArrayType(iterType) {
		errors.ThrowExpectArrayError(node.Iterator.Position(), iterType)
	}

	et := hir.AsArrayType(iterType).Elem
	loop := &hir.For{
		Iterator: iterator,
		Cursor: &hir.LocalVarDef{
			VarDecl: hir.VarDecl{
				Mut:  node.CursorMut,
				Type: et,
				Name: node.Cursor.Source(),
			},
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

func (self *Analyser) analyseMatch(node *ast.Match) (*hir.Match, hir.BlockEof) {
	value := self.analyseExpr(nil, node.Value)
	vtObj := value.GetType()
	if !hir.IsUnionType(vtObj) {
		errors.ThrowExpectUnionTypeError(node.Value.Position(), vtObj)
	}
	vt := hir.AsUnionType(vtObj)

	cases := make([]pair.Pair[hir.Type, *hir.Block], len(node.Cases))
	for i, caseNode := range node.Cases {
		caseCond := self.analyseType(caseNode.First)
		if !vt.Contain(caseCond) {
			errors.ThrowTypeMismatchError(caseNode.First.Position(), caseCond, vtObj)
		} else if hir.IsUnionType(caseCond) {
			errors.ThrowNotExpectUnionTypeError(caseNode.First.Position(), caseCond)
		}
		var newValue *hir.LocalVarDef
		var fn func(_LocalScope)
		if v, ok := value.(hir.Variable); ok {
			newValue = &hir.LocalVarDef{
				VarDecl: hir.VarDecl{
					Mut:  v.Mutable(),
					Type: caseCond,
					Name: v.GetName(),
				},
				Value: &hir.UnUnion{
					Type:  caseCond,
					Value: v,
				},
			}
			fn = func(scope _LocalScope) {
				scope.SetValue(v.GetName(), newValue)
			}
		}
		caseBody, _ := self.analyseBlock(caseNode.Second, fn)
		if newValue != nil {
			caseBody.Stmts.PushFront(newValue)
		}
		cases[i] = pair.NewPair(caseCond, caseBody)
	}

	var other util.Option[*hir.Block]
	if otherNode, ok := node.Other.Value(); ok {
		caseBody, _ := self.analyseBlock(otherNode, nil)
		other = util.Some(caseBody)
	}

	return &hir.Match{
		Value: value,
		Cases: cases,
		Other: other,
	}, hir.BlockEofNone
}
