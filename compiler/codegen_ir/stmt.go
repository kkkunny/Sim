package codegen_ir

import (
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/hashset"
	stliter "github.com/kkkunny/stl/container/iter"
	"github.com/kkkunny/stl/container/pair"
	stlslices "github.com/kkkunny/stl/slices"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/runtime/types"
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
	case hir.Expr:
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

func (self *CodeGenerator) codegenBlock(ir *hir.Block, afterBlockCreate func(block *mir.Block)) (*mir.Block, *mir.Block) {
	from := self.builder.Current()
	block := from.Belong().NewBlock()
	if afterBlockCreate != nil {
		afterBlockCreate(block)
	}

	self.builder.MoveTo(block)
	defer self.builder.MoveTo(from)

	self.codegenFlatBlock(ir)

	return block, self.builder.Current()
}

func (self *CodeGenerator) codegenReturn(ir *hir.Return) {
	if v, ok := ir.Value.Value(); ok {
		self.builder.BuildReturn(self.codegenExpr(v, true))
	} else {
		self.builder.BuildReturn()
	}
}

func (self *CodeGenerator) codegenLocalVariable(ir *hir.LocalVarDef) mir.Value {
	t := self.codegenTypeOnly(ir.Type)
	var ptr mir.Value
	if !ir.Escaped {
		ptr = self.builder.BuildAllocFromStack(t)
	} else {
		ptr = self.buildMalloc(t)
	}
	self.values.Set(ir, ptr)
	value := self.codegenExpr(ir.Value, true)
	self.builder.BuildStore(value, ptr)
	return ptr
}

func (self *CodeGenerator) codegenMultiLocalVariable(ir *hir.MultiLocalVarDef) mir.Value {
	for _, varNode := range ir.Vars {
		self.codegenLocalVariable(varNode)
	}
	self.codegenUnTuple(ir.Value, stlslices.As[*hir.LocalVarDef, []*hir.LocalVarDef, hir.Expr, []hir.Expr](ir.Vars))
	return nil
}

func (self *CodeGenerator) codegenIfElse(ir *hir.IfElse) {
	blocks := self.codegenIfElseNode(ir)

	var brenchEndBlocks []*mir.Block
	var endBlock *mir.Block
	if ir.HasElse() {
		brenchEndBlocks = blocks
		end := stliter.All[*mir.Block](dynarray.NewDynArrayWith(brenchEndBlocks...), func(v *mir.Block) bool {
			return v.Terminated()
		})
		if end {
			return
		}
		endBlock = blocks[0].Belong().NewBlock()
	} else {
		brenchEndBlocks, endBlock = blocks[:len(blocks)-1], blocks[len(blocks)-1]
	}

	for _, brenchEndBlock := range brenchEndBlocks {
		self.builder.MoveTo(brenchEndBlock)
		self.builder.BuildUnCondJump(endBlock)
	}
	self.builder.MoveTo(endBlock)
}

func (self *CodeGenerator) codegenIfElseNode(ir *hir.IfElse) []*mir.Block {
	if condNode, ok := ir.Cond.Value(); ok {
		cond := self.codegenExpr(condNode, true)
		trueStartBlock, trueEndBlock := self.codegenBlock(ir.Body, nil)
		falseBlock := trueStartBlock.Belong().NewBlock()
		self.builder.BuildCondJump(cond, trueStartBlock, falseBlock)
		self.builder.MoveTo(falseBlock)

		if nextNode, ok := ir.Next.Value(); ok {
			blocks := self.codegenIfElseNode(nextNode)
			return append([]*mir.Block{trueEndBlock}, blocks...)
		} else {
			return []*mir.Block{trueEndBlock, falseBlock}
		}
	} else {
		self.codegenFlatBlock(ir.Body)
		return []*mir.Block{self.builder.Current()}
	}
}

type loop interface {
	SetOutBlock(block *mir.Block)
	GetOutBlock() (*mir.Block, bool)
	GetNextBlock() *mir.Block
}

type whileLoop struct {
	Cond *mir.Block
	Out       *mir.Block
}

func (self *whileLoop) SetOutBlock(block *mir.Block) {
	self.Out = block
}

func (self *whileLoop) GetOutBlock() (*mir.Block, bool) {
	return self.Out, self.Out != nil
}

func (self *whileLoop) GetNextBlock() *mir.Block {
	return self.Cond
}

func (self *CodeGenerator) codegenWhile(ir *hir.While) {
	f := self.builder.Current().Belong()
	condBlock, endBlock := f.NewBlock(),  f.NewBlock()

	self.builder.BuildUnCondJump(condBlock)
	self.builder.MoveTo(condBlock)
	cond := self.codegenExpr(ir.Cond, true)

	bodyEntryBlock, bodyEndBlock := self.codegenBlock(ir.Body, func(block *mir.Block) {
		self.loops.Set(ir, &whileLoop{Cond: condBlock, Out: endBlock})
	})
	self.builder.BuildCondJump(cond, bodyEntryBlock, endBlock)

	self.builder.MoveTo(bodyEndBlock)
	self.builder.BuildUnCondJump(condBlock)

	self.builder.MoveTo(endBlock)
}

func (self *CodeGenerator) codegenBreak(ir *hir.Break) {
	loop := self.loops.Get(ir.Loop)
	if _, ok := loop.GetOutBlock(); !ok {
		loop.SetOutBlock(self.builder.Current().Belong().NewBlock())
	}
	endBlock, _ := loop.GetOutBlock()
	self.builder.BuildUnCondJump(endBlock)
}

func (self *CodeGenerator) codegenContinue(ir *hir.Continue) {
	loop := self.loops.Get(ir.Loop)
	self.builder.BuildUnCondJump(loop.GetNextBlock())
}

type forRange struct {
	Action *mir.Block
	Out    *mir.Block
}

func (self *forRange) SetOutBlock(*mir.Block) {}

func (self *forRange) GetOutBlock() (*mir.Block, bool) {
	return self.Out, true
}

func (self *forRange) GetNextBlock() *mir.Block {
	return self.Action
}

func (self *CodeGenerator) codegenFor(ir *hir.For) {
	// pre
	size := hir.AsType[*hir.ArrayType](ir.Iterator.GetType()).Size
	iter := self.codegenExpr(ir.Iterator, false)
	indexPtr := self.builder.BuildAllocFromStack(self.ctx.Usize())
	self.builder.BuildStore(mir.NewInt(indexPtr.ElemType().(mir.IntType), 0), indexPtr)
	cursorPtr := self.codegenLocalVariable(ir.Cursor)
	condBlock := self.builder.Current().Belong().NewBlock()
	self.builder.BuildUnCondJump(condBlock)
	self.builder.MoveTo(condBlock)
	var actionBlock *mir.Block

	beforeBody := func(entry *mir.Block) {
		curBlock := self.builder.Current()
		defer func() {
			self.builder.MoveTo(curBlock)
		}()

		actionBlock = curBlock.Belong().NewBlock()
		outBlock := curBlock.Belong().NewBlock()

		// cond
		self.builder.MoveTo(condBlock)
		index := self.builder.BuildLoad(indexPtr)
		self.builder.BuildCondJump(
			self.builder.BuildCmp(mir.CmpKindLT, index, mir.NewInt(indexPtr.ElemType().(mir.IntType), int64(size))),
			entry,
			outBlock,
		)

		// action
		self.builder.MoveTo(actionBlock)
		self.builder.BuildStore(self.builder.BuildAdd(index, mir.NewInt(indexPtr.ElemType().(mir.IntType), 1)), indexPtr)
		self.builder.BuildUnCondJump(condBlock)

		// body
		self.builder.MoveTo(entry)
		self.builder.BuildStore(self.buildArrayIndex(iter, index, false), cursorPtr)

		self.loops.Set(ir, &forRange{
			Action: actionBlock,
			Out:    outBlock,
		})
	}

	_, endBlock := self.codegenBlock(ir.Body, beforeBody)
	if !endBlock.Terminated() {
		self.builder.MoveTo(endBlock)
		self.builder.BuildUnCondJump(actionBlock)
	}
	if outBlock, ok := self.loops.Get(ir).GetOutBlock(); ok {
		self.builder.MoveTo(outBlock)
	}
}

func (self *CodeGenerator) codegenMatch(ir *hir.Match) {
	if ir.Other.IsNone() && len(ir.Cases) == 0 {
		return
	}

	vtObj, vtRtObj := self.codegenType(ir.Value.GetType())
	vt, vtRt := vtObj.(mir.StructType), vtRtObj.(*types.UnionType)
	value := self.codegenExpr(ir.Value, true)
	index := self.buildStructIndex(value, 1)

	curBlock := self.builder.Current()
	endBlock := curBlock.Belong().NewBlock()

	existConds := hashset.NewHashSet[int]()
	cases := make([]pair.Pair[mir.Const, *mir.Block], 0, len(ir.Cases))
	for _, c := range ir.Cases {
		_, caseTypeRtObj := self.codegenType(c.First)
		caseIndex := vtRt.IndexElem(caseTypeRtObj)
		if existConds.Contain(caseIndex) {
			continue
		}

		existConds.Add(caseIndex)
		caseBlock, caseCurBlock := self.codegenBlock(c.Second, nil)
		self.builder.MoveTo(caseCurBlock)
		self.builder.BuildUnCondJump(endBlock)

		cases = append(cases, pair.NewPair[mir.Const, *mir.Block](mir.NewInt(vt.Elems()[1].(mir.IntType), int64(caseIndex)), caseBlock))
	}

	var otherBlock *mir.Block
	if otherIr, ok := ir.Other.Value(); ok {
		var otherCurBlock *mir.Block
		otherBlock, otherCurBlock = self.codegenBlock(otherIr, nil)
		self.builder.MoveTo(otherCurBlock)
		self.builder.BuildUnCondJump(endBlock)
	} else {
		otherBlock = endBlock
	}

	self.builder.MoveTo(curBlock)
	self.builder.BuildSwitch(index, otherBlock, cases...)

	self.builder.MoveTo(endBlock)
}
