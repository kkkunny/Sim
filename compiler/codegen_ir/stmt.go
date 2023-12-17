package codegen_ir

import (
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/iterator"

	"github.com/kkkunny/Sim/hir"
	"github.com/kkkunny/Sim/mir"
)

func (self *CodeGenerator) codegenStmt(node hir.Stmt) {
	switch stmtNode := node.(type) {
	case *hir.Return:
		self.codegenReturn(stmtNode)
	case *hir.VarDef:
		self.codegenLocalVariable(stmtNode)
	case *hir.IfElse:
		self.codegenIfElse(stmtNode)
	case hir.Expr:
		self.codegenExpr(stmtNode, false)
	case *hir.EndlessLoop:
		self.codegenEndlessLoop(stmtNode)
	case *hir.Break:
		self.codegenBreak(stmtNode)
	case *hir.Continue:
		self.codegenContinue(stmtNode)
	case *hir.For:
		self.codegenFor(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFlatBlock(node *hir.Block) {
	for iter := node.Stmts.Iterator(); iter.Next(); {
		self.codegenStmt(iter.Value())
	}
}

func (self *CodeGenerator) codegenBlock(node *hir.Block, afterBlockCreate func(block *mir.Block)) (*mir.Block, *mir.Block) {
	from := self.builder.Current()
	block := from.Belong().NewBlock()
	if afterBlockCreate != nil {
		afterBlockCreate(block)
	}

	self.builder.MoveTo(block)
	defer self.builder.MoveTo(from)

	self.codegenFlatBlock(node)

	return block, self.builder.Current()
}

func (self *CodeGenerator) codegenReturn(node *hir.Return) {
	if v, ok := node.Value.Value(); ok {
		self.builder.BuildReturn(self.codegenExpr(v, true))
	} else {
		self.builder.BuildReturn()
	}
}

func (self *CodeGenerator) codegenLocalVariable(node *hir.VarDef) mir.Value {
	value := self.codegenExpr(node.Value, true)
	ptr := self.builder.BuildAllocFromStack(self.codegenType(node.Type))
	self.builder.BuildStore(value, ptr)
	self.values.Set(node, ptr)
	return ptr
}

func (self *CodeGenerator) codegenIfElse(node *hir.IfElse) {
	blocks := self.codegenIfElseNode(node)

	var brenchEndBlocks []*mir.Block
	var endBlock *mir.Block
	if node.HasElse() {
		brenchEndBlocks = blocks
		end := iterator.All(dynarray.NewDynArrayWith(brenchEndBlocks...), func(v *mir.Block) bool {
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

func (self *CodeGenerator) codegenIfElseNode(node *hir.IfElse) []*mir.Block {
	if condNode, ok := node.Cond.Value(); ok {
		cond := self.codegenExpr(condNode, true)
		trueStartBlock, trueEndBlock := self.codegenBlock(node.Body, nil)
		falseBlock := trueStartBlock.Belong().NewBlock()
		self.builder.BuildCondJump(cond, trueStartBlock, falseBlock)
		self.builder.MoveTo(falseBlock)

		if nextNode, ok := node.Next.Value(); ok {
			blocks := self.codegenIfElseNode(nextNode)
			return append([]*mir.Block{trueEndBlock}, blocks...)
		} else {
			return []*mir.Block{trueEndBlock, falseBlock}
		}
	} else {
		self.codegenFlatBlock(node.Body)
		return []*mir.Block{self.builder.Current()}
	}
}

type loop interface {
	SetOutBlock(block *mir.Block)
	GetOutBlock() (*mir.Block, bool)
	GetNextBlock() *mir.Block
}

type endlessLoop struct {
	BodyEntry *mir.Block
	Out       *mir.Block
}

func (self *endlessLoop) SetOutBlock(block *mir.Block) {
	self.Out = block
}

func (self *endlessLoop) GetOutBlock() (*mir.Block, bool) {
	return self.Out, self.Out != nil
}

func (self *endlessLoop) GetNextBlock() *mir.Block {
	return self.BodyEntry
}

func (self *CodeGenerator) codegenEndlessLoop(node *hir.EndlessLoop) {
	entryBlock, endBlock := self.codegenBlock(node.Body, func(block *mir.Block) {
		self.loops.Set(node, &endlessLoop{BodyEntry: block})
	})
	self.builder.BuildUnCondJump(entryBlock)
	if !endBlock.Terminated() {
		self.builder.MoveTo(endBlock)
		self.builder.BuildUnCondJump(entryBlock)
	}
	if outBlock, ok := self.loops.Get(node).GetOutBlock(); ok {
		self.builder.MoveTo(outBlock)
	}
}

func (self *CodeGenerator) codegenBreak(node *hir.Break) {
	loop := self.loops.Get(node.Loop)
	if _, ok := loop.GetOutBlock(); !ok {
		loop.SetOutBlock(self.builder.Current().Belong().NewBlock())
	}
	endBlock, _ := loop.GetOutBlock()
	self.builder.BuildUnCondJump(endBlock)
}

func (self *CodeGenerator) codegenContinue(node *hir.Continue) {
	loop := self.loops.Get(node.Loop)
	self.builder.BuildUnCondJump(loop.GetNextBlock())
}

type forRange struct {
	Action *mir.Block
	Out    *mir.Block
}

func (self *forRange) SetOutBlock(block *mir.Block) {}

func (self *forRange) GetOutBlock() (*mir.Block, bool) {
	return self.Out, true
}

func (self *forRange) GetNextBlock() *mir.Block {
	return self.Action
}

func (self *CodeGenerator) codegenFor(node *hir.For) {
	// pre
	iter := self.codegenExpr(node.Iterator, false)
	iterArrayTypeMean := node.Iterator.GetType().(*hir.ArrayType)
	indexPtr := self.builder.BuildAllocFromStack(self.ctx.Usize())
	self.builder.BuildStore(mir.NewInt(indexPtr.ElemType().(mir.IntType), 0), indexPtr)
	cursorPtr := self.codegenLocalVariable(node.Cursor)
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
		self.builder.BuildCondJump(
			self.builder.BuildCmp(mir.CmpKindLT, self.builder.BuildLoad(indexPtr), mir.NewInt(indexPtr.ElemType().(mir.IntType), int64(iterArrayTypeMean.Size))),
			entry,
			outBlock,
		)

		// action
		self.builder.MoveTo(actionBlock)
		self.builder.BuildStore(self.builder.BuildAdd(self.builder.BuildLoad(indexPtr), mir.NewInt(indexPtr.ElemType().(mir.IntType), 1)), indexPtr)
		self.builder.BuildUnCondJump(condBlock)

		// body
		self.builder.MoveTo(entry)
		self.builder.BuildStore(self.buildArrayIndex(iter, self.builder.BuildLoad(indexPtr), false), cursorPtr)

		self.loops.Set(node, &forRange{
			Action: actionBlock,
			Out:    outBlock,
		})
	}

	_, endBlock := self.codegenBlock(node.Body, beforeBody)
	if !endBlock.Terminated() {
		self.builder.MoveTo(endBlock)
		self.builder.BuildUnCondJump(actionBlock)
	}
	if outBlock, ok := self.loops.Get(node).GetOutBlock(); ok {
		self.builder.MoveTo(outBlock)
	}
}
