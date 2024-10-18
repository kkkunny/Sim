package analyse

import (
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/oldhir"
	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/ast"

	errors "github.com/kkkunny/Sim/compiler/error"
)

func (self *Analyser) analyseStmt(node ast.Stmt) (oldhir.Stmt, oldhir.BlockEof) {
	switch stmtNode := node.(type) {
	case *ast.Return:
		ret := self.analyseReturn(stmtNode)
		return ret, oldhir.BlockEofReturn
	case *ast.SingleVariableDef:
		return self.analyseSingleLocalVariable(stmtNode), oldhir.BlockEofNone
	case *ast.MultipleVariableDef:
		return self.analyseLocalMultiVariable(stmtNode), oldhir.BlockEofNone
	case *ast.Block:
		return self.analyseBlock(stmtNode, nil)
	case *ast.IfElse:
		return self.analyseIfElse(stmtNode)
	case ast.Expr:
		return self.analyseExpr(nil, stmtNode).(oldhir.ExprStmt), oldhir.BlockEofNone
	case *ast.While:
		return self.analyseWhile(stmtNode)
	case *ast.Break:
		return self.analyseBreak(stmtNode), oldhir.BlockEofBreakLoop
	case *ast.Continue:
		return self.analyseContinue(stmtNode), oldhir.BlockEofNextLoop
	case *ast.For:
		return self.analyseFor(stmtNode)
	case *ast.Match:
		return self.analyseMatch(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseBlock(node *ast.Block, afterBlockCreate func(scope _LocalScope)) (*oldhir.Block, oldhir.BlockEof) {
	blockScope := _NewBlockScope(self.localScope)
	if afterBlockCreate != nil {
		afterBlockCreate(blockScope)
	}

	self.localScope = blockScope
	defer func() {
		self.localScope = self.localScope.GetParent().(_LocalScope)
	}()

	var jump oldhir.BlockEof
	stmts := linkedlist.NewLinkedList[oldhir.Stmt]()
	for iter := node.Stmts.Iterator(); iter.Next(); {
		stmt, stmtJump := self.analyseStmt(iter.Value())
		if b, ok := stmt.(*oldhir.Block); ok {
			for iter := b.Stmts.Iterator(); iter.Next(); {
				stmts.PushBack(iter.Value())
			}
		} else {
			stmts.PushBack(stmt)
		}
		jump = max(jump, stmtJump)
	}
	return &oldhir.Block{Stmts: stmts}, jump
}

func (self *Analyser) analyseReturn(node *ast.Return) *oldhir.Return {
	f := self.localScope.GetFunc()
	ft := f.GetFuncType()
	if v, ok := node.Value.Value(); ok {
		value := self.expectExpr(ft.Ret, v)
		return oldhir.NewReturn(f, value)
	} else {
		if !ft.Ret.EqualTo(oldhir.NoThing) {
			errors.ThrowTypeMismatchError(node.Position(), ft.Ret, oldhir.NoThing)
		}
		return oldhir.NewReturn(f)
	}
}

func (self *Analyser) analyseSingleLocalVariable(node *ast.SingleVariableDef) *oldhir.LocalVarDef {
	v := oldhir.NewLocalVarDef(node.Var.Mutable, node.Var.Name.Source(), nil)
	if !self.localScope.SetValue(v.Name, v) {
		errors.ThrowIdentifierDuplicationError(node.Var.Name.Position, node.Var.Name)
	}

	if typeNode, ok := node.Var.Type.Value(); ok {
		v.Type = self.analyseType(typeNode)
		if valueNode, ok := node.Value.Value(); ok {
			v.SetValue(self.expectExpr(v.Type, valueNode))
		} else {
			v.SetValue(self.getTypeDefaultValue(typeNode.Position(), v.Type))
		}
	} else {
		v.SetValue(self.analyseExpr(nil, node.Value.MustValue()))
		v.Type = v.Value.MustValue().GetType()
	}
	return v
}

func (self *Analyser) analyseLocalMultiVariable(node *ast.MultipleVariableDef) *oldhir.MultiLocalVarDef {
	allHaveType := true
	vars := stlslices.Map(node.Vars, func(_ int, e ast.VarDef) *oldhir.LocalVarDef {
		v := oldhir.NewLocalVarDef(e.Mutable, e.Name.Source(), nil)
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
	varTypes := stlslices.Map(vars, func(_ int, e *oldhir.LocalVarDef) oldhir.Type {
		return e.GetType()
	})

	var value oldhir.Expr
	if valueNode, ok := node.Value.Value(); ok && allHaveType {
		value = self.expectExpr(&oldhir.TupleType{Elems: varTypes}, valueNode)
	} else if ok {
		value = self.analyseExpr(&oldhir.TupleType{Elems: varTypes}, valueNode)
		vt, ok := value.GetType().(*oldhir.TupleType)
		if !ok {
			errors.ThrowExpectTupleError(valueNode.Position(), value.GetType())
		} else if len(vt.Elems) != len(vars) {
			errors.ThrowParameterNumberNotMatchError(valueNode.Position(), uint(len(vars)), uint(len(vt.Elems)))
		}
		for i, v := range vars {
			v.Type = vt.Elems[i]
		}
	} else {
		elems := stlslices.Map(vars, func(i int, e *oldhir.LocalVarDef) oldhir.Expr {
			return self.getTypeDefaultValue(node.Vars[i].Type.MustValue().Position(), e.GetType())
		})
		value = &oldhir.Tuple{Elems: elems}
	}

	return &oldhir.MultiLocalVarDef{
		Vars:  vars,
		Value: value,
	}
}

func (self *Analyser) analyseIfElse(node *ast.IfElse) (*oldhir.IfElse, oldhir.BlockEof) {
	if condNode, ok := node.Cond.Value(); ok {
		cond := self.expectExpr(self.pkgScope.Bool(), condNode)
		body, jump := self.analyseBlock(node.Body, nil)

		var next optional.Optional[*oldhir.IfElse]
		if nextNode, ok := node.Next.Value(); ok {
			nextIf, nextJump := self.analyseIfElse(nextNode)
			next = optional.Some(nextIf)
			jump = max(jump, nextJump)
		} else {
			jump = oldhir.BlockEofNone
		}

		return &oldhir.IfElse{
			Cond: optional.Some(cond),
			Body: body,
			Next: next,
		}, jump
	} else {
		body, jump := self.analyseBlock(node.Body, nil)
		return &oldhir.IfElse{Body: body}, jump
	}
}

func (self *Analyser) analyseWhile(node *ast.While) (*oldhir.While, oldhir.BlockEof) {
	cond := self.expectExpr(self.pkgScope.Bool(), node.Cond)
	loop := &oldhir.While{Cond: cond}
	body, eof := self.analyseBlock(node.Body, func(scope _LocalScope) {
		scope.SetLoop(loop)
	})
	loop.Body = body

	if eof == oldhir.BlockEofNextLoop || eof == oldhir.BlockEofBreakLoop {
		eof = oldhir.BlockEofNone
	}
	return loop, eof
}

func (self *Analyser) analyseBreak(node *ast.Break) *oldhir.Break {
	loop := self.localScope.GetLoop()
	if loop == nil {
		errors.ThrowLoopControlError(node.Position())
	}
	return &oldhir.Break{Loop: loop}
}

func (self *Analyser) analyseContinue(node *ast.Continue) *oldhir.Continue {
	loop := self.localScope.GetLoop()
	if loop == nil {
		errors.ThrowLoopControlError(node.Position())
	}
	return &oldhir.Continue{Loop: loop}
}

func (self *Analyser) analyseFor(node *ast.For) (*oldhir.For, oldhir.BlockEof) {
	iterator := self.analyseExpr(nil, node.Iterator)
	iterType := iterator.GetType()
	if !oldhir.IsType[*oldhir.ArrayType](iterType) {
		errors.ThrowExpectArrayError(node.Iterator.Position(), iterType)
	}

	et := oldhir.AsType[*oldhir.ArrayType](iterType).Elem
	loop := &oldhir.For{
		Iterator: iterator,
		Cursor:   oldhir.NewLocalVarDef(node.CursorMut, node.Cursor.Source(), et),
	}
	body, eof := self.analyseBlock(node.Body, func(scope _LocalScope) {
		if !scope.SetValue(loop.Cursor.Name, loop.Cursor) {
			errors.ThrowIdentifierDuplicationError(node.Cursor.Position, node.Cursor)
		}
		scope.SetLoop(loop)
	})
	loop.Body = body

	if eof == oldhir.BlockEofNextLoop || eof == oldhir.BlockEofBreakLoop {
		eof = oldhir.BlockEofNone
	}
	return loop, eof
}

func (self *Analyser) analyseMatch(node *ast.Match) (*oldhir.Match, oldhir.BlockEof) {
	value := self.analyseExpr(nil, node.Value)
	vtObj := value.GetType()
	vt, ok := oldhir.TryType[*oldhir.EnumType](vtObj)
	if !ok {
		errors.ThrowExpectEnumTypeError(node.Value.Position(), vtObj)
	}

	cases := linkedhashmap.StdWithCap[string, *oldhir.MatchCase](uint(len(node.Cases)))
	for _, caseNode := range node.Cases {
		caseName := caseNode.Name.Source()
		if !vt.Fields.Contain(caseName) {
			errors.ThrowUnknownIdentifierError(caseNode.Name.Position, caseNode.Name)
		} else if cases.Contain(caseName) {
			errors.ThrowIdentifierDuplicationError(caseNode.Name.Position, caseNode.Name)
		}
		caseDef := vt.Fields.Get(caseName)
		if caseDef.Elem.IsNone() && len(caseNode.Elems) != 0 {
			errors.ThrowParameterNumberNotMatchError(reader.MixPosition(caseNode.Name.Position, caseNode.ElemEnd), 0, uint(len(caseNode.Elems)))
		} else if caseDef.Elem.IsSome() && len(caseNode.Elems) != 1 {
			errors.ThrowParameterNumberNotMatchError(reader.MixPosition(caseNode.Name.Position, caseNode.ElemEnd), 1, uint(len(caseNode.Elems)))
		}
		elems := stlslices.Map(caseNode.Elems, func(i int, e ast.MatchCaseElem) *oldhir.Param {
			return &oldhir.Param{
				Mut:  e.Mutable,
				Type: caseDef.Elem.MustValue(),
				Name: optional.Some(e.Name.Source()),
			}
		})
		fn := func(scope _LocalScope) {
			for i, elemNode := range caseNode.Elems {
				elemName := elemNode.Name.Source()
				if !scope.SetValue(elemName, elems[i]) {
					errors.ThrowIdentifierDuplicationError(elemNode.Name.Position, elemNode.Name)
				}
			}
		}
		cases.Set(caseName, &oldhir.MatchCase{
			Name:  caseName,
			Elems: elems,
			Body:  stlval.IgnoreWith(self.analyseBlock(caseNode.Body, fn)),
		})
	}

	var other optional.Optional[*oldhir.Block]
	if otherNode, ok := node.Other.Value(); ok {
		other = optional.Some(stlval.IgnoreWith(self.analyseBlock(otherNode, nil)))
	}

	if other.IsNone() && cases.Length() != vt.Fields.Length() {
		errors.ThrowExpectMoreCase(node.Value.Position(), vtObj, cases.Length(), vt.Fields.Length())
	}

	return &oldhir.Match{
		Value: value,
		Cases: cases,
		Other: other,
	}, oldhir.BlockEofNone
}
