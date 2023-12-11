package pass

import (
	stlbasic "github.com/kkkunny/stl/basic"

	"github.com/kkkunny/Sim/mir"
)

// DeadCodeElimination 死代码消除
var DeadCodeElimination = new(_DeadCodeElimination)

type _DeadCodeElimination struct {}

func (self *_DeadCodeElimination) Run(module *mir.Module){
	for cursor:=module.Globals().Front(); cursor!=nil; cursor=cursor.Next(){
		self.walkGlobal(cursor.Value)
	}
}

func (self *_DeadCodeElimination) walkGlobal(ir mir.Global){
	switch global := ir.(type) {
	case *mir.Function:
		for cursor:=global.Blocks().Front(); cursor!=nil; cursor=cursor.Next(){
			self.walkBlock(cursor.Value)
		}
	}
}

func (self *_DeadCodeElimination) walkBlock(ir *mir.Block){
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

func (self *_DeadCodeElimination) walkStmt(ir mir.Stmt)bool{
	return stlbasic.Is[mir.Terminating](ir)
}
