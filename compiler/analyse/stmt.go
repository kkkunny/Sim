package analyse

import (
	"github.com/kkkunny/stl/container/linkedhashmap"
	"github.com/kkkunny/stl/container/set"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/ast"
	errors "github.com/kkkunny/Sim/compiler/error"
	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/global"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
	"github.com/kkkunny/Sim/compiler/reader"
)

func (self *Analyser) analyseFuncBody(f local.CallableDef, params []ast.Param, node *ast.Block) *local.Block {
	block := local.NewFuncBody(f)

	parent := self.scope
	self.scope = block
	defer func() {
		self.scope = parent
	}()

	switch fdef := f.(type) {
	case *global.FuncDef:
		for _, compileParam := range fdef.GenericParams() {
			self.scope.SetIdent(compileParam.String(), compileParam)
		}
	case *global.OriginMethodDef:
		selfType := stlval.TernaryAction(len(fdef.From().GenericParams()) > 0, func() hir.Type {
			return global.NewGenericCustomTypeDef(
				fdef.From(),
				stlslices.Map(fdef.From().GenericParams(), func(_ int, gp types.GenericParamType) hir.Type {
					return gp
				})...,
			)
		}, func() hir.Type {
			return fdef.From()
		})
		self.scope.SetIdent("Self", selfType)
		for _, compileParam := range fdef.From().GenericParams() {
			self.scope.SetIdent(compileParam.String(), compileParam)
		}
		for _, compileParam := range fdef.GenericParams() {
			self.scope.SetIdent(compileParam.String(), compileParam)
		}
	}

	paramNameSet := set.StdHashSetWithCap[string](uint(len(params)))
	for i, p := range f.Params() {
		if name, ok := p.GetName(); ok && (!paramNameSet.Add(name) || !self.scope.SetIdent(name, p)) {
			errors.ThrowIdentifierDuplicationError(params[i].Name.MustValue().Position, params[i].Name.MustValue())
		}
	}

	self.analyseFlatBlock(block, node)

	if block.BlockEndType() < local.BlockEndTypeFuncRet {
		retType := f.CallableType().Ret()
		if !types.Is[types.NoThingType](retType, true) {
			errors.ThrowMissingReturnValueError(node.Position(), retType)
		}
		block.PushBack(local.NewReturn())
	}
	return block
}

