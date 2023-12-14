package function

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/pair"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mir"
)

// CodeBlockMerge 代码块合并
var CodeBlockMerge = new(_CodeBlockMerge)

type _CodeBlockMerge struct {
	blockFroms hashmap.HashMap[*mir.Block, *hashset.HashSet[*mir.Block]]
	mustBeSeparatedList []hashset.HashSet[*mir.Block]
}

func (self *_CodeBlockMerge) init(ir mir.Function){
	self.blockFroms = hashmap.NewHashMapWithCapacity[*mir.Block, *hashset.HashSet[*mir.Block]](uint(ir.Blocks().Len()))
	self.mustBeSeparatedList = make([]hashset.HashSet[*mir.Block], 0)
}

func (self *_CodeBlockMerge) Run(ir *mir.Function){
	// 初始化每个块的来源列表
	for cursor:=ir.Blocks().Front(); cursor!=nil; cursor=cursor.Next(){
		froms := hashset.NewHashSet[*mir.Block]()
		self.blockFroms.Set(cursor.Value, &froms)
	}
	// 消除每个块内的死代码，构建来源图
	for cursor:=ir.Blocks().Front(); cursor!=nil; cursor=cursor.Next(){
		self.walkBlock(cursor.Value)
	}
	// 合并只有一个来源的块
	for cursor:=ir.Blocks().Front(); cursor!=nil; cursor=cursor.Next(){
		self.mergeBlock(cursor.Value)
	}
	// 删除空块
	for cursor:=ir.Blocks().Front(); cursor!=nil;{
		nextCursor := cursor.Next()
		if cursor.Value.Stmts().Len() == 0{
			ir.Blocks().Remove(cursor)
		}
		cursor = nextCursor
	}
}

func (self *_CodeBlockMerge) walkBlock(ir *mir.Block){
	// 找到结束语句的下一个语句
	cursor := ir.Stmts().Front()
	for ; cursor!=nil; cursor=cursor.Next(){
		if self.walkStmt(cursor.Value){
			cursor = cursor.Next()
			break
		}
	}

	// 删除包括下一个的所有语句
	for cursor != nil{
		next := cursor.Next()
		ir.Stmts().Remove(cursor)
		cursor = next
	}
}

func (self *_CodeBlockMerge) walkStmt(ir mir.Stmt)bool{
	switch stmt := ir.(type) {
	case mir.Jump:
		for _, to := range stmt.Targets(){
			self.blockFroms.Get(to).Add(ir.Belong())
		}
	case *mir.Phi:
		froms := lo.Map(stmt.Froms(), func(item pair.Pair[*mir.Block, mir.Value], _ int) *mir.Block {
			return item.First
		})
		self.mustBeSeparatedList = append(self.mustBeSeparatedList, hashset.NewHashSetWith(froms...))
	}
	return stlbasic.Is[mir.Terminating](ir)
}

func (self *_CodeBlockMerge) getNextBlock(block *mir.Block)*mir.Block{
	froms := self.blockFroms.Get(block)
	if froms.Length() != 1{
		return block
	}
	iter := froms.Iterator()
	iter.Next()
	nextBlock := iter.Value()
	if nextBlock.Stmts().Len() == 0{
		return self.getNextBlock(nextBlock)
	}
	if !stlbasic.Is[mir.UnCondJump](nextBlock.Stmts().Back().Value){
		return block
	}
	return nextBlock
}

func (self *_CodeBlockMerge) mergeBlock(block *mir.Block){
	nextBlock := self.getNextBlock(block)
	if nextBlock == block{
		return
	}
	for _, list := range self.mustBeSeparatedList{
		if list.Contain(block) && list.Contain(nextBlock){
			return
		}
	}
	nextBlock.Stmts().Remove(nextBlock.Stmts().Back())
	if stlbasic.Is[*mir.Phi](block.Stmts().Front().Value){
		block.Stmts().Remove(block.Stmts().Front())
	}
	nextBlock.Stmts().PushBackList(block.Stmts())
	block.Stmts().Init()
}
