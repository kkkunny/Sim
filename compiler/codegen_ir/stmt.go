package codegen_ir

import (
	"slices"

	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/oldhir"
)

func (self *CodeGenerator) codegenStmt(ir oldhir.Stmt) {
	switch stmt := ir.(type) {
	case *oldhir.Return:
		self.codegenReturn(stmt)
	case *oldhir.LocalVarDef:
		self.codegenLocalVariable(stmt)
	case *oldhir.MultiLocalVarDef:
		self.codegenMultiLocalVariable(stmt)
	case *oldhir.IfElse:
		self.codegenIfElse(stmt)
	case oldhir.ExprStmt:
		self.codegenExpr(stmt, false)
	case *oldhir.While:
		self.codegenWhile(stmt)
	case *oldhir.Break:
		self.codegenBreak(stmt)
	case *oldhir.Continue:
		self.codegenContinue(stmt)
	case *oldhir.For:
		self.codegenFor(stmt)
	case *oldhir.Match:
		self.codegenMatch(stmt)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFlatBlock(ir *oldhir.Block) {
	for iter := ir.Stmts.Iterator(); iter.Next(); {
		self.codegenStmt(iter.Value())
	}
}

func (self *CodeGenerator) codegenBlock(ir *oldhir.Block, afterBlockCreate func(block llvm.Block)) (llvm.Block, llvm.Block) {
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

func (self *CodeGenerator) codegenReturn(ir *oldhir.Return) {
	if vir, ok := ir.Value.Value(); ok {
		v := self.codegenExpr(vir, true)
		if vir.GetType().EqualTo(oldhir.NoThing) || vir.GetType().EqualTo(oldhir.NoReturn) {
			self.builder.CreateRet(nil)
		} else {
			self.builder.CreateRet(&v)
		}
	} else {
		self.builder.CreateRet(nil)
	}
}

func (self *CodeGenerator) codegenLocalVariable(ir *oldhir.LocalVarDef) llvm.Value {
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
		self.builder.CreateStore(value, ptr)
	}
	return ptr
}

func (self *CodeGenerator) codegenMultiLocalVariable(ir *oldhir.MultiLocalVarDef) llvm.Value {
	for _, varNode := range ir.Vars {
		self.codegenLocalVariable(varNode)
	}
	self.codegenUnTuple(ir.Value, stlslices.As[*oldhir.LocalVarDef, oldhir.Expr](ir.Vars))
	return nil
}

func (self *CodeGenerator) codegenIfElse(ir *oldhir.IfElse) {
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

func (self *CodeGenerator) codegenIfElseNode(ir *oldhir.IfElse) []llvm.Block {
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

func (self *CodeGenerator) codegenWhile(ir *oldhir.While) {
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

func (self *CodeGenerator) codegenBreak(ir *oldhir.Break) {
	loop := self.loops.Get(ir.Loop)
	if _, ok := loop.GetOutBlock(); !ok {
		loop.SetOutBlock(self.builder.CurrentFunction().NewBlock(""))
	}
	endBlock, _ := loop.GetOutBlock()
	self.builder.CreateBr(endBlock)
}

func (self *CodeGenerator) codegenContinue(ir *oldhir.Continue) {
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

func (self *CodeGenerator) codegenFor(ir *oldhir.For) {
	// pre
	size := oldhir.AsType[*oldhir.ArrayType](ir.Iterator.GetType()).Size
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

func (self *CodeGenerator) codegenMatch(ir *oldhir.Match) {
	if ir.Other.IsNone() && ir.Cases.Empty() {
		return
	}

	etIr := oldhir.AsType[*oldhir.EnumType](ir.Value.GetType())
	t := self.codegenType(ir.Value.GetType())
	value := self.codegenExpr(ir.Value, false)
	index := stlval.TernaryAction(etIr.IsSimple(), func() llvm.Value {
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
		caseIndex := slices.Index(etIr.Fields.Keys(), iter.Value().E1())
		caseBlock, caseCurBlock := self.codegenBlock(iter.Value().E2().Body, func(block llvm.Block) {
			if stlslices.Empty(iter.Value().E2().Elems) {
				return
			}

			caseCurBlock := self.builder.CurrentBlock()
			defer self.builder.MoveToAfter(caseCurBlock)
			self.builder.MoveToAfter(block)

			caseType := self.codegenTupleType(oldhir.NewTupleType(stlslices.Map(iter.Value().E2().Elems, func(_ int, e *oldhir.Param) oldhir.Type {
				return e.GetType()
			})...))
			for i, elem := range iter.Value().E2().Elems {
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
