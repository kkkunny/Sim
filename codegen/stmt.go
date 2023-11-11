package codegen

import (
	"github.com/kkkunny/go-llvm"
	"github.com/kkkunny/stl/container/dynarray"
	"github.com/kkkunny/stl/container/iterator"

	"github.com/kkkunny/Sim/mean"
	"github.com/kkkunny/Sim/util"
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
	case *mean.Loop:
		self.codegenLoop(stmtNode)
	case *mean.Break:
		self.codegenBreak(stmtNode)
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
	if v, ok := node.Value.Value(); ok {
		v := self.codegenExpr(v, true)
		self.builder.CreateRet(&v)
	} else {
		self.builder.CreateRet(nil)
	}
}

func (self *CodeGenerator) codegenLocalVariable(node *mean.Variable) {
	t := self.codegenType(node.Type)
	value := self.codegenExpr(node.Value, true)
	ptr := self.builder.CreateAlloca("", t)
	self.builder.CreateStore(value, ptr)
	self.values[node] = ptr
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

type Loop struct {
	Entry llvm.Block
	End   util.Option[llvm.Block]
}

func (self *CodeGenerator) codegenLoop(node *mean.Loop) {
	nextBlock, nextEndBlock := self.codegenBlock(node.Body, func(block llvm.Block) {
		self.loops.Set(node, &Loop{Entry: block})
	})
	self.builder.CreateBr(nextBlock)
	if !nextEndBlock.IsTerminating() {
		self.builder.MoveToAfter(nextEndBlock)
		self.builder.CreateBr(nextBlock)
	}
	if endBlock, ok := self.loops.Get(node).End.Value(); ok {
		self.builder.MoveToAfter(endBlock)
	}
}

func (self *CodeGenerator) codegenBreak(node *mean.Break) {
	loop := self.loops.Get(node.Loop)
	if loop.End.IsNone() {
		loop.End = util.Some(self.builder.CurrentBlock().Belong().NewBlock(""))
	}
	endBlock, _ := loop.End.Value()
	self.builder.CreateBr(endBlock)
}
