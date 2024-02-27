package module

import (
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/queue"

	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/pass/function"
)

// DeadCodeElimination 死码消除
var DeadCodeElimination = new(_DeadCodeElimination)

type _DeadCodeElimination struct {
	valueRefCountTable hashmap.HashMap[mir.GlobalValue, uint64]
}

func (self *_DeadCodeElimination) init(_ mir.Module) {
	self.valueRefCountTable = hashmap.NewHashMap[mir.GlobalValue, uint64]()
}

func (self *_DeadCodeElimination) Run(ir *mir.Module) {
	// FIXME: 缺少可达性分析，导致两个互相引用的变量不会被删除掉
	// 计算每个变量的引用次数
	for cursor := ir.Globals().Front(); cursor != nil; cursor = cursor.Next() {
		self.walkGlobal(cursor.Value)
	}
	// 删除未被引用的变量
	for cursor := ir.Globals().Front(); cursor != nil; {
		next := cursor.Next()
		if gv, ok := cursor.Value.(mir.GlobalValue); ok && self.valueRefCountTable.Get(gv) == 0 {
			ir.Globals().Remove(cursor)
		}
		cursor = next
	}
}

func (self *_DeadCodeElimination) walkGlobal(ir mir.Global) {
	switch global := ir.(type) {
	case *mir.Function:
		function.Run(global, function.DeadCodeElimination)
		for blockCursor := global.Blocks().Front(); blockCursor != nil; blockCursor = blockCursor.Next() {
			for stmtCursor := blockCursor.Value.Stmts().Front(); stmtCursor != nil; stmtCursor = stmtCursor.Next() {
				self.countGlobalValueRefs(append(stmtCursor.Value.ReadRefValues(), stmtCursor.Value.WriteRefValues()...))
			}
		}
		// 入口函数额外+1
		if global.IsStartFunction() {
			self.valueRefCountTable.Set(global, self.valueRefCountTable.Get(global)+1)
		}
	}
	self.countGlobalValueRefs(append(ir.ReadRefValues(), ir.WriteRefValues()...))
}

func (self *_DeadCodeElimination) countGlobalValueRefs(values []mir.Value) {
	varQueue := queue.NewQueueWith(values...)
	for !varQueue.Empty() {
		if ref, ok := varQueue.Peek().(mir.GlobalValue); ok {
			self.valueRefCountTable.Set(ref, self.valueRefCountTable.Get(ref)+1)
		}
		for _, v := range varQueue.Pop().ReadRefValues() {
			varQueue.Push(v)
		}
	}
}