func (self *Analyser) analyseStmt(node ast.Stmt) hir.Local {
	switch stmtNode := node.(type) {
	case *ast.Return:
		ret := self.analyseReturn(stmtNode)
		return ret
	case *ast.SingleVariableDef:
		return self.analyseSingleLocalVariable(stmtNode)
	case *ast.MultipleVariableDef:
		return self.analyseLocalMultiVariable(stmtNode)
	case *ast.Block:
		return self.analyseBlock(stmtNode, nil)
	case *ast.IfElse:
		return self.analyseIfElse(stmtNode)
	case ast.Expr:
		return local.NewExpr(self.analyseExpr(nil, stmtNode))
	case *ast.While:
		return self.analyseWhile(stmtNode)
	case *ast.Break:
		return self.analyseBreak(stmtNode)
	case *ast.Continue:
		return self.analyseContinue(stmtNode)
	case *ast.For:
		return self.analyseFor(stmtNode)
	case *ast.Match:
		return self.analyseMatch(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *Analyser) analyseReturn(node *ast.Return) *local.Return {
	c := self.scope.(*local.Block).Belong()
	ct := c.CallableType()
	if v, ok := node.Value.Value(); ok {
		value := self.expectExpr(ct.Ret(), v)
		return local.NewReturn(value)
	} else {
		if !types.Is[types.NoThingType](ct.Ret()) {
			errors.ThrowTypeMismatchError(node.Position(), ct.Ret(), types.NoThing)
		}
		return local.NewReturn()
	}
}

func (self *Analyser) analyseSingleLocalVariable(node *ast.SingleVariableDef) *local.SingleVarDef {
	name := node.Var.Name.Source()

	var t hir.Type
	var v hir.Value
	if typeNode, ok := node.Var.Type.Value(); ok {
		t = self.analyseType(typeNode)
		if valueNode, ok := node.Value.Value(); ok {
			v = self.expectExpr(t, valueNode)
		} else {
			v = self.getTypeDefaultValue(typeNode.Position(), t)
		}
	} else {
		v = self.analyseExpr(nil, node.Value.MustValue())
		t = v.Type()
	}

	ident := local.NewSingleVarDef(node.Var.Mutable, name, t, v)
	if !self.scope.SetIdent(name, ident) {
		errors.ThrowIdentifierDuplicationError(node.Var.Name.Position, node.Var.Name)
	}
	return ident
}

func (self *Analyser) analyseLocalMultiVariable(node *ast.MultipleVariableDef) *local.MultiVarDef {
	allHaveType := true
	varTypes := stlslices.Map(node.Vars, func(_ int, vNode ast.VarDef) hir.Type {
		typeNode, ok := vNode.Type.Value()
		if !ok {
			allHaveType = false
			return nil
		}
		return self.analyseType(typeNode)
	})

	var value hir.Value
	if valueNode, ok := node.Value.Value(); ok && allHaveType {
		value = self.expectExpr(types.NewTupleType(varTypes...), valueNode)
	} else if ok && stlval.Is[*ast.Tuple](valueNode) {
		valueNodes := valueNode.(*ast.Tuple).Elems
		if len(valueNodes) != len(varTypes) {
			errors.ThrowParameterNumberNotMatchError(valueNode.Position(), uint(len(varTypes)), uint(len(valueNodes)))
		}
		elems := stlslices.Map(valueNodes, func(i int, vNode ast.Expr) hir.Value {
			vt := varTypes[i]
			if vt == nil {
				v := self.analyseExpr(nil, vNode)
				varTypes[i] = v.Type()
				return v
			} else {
				return self.expectExpr(vt, vNode)
			}
		})
		value = local.NewTupleExpr(elems...)
	} else if ok {
		value = self.analyseExpr(types.NewTupleType(varTypes...), valueNode)
		vTt, ok := types.As[types.TupleType](value.Type())
		if !ok {
			errors.ThrowExpectTupleError(valueNode.Position(), value.Type())
		} else if len(vTt.Elems()) != len(varTypes) {
			errors.ThrowParameterNumberNotMatchError(valueNode.Position(), uint(len(varTypes)), uint(len(vTt.Elems())))
		}
		for i, vt := range varTypes {
			if vt == nil {
				varTypes[i] = vTt.Elems()[i]
			} else if !vt.Equal(vTt.Elems()[i]) {
				errors.ThrowTypeMismatchError(node.Vars[i].Name.Position, vt, vTt.Elems()[i])
			}
		}
	} else {
		elems := stlslices.Map(varTypes, func(i int, vt hir.Type) hir.Value {
			return self.getTypeDefaultValue(node.Vars[i].Type.MustValue().Position(), vt)
		})
		value = local.NewTupleExpr(elems...)
	}

	vars := stlslices.Map(node.Vars, func(i int, vNode ast.VarDef) values.VarDecl {
		name := vNode.Name.Source()
		v := values.NewVarDecl(vNode.Mutable, name, varTypes[i])
		if !self.scope.SetIdent(name, v) {
			errors.ThrowIdentifierDuplicationError(vNode.Name.Position, vNode.Name)
		}
		return v
	})
	return local.NewMultiVarDef(vars, value)
}

func (self *Analyser) analyseFlatBlock(block *local.Block, node *ast.Block) {
	for iter := node.Stmts.Iterator(); iter.Next(); {
		stmt := self.analyseStmt(iter.Value())
		if stmtBlock, ok := stmt.(*local.Block); ok {
			block.Append(stmtBlock)
		} else {
			block.PushBack(stmt)
		}
	}
}

func (self *Analyser) analyseBlock(node *ast.Block, onAfterCreate func(block *local.Block)) *local.Block {
	block := local.NewBlock(self.scope.(*local.Block))
	if onAfterCreate != nil {
		onAfterCreate(block)
	}

	parent := self.scope
	self.scope = block
	defer func() {
		self.scope = parent
	}()

	self.analyseFlatBlock(block, node)
	return block
}

func (self *Analyser) analyseIfElse(node *ast.IfElse) *local.IfElse {
	if condNode, ok := node.Cond.Value(); ok {
		cond := self.expectExpr(types.Bool, condNode)
		body := self.analyseBlock(node.Body, nil)

		ifStmt := local.NewIfElse(body, cond)

		if nextNode, ok := node.Next.Value(); ok {
			ifStmt.SetNext(self.analyseIfElse(nextNode))
		}
		return ifStmt
	} else {
		body := self.analyseBlock(node.Body, nil)
		return local.NewIfElse(body)
	}
}

func (self *Analyser) analyseWhile(node *ast.While) *local.While {
	cond := self.expectExpr(types.Bool, node.Cond)
	var loop *local.While
	self.analyseBlock(node.Body, func(block *local.Block) {
		loop = local.NewWhile(cond, block)
		block.SetLoop(loop)
	})
	return loop
}

func (self *Analyser) analyseFor(node *ast.For) *local.For {
	iter := self.analyseExpr(nil, node.Iterator)
	iterType := iter.Type()
	at, ok := types.As[types.ArrayType](iterType)
	if !ok {
		errors.ThrowExpectArrayError(node.Iterator.Position(), iterType)
	}

	cursorName := node.Cursor.Source()
	cursor := values.NewVarDecl(node.CursorMut, cursorName, at.Elem())
	var loop *local.For
	self.analyseBlock(node.Body, func(block *local.Block) {
		if !block.SetIdent(cursorName, cursor) {
			errors.ThrowIdentifierDuplicationError(node.Cursor.Position, node.Cursor)
		}
		loop = local.NewFor(cursor, iter, block)
		block.SetLoop(loop)
	})
	return loop
}

func (self *Analyser) analyseMatch(node *ast.Match) *local.Match {
	value := self.analyseExpr(nil, node.Value)
	vt := value.Type()
	vEt, ok := types.As[types.EnumType](vt)
	if !ok {
		errors.ThrowExpectEnumTypeError(node.Value.Position(), vt)
	}

	cases := linkedhashmap.StdWithCap[string, *local.MatchCase](uint(len(node.Cases)))
	for _, caseNode := range node.Cases {
		caseName := caseNode.Name.Source()
		if !vEt.EnumFields().Contain(caseName) {
			errors.ThrowUnknownIdentifierError(caseNode.Name.Position, caseNode.Name)
		} else if cases.Contain(caseName) {
			errors.ThrowIdentifierDuplicationError(caseNode.Name.Position, caseNode.Name)
		}

		field := vEt.EnumFields().Get(caseName)
		fieldElem, ok := field.Elem()
		if !ok && caseNode.Elem.IsSome() {
			errors.ThrowParameterNumberNotMatchError(reader.MixPosition(caseNode.Name.Position, caseNode.ElemEnd), 0, 1)
		} else if ok && caseNode.Elem.IsNone() {
			errors.ThrowParameterNumberNotMatchError(reader.MixPosition(caseNode.Name.Position, caseNode.ElemEnd), 1, 0)
		}

		var caseVar values.VarDecl
		var caseBodyFn func(block *local.Block)
		if caseVarNode, ok := caseNode.Elem.Value(); ok {
			caseVarName := caseVarNode.Name.Source()
			caseVar = values.NewVarDecl(caseVarNode.Mutable, caseVarName, fieldElem)
			caseBodyFn = func(block *local.Block) {
				if !block.SetIdent(caseVarName, caseVar) {
					errors.ThrowIdentifierDuplicationError(caseVarNode.Name.Position, caseVarNode.Name)
				}
			}
		}

		body := self.analyseBlock(caseNode.Body, caseBodyFn)
		cases.Set(caseName, local.NewMatchCase(caseName, caseVar, body))
	}

	stmt := local.NewMatch(value, cases.Values()...)

	if otherNode, ok := node.Other.Value(); ok {
		stmt.SetOther(self.analyseBlock(otherNode, nil))
	} else if cases.Length() != vEt.EnumFields().Length() {
		errors.ThrowExpectMoreCase(node.Value.Position(), vt, cases.Length(), vEt.EnumFields().Length())
	}

	return stmt
}

func (self *Analyser) analyseContinue(node *ast.Continue) *local.Continue {
	loop, ok := self.scope.(*local.Block).Loop()
	if !ok {
		errors.ThrowLoopControlError(node.Position())
	}
	return local.NewContinue(loop)
}

func (self *Analyser) analyseBreak(node *ast.Break) *local.Break {
	loop, ok := self.scope.(*local.Block).Loop()
	if !ok {
		errors.ThrowLoopControlError(node.Position())
	}
	return local.NewBreak(loop)
}
