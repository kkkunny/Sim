package function

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/list"

	"github.com/kkkunny/Sim/mir"
)

// UnreachableCodeElimination 不可达代码消除
var UnreachableCodeElimination = new(_UnreachableCodeElimination)

type _UnreachableCodeElimination struct {
	blockFroms hashmap.HashMap[*mir.Block, *hashset.HashSet[*mir.Block]]
}

func (self *_UnreachableCodeElimination) init(ir mir.Function){
	self.blockFroms = hashmap.NewHashMapWithCapacity[*mir.Block, *hashset.HashSet[*mir.Block]](uint(ir.Blocks().Len()))
}

func (self *_UnreachableCodeElimination) Run(ir *mir.Function){
	// 初始化每个块的来源列表
	for cursor:=ir.Blocks().Front(); cursor!=nil; cursor=cursor.Next(){
		froms := hashset.NewHashSet[*mir.Block]()
		self.blockFroms.Set(cursor.Value, &froms)
	}
	// 消除每个块内的死代码，构建来源图
	for cursor:=ir.Blocks().Front(); cursor!=nil; cursor=cursor.Next(){
		self.walkBlock(cursor.Value)
	}
	// 删除没有来源不是entry的块
	for {
		var needLoop bool
		for cursor:=ir.Blocks().Front(); cursor!=nil;{
			next := cursor.Next()
			if !cursor.Value.IsEntry() && self.blockFroms.Get(cursor.Value).Empty(){
				self.removeBlock(ir, cursor)
				needLoop = true
			}
			cursor = next
		}
		if !needLoop{
			break
		}
	}
}

func (self *_UnreachableCodeElimination) walkBlock(ir *mir.Block){
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

func (self *_UnreachableCodeElimination) walkStmt(ir mir.Stmt)bool{
	if jump, ok := ir.(mir.Jump); ok{
		for _, to := range jump.Targets(){
			self.blockFroms.Get(to).Add(ir.Belong())
		}
	}
	return stlbasic.Is[mir.Terminating](ir)
}

func (self *_UnreachableCodeElimination) removeBlock(f *mir.Function, cursor *list.Element[*mir.Block]){
	block := cursor.Value
	f.Blocks().Remove(cursor)
	self.blockFroms.Remove(block)
	for iter:=self.blockFroms.Iterator(); iter.Next(); {
		iter.Value().Second.Remove(block)
	}
}
