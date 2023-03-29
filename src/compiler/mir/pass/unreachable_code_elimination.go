package pass

import (
	list "github.com/bahlo/generic-list-go"
	"github.com/kkkunny/Sim/src/compiler/mir"
	"github.com/kkkunny/stl/set"
)

// 不可达代码消除
type unreachableCodeElimination struct{}

func newUnreachableCodeElimination() *unreachableCodeElimination {
	return &unreachableCodeElimination{}
}
func (self *unreachableCodeElimination) walk(pkg *mir.Package) {
	for cursor := pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		self.walkGlobal(cursor.Value)
	}
}
func (self *unreachableCodeElimination) walkGlobal(ir mir.Global) {
	switch global := ir.(type) {
	case *mir.Function:
		if global.Blocks.Len() == 0 {
			return
		}
		blocks := set.NewHashSet[*mir.Block]()
		self.walkBlock(blocks, global.Blocks.Front().Value)
		for cursor := global.Blocks.Front(); cursor != nil; {
			next := cursor.Next()
			if !blocks.Contain(cursor.Value) {
				global.Blocks.Remove(cursor)
			}
			cursor = next
		}
	}
}
func (self *unreachableCodeElimination) walkBlock(walkedBlocks *set.HashSet[*mir.Block], ir *mir.Block) {
	if !walkedBlocks.Add(ir) {
		return
	}
	handle := func(end *list.Element[mir.Inst]) {
		for endNext := end.Next(); endNext != nil; {
			next := endNext.Next()
			ir.Insts.Remove(endNext)
			endNext = next
		}
	}
	for cursor := ir.Insts.Front(); cursor != nil; cursor = cursor.Next() {
		switch inst := cursor.Value.(type) {
		case *mir.Unreachable, *mir.Return:
			handle(cursor)
			return
		case *mir.Jmp:
			handle(cursor)
			self.walkBlock(walkedBlocks, inst.Dst)
			return
		case *mir.CondJmp:
			handle(cursor)
			self.walkBlock(walkedBlocks, inst.True)
			self.walkBlock(walkedBlocks, inst.False)
			return
		case *mir.Call:
			f, ok := inst.Func.(*mir.Function)
			if !ok || !f.NoReturn {
				continue
			}
			if _, ok := cursor.Next().Value.(*mir.Unreachable); ok {
				handle(cursor.Next())
				return
			}
			handle(cursor)
			ir.NewUnreachable()
			return
		}
	}
}
