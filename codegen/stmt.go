package codegen

import (
	"github.com/kkkunny/go-llvm"

	"github.com/kkkunny/Sim/mean"
)

func (self *CodeGenerator) codegenStmt(node mean.Stmt) {
	switch stmtNode := node.(type) {
	case *mean.Return:
		self.codegenReturn(stmtNode)
	case *mean.Variable:
		self.codegenVariable(stmtNode)
	case *mean.IfElse:
		self.codegenIfElse(stmtNode)
	case mean.Expr:
		self.codegenExpr(stmtNode)
	default:
		panic("unreachable")
	}
}

func (self *CodeGenerator) codegenFlatBlock(node *mean.Block) {
	for iter := node.Stmts.Iterator(); iter.Next(); {
		self.codegenStmt(iter.Value())
	}
}

func (self *CodeGenerator) codegenBlock(node *mean.Block) (llvm.Block, llvm.Block) {
	from := self.builder.CurrentBlock()
	defer self.builder.MoveToAfter(from)

	block := from.Belong().NewBlock("")
	self.builder.MoveToAfter(block)

	self.codegenFlatBlock(node)

	return block, self.builder.CurrentBlock()
}

func (self *CodeGenerator) codegenReturn(node *mean.Return) {
	if v, ok := node.Value.Value(); ok {
		v := self.codegenExpr(v)
		self.builder.CreateRet(&v)
	} else {
		self.builder.CreateRet(nil)
	}
}

func (self *CodeGenerator) codegenVariable(node *mean.Variable) {
	t := self.codegenType(node.Type)
	value := self.codegenExpr(node.Value)
	ptr := self.builder.CreateAlloca("", t)
	self.builder.CreateStore(value, ptr)
	self.values[node] = ptr
}

func (self *CodeGenerator) codegenIfElse(node *mean.IfElse) {
	blocks := self.codegenIfElseNode(node)

	var brenchEndBlocks []llvm.Block
	var endBlock llvm.Block
	if node.HasElse() {
		brenchEndBlocks, endBlock = blocks, blocks[0].Belong().NewBlock("")
	} else {
		brenchEndBlocks, endBlock = blocks[:len(blocks)-1], blocks[len(blocks)-1]
	}

	for _, brenchEndBlock := range brenchEndBlocks {
		self.builder.MoveToAfter(brenchEndBlock)
		self.builder.CreateBr(endBlock)
	}
	self.builder.MoveToAfter(endBlock)
}

func (self *CodeGenerator) codegenIfElseNode(node *mean.IfElse) []llvm.Block {
	if condNode, ok := node.Cond.Value(); ok {
		cond := self.codegenExpr(condNode)
		trueStartBlock, trueEndBlock := self.codegenBlock(node.Body)
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
