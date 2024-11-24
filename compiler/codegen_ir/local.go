package codegen_ir

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/optional"
	stlslices "github.com/kkkunny/stl/container/slices"
	stlval "github.com/kkkunny/stl/value"

	"github.com/kkkunny/Sim/compiler/hir"
	"github.com/kkkunny/Sim/compiler/hir/local"
	"github.com/kkkunny/Sim/compiler/hir/types"
	"github.com/kkkunny/Sim/compiler/hir/values"
)

func (self *CodeGenerator) codegenLocal(ir hir.Local) {
	switch ir := ir.(type) {
	case *local.Return:
		self.codegenReturn(ir)
	case *local.SingleVarDef:
		self.codegenLocalVariable(ir)
	case *local.MultiVarDef:
		self.codegenMultiLocalVariable(ir)
	case *local.IfElse:
		self.codegenIfElse(ir)
	case *local.While:
		self.codegenWhile(ir)
	case *local.Break:
		self.codegenBreak(ir)
	case *local.Continue:
		self.codegenContinue(ir)
	case *local.For:
		self.codegenFor(ir)
	case *local.Match:
		self.codegenMatch(ir)
	case *local.Expr:
		self.codegenValue(ir.Value(), false)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFlatBlock(ir *local.Block) {
	for iter := ir.Stmts().Iterator(); iter.Next(); {
		self.codegenLocal(iter.Value())
	}
}

func (self *CodeGenerator) codegenBlock(ir *local.Block, afterBlockCreate func(block llvm.Block)) (llvm.Block, llvm.Block) {
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

func (self *CodeGenerator) codegenReturn(ir *local.Return) {
	if vIr, ok := ir.Value(); ok {
		v := self.codegenValue(vIr, true)
		if types.Is[types.NoThingType](vIr.Type(), true) || types.Is[types.NoReturnType](vIr.Type(), true) {
			self.builder.CreateRet(nil)
		} else {
			self.builder.CreateRet(&v)
		}
	} else {
		self.builder.CreateRet(nil)
	}
}

func (self *CodeGenerator) codegenVarDecl(ir values.VarDecl) llvm.Value {
	t := self.codegenType(ir.Type())
	ptr := self.builder.CreateAlloca("", t)
	self.values.Set(ir, ptr)
	return ptr
}

func (self *CodeGenerator) codegenLocalVariable(ir *local.SingleVarDef) llvm.Value {
	t := self.codegenType(ir.Type())
	var ptr llvm.Value
	if !ir.Escaped() {
		ptr = self.builder.CreateAlloca("", t)
	} else {
		ptr = self.buildMalloc(t)
	}
	self.values.Set(ir, ptr)
	value := self.codegenValue(ir.Value(), true)
	self.builder.CreateStore(value, ptr)
	return ptr
}

func (self *CodeGenerator) codegenMultiLocalVariable(ir *local.MultiVarDef) llvm.Value {
	for _, varIr := range ir.Vars() {
		self.codegenVarDecl(varIr)
	}
	self.codegenUnTuple(ir.Value(), stlslices.As[values.VarDecl, hir.Value](ir.Vars()))
	return nil
}

func (self *CodeGenerator) codegenIfElse(ir *local.IfElse) {
	blocks := self.codegenIfElseNode(ir)

	var brenchEndBlocks []llvm.Block
	var endBlock llvm.Block
	if _, ok := ir.Next(); ok {
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

func (self *CodeGenerator) codegenIfElseNode(ir *local.IfElse) []llvm.Block {
	if condNode, ok := ir.Cond(); ok {
		cond := self.builder.CreateTrunc("", self.codegenValue(condNode, true), self.builder.BooleanType())
		trueStartBlock, trueEndBlock := self.codegenBlock(ir.Body(), nil)
		falseBlock := trueStartBlock.Belong().NewBlock("")
		self.builder.CreateCondBr(cond, trueStartBlock, falseBlock)
		self.builder.MoveToAfter(falseBlock)

		if nextNode, ok := ir.Next(); ok {
			blocks := self.codegenIfElseNode(nextNode)
			return append([]llvm.Block{trueEndBlock}, blocks...)
		} else {
			return []llvm.Block{trueEndBlock, falseBlock}
		}
	} else {
		self.codegenFlatBlock(ir.Body())
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

func (self *CodeGenerator) codegenWhile(ir *local.While) {
	f := self.builder.CurrentFunction()
	condBlock, endBlock := f.NewBlock(""), f.NewBlock("")

	self.builder.CreateBr(condBlock)
	self.builder.MoveToAfter(condBlock)
	cond := self.builder.CreateTrunc("", self.codegenValue(ir.Cond(), true), self.builder.BooleanType())

	bodyEntryBlock, bodyEndBlock := self.codegenBlock(ir.Body(), func(block llvm.Block) {
		self.loops.Set(ir, &whileLoop{Cond: condBlock, Out: optional.Some(endBlock)})
	})
	self.builder.CreateCondBr(cond, bodyEntryBlock, endBlock)

	self.builder.MoveToAfter(bodyEndBlock)
	self.builder.CreateBr(condBlock)

	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) codegenBreak(ir *local.Break) {
	loop := self.loops.Get(ir.Loop())
	if _, ok := loop.GetOutBlock(); !ok {
		loop.SetOutBlock(self.builder.CurrentFunction().NewBlock(""))
	}
	endBlock, _ := loop.GetOutBlock()
	self.builder.CreateBr(endBlock)
}

func (self *CodeGenerator) codegenContinue(ir *local.Continue) {
	loop := self.loops.Get(ir.Loop())
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

func (self *CodeGenerator) codegenFor(ir *local.For) {
	// pre
	size := stlval.IgnoreWith(types.As[types.ArrayType](ir.Iter().Type())).Size()
	iter := self.codegenValue(ir.Iter(), false)
	indexPtr := self.builder.CreateAlloca("", self.builder.Isize())
	self.builder.CreateStore(self.builder.ConstIsize(0), indexPtr)
	cursorPtr := self.codegenVarDecl(ir.Cursor())
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
		index := self.builder.CreateLoad("", self.builder.Isize(), indexPtr)
		self.builder.CreateCondBr(
			self.builder.CreateIntCmp("", llvm.IntULT, index, self.builder.ConstIsize(int64(size))),
			entry,
			outBlock,
		)

		// action
		self.builder.MoveToAfter(actionBlock)
		self.builder.CreateStore(self.builder.CreateUAdd("", index, self.builder.ConstIsize(1)), indexPtr)
		self.builder.CreateBr(condBlock)

		// body
		self.builder.MoveToAfter(entry)
		self.builder.CreateStore(self.builder.CreateArrayIndex(iter.Type().(llvm.ArrayType), iter, index, false), cursorPtr)

		self.loops.Set(ir, &forRange{
			Action: actionBlock,
			Out:    outBlock,
		})
	}

	_, endBlock := self.codegenBlock(ir.Body(), beforeBody)
	if !endBlock.IsTerminating() {
		self.builder.MoveToAfter(endBlock)
		self.builder.CreateBr(actionBlock)
	}
	if outBlock, ok := self.loops.Get(ir).GetOutBlock(); ok {
		self.builder.MoveToAfter(outBlock)
	}
}

func (self *CodeGenerator) codegenMatch(ir *local.Match) {
	if _, ok := ir.Other(); !ok && len(ir.Cases()) == 0 {
		return
	}

	type casePair = struct {
		Value llvm.Value
		Block llvm.Block
	}

	etIr := stlval.IgnoreWith(types.As[types.EnumType](ir.Cond().Type()))
	t := self.codegenType(ir.Cond().Type())
	value := self.codegenValue(ir.Cond(), false)
	index := stlval.TernaryAction(etIr.Simple(), func() llvm.Value {
		if value.Type().Equal(t) {
			return value
		} else {
			return self.builder.CreateLoad("", t, value)
		}
	}, func() llvm.Value {
		return self.builder.CreateStructIndex(t.(llvm.StructType), value, 1, false)
	})

	curBlock := self.builder.CurrentBlock()
	endBlock := curBlock.Belong().NewBlock("")

	cases := make([]casePair, 0, len(ir.Cases()))
	for _, caseIr := range ir.Cases() {
		var i int
		for j, name := range etIr.EnumFields().Keys() {
			if name == caseIr.Name() {
				i = j
			}
		}
		caseBlock, caseCurBlock := self.codegenBlock(caseIr.Body(), func(block llvm.Block) {
			caseVar, ok := caseIr.Var()
			if !ok {
				return
			}

			caseCurBlock := self.builder.CurrentBlock()
			defer self.builder.MoveToAfter(caseCurBlock)
			self.builder.MoveToAfter(block)

			self.values.Set(caseVar, self.builder.CreateStructIndex(t.(llvm.StructType), value, 0, true))
		})
		self.builder.MoveToAfter(caseCurBlock)
		self.builder.CreateBr(endBlock)

		cases = append(cases, casePair{Value: self.builder.ConstInteger(index.Type().(llvm.IntegerType), int64(i)), Block: caseBlock})
	}

	var otherBlock llvm.Block
	if otherIr, ok := ir.Other(); ok {
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
	if stlslices.All(ir.Cases(), func(_ int, c *local.MatchCase) bool {
		return c.Body().BlockEndType() == local.BlockEndTypeFuncRet
	}) {
		self.builder.CreateUnreachable()
	}
}
