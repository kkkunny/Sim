package analyse

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/linkedlist"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"

	"github.com/kkkunny/Sim/compiler/reader"

	"github.com/kkkunny/Sim/compiler/ast"

	errors "github.com/kkkunny/Sim/compiler/error"
	"github.com/kkkunny/Sim/compiler/util"
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
		return self.analyseExpr(nil, stmtNode).(hir.ExprStmt), hir.BlockEofNone
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
		if !ft.Ret.EqualTo(hir.NoThing) {
			errors.ThrowTypeMismatchError(node.Position(), ft.Ret, hir.NoThing)
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
			v.Value = optional.Some(self.expectExpr(v.Type, valueNode))
		} else {
			v.Value = optional.Some[hir.Expr](self.getTypeDefaultValue(typeNode.Position(), v.Type))
		}
	} else {
		v.Value = optional.Some(self.analyseExpr(nil, node.Value.MustValue()))
		v.Type = v.Value.MustValue().GetType()
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

	return &hir.MultiLocalVarDef{
		Vars:  vars,
		Value: value,
	}
}

func (self *Analyser) analyseIfElse(node *ast.IfElse) (*hir.IfElse, hir.BlockEof) {
	if condNode, ok := node.Cond.Value(); ok {
		cond := self.expectExpr(self.pkgScope.Bool(), condNode)
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
	cond := self.expectExpr(self.pkgScope.Bool(), node.Cond)
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
	if !hir.IsType[*hir.ArrayType](iterType) {
		errors.ThrowExpectArrayError(node.Iterator.Position(), iterType)
	}

	et := hir.AsType[*hir.ArrayType](iterType).Elem
	loop := &hir.For{
		Iterator: iterator,
		Cursor: &hir.LocalVarDef{
			VarDecl: hir.VarDecl{
				Mut:  node.CursorMut,
				Type: et,
				Name: node.Cursor.Source(),
			},
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
	vt, ok := hir.TryType[*hir.EnumType](vtObj)
	if !ok {
		errors.ThrowExpectEnumTypeError(node.Value.Position(), vtObj)
	}

	cases := linkedhashmap.NewLinkedHashMapWithCapacity[string, *hir.MatchCase](uint(len(node.Cases)))
	for _, caseNode := range node.Cases {
		caseName := caseNode.Name.Source()
		if !vt.Fields.ContainKey(caseName) {
			errors.ThrowUnknownIdentifierError(caseNode.Name.Position, caseNode.Name)
		} else if cases.ContainKey(caseName) {
			errors.ThrowIdentifierDuplicationError(caseNode.Name.Position, caseNode.Name)
		}
		caseDef := vt.Fields.Get(caseName)
		if caseDef.Elem.IsNone() && len(caseNode.Elems) != 0 {
			errors.ThrowParameterNumberNotMatchError(reader.MixPosition(caseNode.Name.Position, caseNode.ElemEnd), 0, uint(len(caseNode.Elems)))
		} else if caseDef.Elem.IsSome() && len(caseNode.Elems) != 1 {
			errors.ThrowParameterNumberNotMatchError(reader.MixPosition(caseNode.Name.Position, caseNode.ElemEnd), 1, uint(len(caseNode.Elems)))
		}
		elems := stlslices.Map(caseNode.Elems, func(i int, e ast.MatchCaseElem) *hir.Param {
			return &hir.Param{
				Mut:  e.Mutable,
				Type: caseDef.Elem.MustValue(),
				Name: util.Some(e.Name.Source()),
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
		cases.Set(caseName, &hir.MatchCase{
			Name:  caseName,
			Elems: elems,
			Body:  stlbasic.IgnoreWith(self.analyseBlock(caseNode.Body, fn)),
		})
	}

	var other util.Option[*hir.Block]
	if otherNode, ok := node.Other.Value(); ok {
		other = util.Some(stlbasic.IgnoreWith(self.analyseBlock(otherNode, nil)))
	}

	if other.IsNone() && cases.Length() != vt.Fields.Length() {
		errors.ThrowExpectMoreCase(node.Value.Position(), vtObj, cases.Length(), vt.Fields.Length())
	}

	return &hir.Match{
		Value: value,
		Cases: cases,
		Other: other,
	}, hir.BlockEofNone
}
