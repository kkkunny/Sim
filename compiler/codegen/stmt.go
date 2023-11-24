package codegen

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/iterator"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenStmt(node mean.Stmt) {
	switch stmtNode := node.(type) {
	case *mean.Return:
		self.codegenReturn(stmtNode)
	case *mean.Variable:
		self.codegenLocalVariable(stmtNode)
	case *mean.IfElse:
		self.codegenIfElse(stmtNode)
	case mean.Expr:
		self.codegenExpr(stmtNode, false)
	case *mean.EndlessLoop:
		self.codegenEndlessLoop(stmtNode)
	case *mean.Break:
		self.codegenBreak(stmtNode)
	case *mean.Continue:
		self.codegenContinue(stmtNode)
	case *mean.For:
		self.codegenFor(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFlatBlock(node *mean.Block) {
	for iter := node.Stmts.Iterator(); iter.Next(); {
		if self.builder.CurrentBlock().IsTerminating() {
			break
		}
		self.codegenStmt(iter.Value())
	}
}

func (self *CodeGenerator) codegenBlock(node *mean.Block, afterBlockCreate func(block llvm.Block)) (llvm.Block, llvm.Block) {
	from := self.builder.CurrentBlock()
	block := from.Belong().NewBlock("")
	if afterBlockCreate != nil {
		afterBlockCreate(block)
	}

	self.builder.MoveToAfter(block)
	defer self.builder.MoveToAfter(from)

	self.codegenFlatBlock(node)

	return block, self.builder.CurrentBlock()
}

func (self *CodeGenerator) codegenReturn(node *mean.Return) {
	ft := self.codegenFuncType(node.Func.GetType().(*mean.FuncType))
	if v, ok := node.Value.Value(); ok {
		v := self.codegenExpr(v, true)
		self.buildRet(ft, &v)
	} else {
		self.buildRet(ft, nil)
	}
}

func (self *CodeGenerator) codegenLocalVariable(node *mean.Variable) llvm.Value {
	t := self.codegenType(node.Type)
	value := self.codegenExpr(node.Value, true)
	ptr := self.builder.CreateAlloca("", t)
	self.builder.CreateStore(value, ptr)
	self.values[node] = ptr
	return ptr
}

func (self *CodeGenerator) codegenIfElse(node *mean.IfElse) {
	blocks := self.codegenIfElseNode(node)

	var brenchEndBlocks []llvm.Block
	var endBlock llvm.Block
	if node.HasElse() {
		brenchEndBlocks = blocks
		end := iterator.All(dynarray.NewDynArrayWith(brenchEndBlocks...), func(v llvm.Block) bool {
			return v.IsTerminating()
		})
		if end {
			return
		}
		endBlock = blocks[0].Belong().NewBlock("")
	} else {
		brenchEndBlocks, endBlock = blocks[:len(blocks)-1], blocks[len(blocks)-1]
	}

	for _, brenchEndBlock := range brenchEndBlocks {
		if brenchEndBlock.IsTerminating() {
			continue
		}
		self.builder.MoveToAfter(brenchEndBlock)
		self.builder.CreateBr(endBlock)
	}
	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) codegenIfElseNode(node *mean.IfElse) []llvm.Block {
	if condNode, ok := node.Cond.Value(); ok {
		cond := self.codegenExpr(condNode, true)
		trueStartBlock, trueEndBlock := self.codegenBlock(node.Body, nil)
		falseBlock := trueStartBlock.Belong().NewBlock("")
		self.builder.CreateCondBr(cond, trueStartBlock, falseBlock)
		self.builder.MoveToAfter(falseBlock)

		if nextNode, ok := node.Next.Value(); ok {
			blocks := self.codegenIfElseNode(nextNode)
			return append([]llvm.Block{trueEndBlock}, blocks...)
		} else {
			return []llvm.Block{trueEndBlock, falseBlock}
		}
	} else {
		self.codegenFlatBlock(node.Body)
		return []llvm.Block{self.builder.CurrentBlock()}
	}
}

type loop interface {
	SetOutBlock(block llvm.Block)
	GetOutBlock() (*llvm.Block, bool)
	GetNextBlock() llvm.Block
}

type endlessLoop struct {
	BodyEntry llvm.Block
	Out       *llvm.Block
}

func (self *endlessLoop) SetOutBlock(block llvm.Block) {
	self.Out = &block
}

func (self *endlessLoop) GetOutBlock() (*llvm.Block, bool) {
	return self.Out, self.Out != nil
}

func (self *endlessLoop) GetNextBlock() llvm.Block {
	return self.BodyEntry
}

func (self *CodeGenerator) codegenEndlessLoop(node *mean.EndlessLoop) {
	entryBlock, endBlock := self.codegenBlock(node.Body, func(block llvm.Block) {
		self.loops.Set(node, &endlessLoop{BodyEntry: block})
	})
	self.builder.CreateBr(entryBlock)
	if !endBlock.IsTerminating() {
		self.builder.MoveToAfter(endBlock)
		self.builder.CreateBr(entryBlock)
	}
	if outBlock, ok := self.loops.Get(node).GetOutBlock(); ok {
		self.builder.MoveToAfter(*outBlock)
	}
}

func (self *CodeGenerator) codegenBreak(node *mean.Break) {
	loop := self.loops.Get(node.Loop)
	if _, ok := loop.GetOutBlock(); !ok {
		loop.SetOutBlock(self.builder.CurrentBlock().Belong().NewBlock(""))
	}
	endBlock, _ := loop.GetOutBlock()
	self.builder.CreateBr(*endBlock)
}

func (self *CodeGenerator) codegenContinue(node *mean.Continue) {
	loop := self.loops.Get(node.Loop)
	self.builder.CreateBr(loop.GetNextBlock())
}

type forRange struct {
	Action llvm.Block
	Out    llvm.Block
}

func (self *forRange) SetOutBlock(block llvm.Block) {}

func (self *forRange) GetOutBlock() (*llvm.Block, bool) {
	return &self.Out, true
}

func (self *forRange) GetNextBlock() llvm.Block {
	return self.Action
}

func (self *CodeGenerator) codegenFor(node *mean.For) {
	// pre
	iter := self.codegenExpr(node.Iterator, false)
	iterArrayTypeMean := node.Iterator.GetType().(*mean.ArrayType)
	indexPtr := self.builder.CreateAlloca("", self.ctx.IntPtrType(self.target))
	self.builder.CreateStore(self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 0), indexPtr)
	cursorPtr := self.codegenLocalVariable(node.Cursor)
	condBlock := self.builder.CurrentBlock().Belong().NewBlock("")
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
		self.builder.CreateCondBr(
			self.builder.CreateIntCmp("", llvm.IntULT, self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr), self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), int64(iterArrayTypeMean.Size))),
			entry,
			outBlock,
		)

		// action
		self.builder.MoveToAfter(actionBlock)
		self.builder.CreateStore(self.builder.CreateUAdd("", self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr), self.ctx.ConstInteger(self.ctx.IntPtrType(self.target), 1)), indexPtr)
		self.builder.CreateBr(condBlock)

		// body
		self.builder.MoveToAfter(entry)
		self.builder.CreateStore(self.buildArrayIndex(self.codegenArrayType(iterArrayTypeMean), iter, self.builder.CreateLoad("", self.ctx.IntPtrType(self.target), indexPtr), true), cursorPtr)

		self.loops.Set(node, &forRange{
			Action: actionBlock,
			Out:    outBlock,
		})
	}

	_, endBlock := self.codegenBlock(node.Body, beforeBody)
	if !endBlock.IsTerminating() {
		self.builder.MoveToAfter(endBlock)
		self.builder.CreateBr(actionBlock)
	}
	if outBlock, ok := self.loops.Get(node).GetOutBlock(); ok {
		self.builder.MoveToAfter(*outBlock)
	}
}
