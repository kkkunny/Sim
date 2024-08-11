package codegen_ir

import (
	"slices"

	"github.com/kkkunny/go-llvm"
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"

	"github.com/kkkunny/Sim/compiler/hir"
)

func (self *CodeGenerator) codegenStmt(ir hir.Stmt) {
	switch stmt := ir.(type) {
	case *hir.Return:
		self.codegenReturn(stmt)
	case *hir.LocalVarDef:
		self.codegenLocalVariable(stmt)
	case *hir.MultiLocalVarDef:
		self.codegenMultiLocalVariable(stmt)
	case *hir.IfElse:
		self.codegenIfElse(stmt)
	case hir.ExprStmt:
		self.codegenExpr(stmt, false)
	case *hir.While:
		self.codegenWhile(stmt)
	case *hir.Break:
		self.codegenBreak(stmt)
	case *hir.Continue:
		self.codegenContinue(stmt)
	case *hir.For:
		self.codegenFor(stmt)
	case *hir.Match:
		self.codegenMatch(stmt)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFlatBlock(ir *hir.Block) {
	for iter := ir.Stmts.Iterator(); iter.Next(); {
		self.codegenStmt(iter.Value())
	}
}

func (self *CodeGenerator) codegenBlock(ir *hir.Block, afterBlockCreate func(block llvm.Block)) (llvm.Block, llvm.Block) {
	from := self.builder.CurrentBlock()
	block := from.Belong().NewBlock("")
	if afterBlockCreate != nil {
		afterBlockCreate(block)
	}

	self.builder.MoveToAfter(block)
	defer self.builder.MoveToAfter(from)

	self.codegenFlatBlock(ir)

	return block, self.builder.CurrentBlock()
}

func (self *CodeGenerator) codegenReturn(ir *hir.Return) {
	if vir, ok := ir.Value.Value(); ok {
		v := self.codegenExpr(vir, true)
		if vir.GetType().EqualTo(hir.NoThing) || vir.GetType().EqualTo(hir.NoReturn) {
			self.buildReturn(nil)
		} else {
			self.buildReturn(vir.GetType(), v)
		}
	} else {
		self.buildReturn(nil)
	}
}

func (self *CodeGenerator) codegenLocalVariable(ir *hir.LocalVarDef) llvm.Value {
	t := self.codegenType(ir.Type)
	var ptr llvm.Value
	if !ir.Escaped {
		ptr = self.builder.CreateAlloca("", t)
	} else {
		ptr = self.buildMalloc(t)
	}
	self.values.Set(ir, ptr)
	if valueIr, ok := ir.Value.Value(); ok {
		value := self.codegenExpr(valueIr, true)
		self.buildStore(ir.Type, value, ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenMultiLocalVariable(ir *hir.MultiLocalVarDef) llvm.Value {
	for _, varNode := range ir.Vars {
		self.codegenLocalVariable(varNode)
	}
	self.codegenUnTuple(ir.Value, stlslices.As[*hir.LocalVarDef, hir.Expr](ir.Vars))
	return nil
}

func (self *CodeGenerator) codegenIfElse(ir *hir.IfElse) {
	blocks := self.codegenIfElseNode(ir)

	var brenchEndBlocks []llvm.Block
	var endBlock llvm.Block
	if ir.HasElse() {
		brenchEndBlocks = blocks
		end := stlslices.All(brenchEndBlocks, func(i int, v llvm.Block) bool {
			return v.IsTerminating()
		})
		if end {
			return
		}
		endBlock = blocks[0].Belong().NewBlock("")
	} else {
		brenchEndBlocks, endBlock = blocks[:len(blocks)-1], stlslices.Last(blocks)
	}

	for _, brenchEndBlock := range brenchEndBlocks {
		self.builder.MoveToAfter(brenchEndBlock)
		self.builder.CreateBr(endBlock)
	}
	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) codegenIfElseNode(ir *hir.IfElse) []llvm.Block {
	if condNode, ok := ir.Cond.Value(); ok {
		cond := self.builder.CreateTrunc("", self.codegenExpr(condNode, true), self.builder.BooleanType())
		trueStartBlock, trueEndBlock := self.codegenBlock(ir.Body, nil)
		falseBlock := trueStartBlock.Belong().NewBlock("")
		self.builder.CreateCondBr(cond, trueStartBlock, falseBlock)
		self.builder.MoveToAfter(falseBlock)

		if nextNode, ok := ir.Next.Value(); ok {
			blocks := self.codegenIfElseNode(nextNode)
			return append([]llvm.Block{trueEndBlock}, blocks...)
		} else {
			return []llvm.Block{trueEndBlock, falseBlock}
		}
	} else {
		self.codegenFlatBlock(ir.Body)
		return []llvm.Block{self.builder.CurrentBlock()}
	}
}

type loop interface {
	SetOutBlock(block llvm.Block)
	GetOutBlock() (llvm.Block, bool)
	GetNextBlock() llvm.Block
}

type whileLoop struct {
	Cond llvm.Block
	Out  optional.Optional[llvm.Block]
}

func (self *whileLoop) SetOutBlock(block llvm.Block) {
	self.Out = optional.Some(block)
}

func (self *whileLoop) GetOutBlock() (llvm.Block, bool) {
	if self.Out.IsNone() {
		return llvm.Block{}, false
	}
	return self.Out.MustValue(), true
}

func (self *whileLoop) GetNextBlock() llvm.Block {
	return self.Cond
}

func (self *CodeGenerator) codegenWhile(ir *hir.While) {
	f := self.builder.CurrentFunction()
	condBlock, endBlock := f.NewBlock(""), f.NewBlock("")

	self.builder.CreateBr(condBlock)
	self.builder.MoveToAfter(condBlock)
	cond := self.builder.CreateTrunc("", self.codegenExpr(ir.Cond, true), self.builder.BooleanType())

	bodyEntryBlock, bodyEndBlock := self.codegenBlock(ir.Body, func(block llvm.Block) {
		self.loops.Set(ir, &whileLoop{Cond: condBlock, Out: optional.Some(endBlock)})
	})
	self.builder.CreateCondBr(cond, bodyEntryBlock, endBlock)

	self.builder.MoveToAfter(bodyEndBlock)
	self.builder.CreateBr(condBlock)

	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) codegenBreak(ir *hir.Break) {
	loop := self.loops.Get(ir.Loop)
	if _, ok := loop.GetOutBlock(); !ok {
		loop.SetOutBlock(self.builder.CurrentFunction().NewBlock(""))
	}
	endBlock, _ := loop.GetOutBlock()
	self.builder.CreateBr(endBlock)
}

func (self *CodeGenerator) codegenContinue(ir *hir.Continue) {
	loop := self.loops.Get(ir.Loop)
	self.builder.CreateBr(loop.GetNextBlock())
}

type forRange struct {
	Action llvm.Block
	Out    llvm.Block
}

func (self *forRange) SetOutBlock(llvm.Block) {}

func (self *forRange) GetOutBlock() (llvm.Block, bool) {
	return self.Out, true
}

func (self *forRange) GetNextBlock() llvm.Block {
	return self.Action
}

func (self *CodeGenerator) codegenFor(ir *hir.For) {
	// pre
	size := hir.AsType[*hir.ArrayType](ir.Iterator.GetType()).Size
	iter := self.codegenExpr(ir.Iterator, false)
	indexPtr := self.builder.CreateAlloca("", self.builder.IntPtrType())
	self.builder.CreateStore(self.builder.ConstIntPtr(0), indexPtr)
	cursorPtr := self.codegenLocalVariable(ir.Cursor)
	condBlock := self.builder.CurrentFunction().NewBlock("")
	self.builder.CreateBr(condBlock)
	self.builder.MoveToAfter(condBlock)
	var actionBlock llvm.Block

	beforeBody := func(entry llvm.Block) {
		curBlock := self.builder.CurrentBlock()
		defer func() {
			self.builder.MoveToAfter(curBlock)
		}()

		actionBlock = curBlock.Belong().NewBlock("")
		outBlock := curBlock.Belong().NewBlock("")

		// cond
		self.builder.MoveToAfter(condBlock)
		index := self.builder.CreateLoad("", self.builder.IntPtrType(), indexPtr)
		self.builder.CreateCondBr(
			self.builder.CreateIntCmp("", llvm.IntULT, index, self.builder.ConstIntPtr(int64(size))),
			entry,
			outBlock,
		)

		// action
		self.builder.MoveToAfter(actionBlock)
		self.builder.CreateStore(self.builder.CreateUAdd("", index, self.builder.ConstIntPtr(1)), indexPtr)
		self.builder.CreateBr(condBlock)

		// body
		self.builder.MoveToAfter(entry)
		self.builder.CreateStore(self.buildArrayIndex(iter.Type().(llvm.ArrayType), iter, index, false), cursorPtr)

		self.loops.Set(ir, &forRange{
			Action: actionBlock,
			Out:    outBlock,
		})
	}

	_, endBlock := self.codegenBlock(ir.Body, beforeBody)
	if !endBlock.IsTerminating() {
		self.builder.MoveToAfter(endBlock)
		self.builder.CreateBr(actionBlock)
	}
	if outBlock, ok := self.loops.Get(ir).GetOutBlock(); ok {
		self.builder.MoveToAfter(outBlock)
	}
}

func (self *CodeGenerator) codegenMatch(ir *hir.Match) {
	if ir.Other.IsNone() && ir.Cases.Empty() {
		return
	}

	etIr := hir.AsType[*hir.EnumType](ir.Value.GetType())
	t := self.codegenType(ir.Value.GetType())
	value := self.codegenExpr(ir.Value, false)
	index := stlbasic.TernaryAction(etIr.IsSimple(), func() llvm.Value {
		if value.Type().Equal(t) {
			return value
		} else {
			return self.builder.CreateLoad("", t, value)
		}
	}, func() llvm.Value {
		return self.buildStructIndex(t.(llvm.StructType), value, 1, false)
	})

	curBlock := self.builder.CurrentBlock()
	endBlock := curBlock.Belong().NewBlock("")

	cases := make([]struct {
		Value llvm.Value
		Block llvm.Block
	}, 0, ir.Cases.Length())
	for iter := ir.Cases.Iterator(); iter.Next(); {
		caseIndex := slices.Index(etIr.Fields.Keys().ToSlice(), iter.Value().First)
		caseBlock, caseCurBlock := self.codegenBlock(iter.Value().Second.Body, func(block llvm.Block) {
			if stlslices.Empty(iter.Value().Second.Elems) {
				return
			}

			caseCurBlock := self.builder.CurrentBlock()
			defer self.builder.MoveToAfter(caseCurBlock)
			self.builder.MoveToAfter(block)

			caseType := self.codegenTupleType(hir.NewTupleType(stlslices.Map(iter.Value().Second.Elems, func(_ int, e *hir.Param) hir.Type {
				return e.GetType()
			})...))
			for i, elem := range iter.Value().Second.Elems {
				self.values.Set(elem, self.buildStructIndex(caseType, self.buildStructIndex(t.(llvm.StructType), value, 0, true), uint(i), true))
			}
		})
		self.builder.MoveToAfter(caseCurBlock)
		self.builder.CreateBr(endBlock)

		cases = append(cases, struct {
			Value llvm.Value
			Block llvm.Block
		}{Value: self.builder.ConstInteger(index.Type().(llvm.IntegerType), int64(caseIndex)), Block: caseBlock})
	}

	var otherBlock llvm.Block
	if otherIr, ok := ir.Other.Value(); ok {
		var otherCurBlock llvm.Block
		otherBlock, otherCurBlock = self.codegenBlock(otherIr, nil)
		self.builder.MoveToAfter(otherCurBlock)
		self.builder.CreateBr(endBlock)
	} else {
		otherBlock = endBlock
	}

	self.builder.MoveToAfter(curBlock)
	self.builder.CreateSwitch(index, otherBlock, cases...)

	self.builder.MoveToAfter(endBlock)
}
