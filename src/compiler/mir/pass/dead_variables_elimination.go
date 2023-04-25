package pass

import (
	"github.com/kkkunny/containers/list"

	"github.com/kkkunny/Sim/src/compiler/mir"
	"github.com/kkkunny/stl/set"
)

// 死变量消除
type deadVariablesElimination struct {
	refedCount map[mir.Global]uint
	walkeds    *set.HashSet[mir.Global]
}

func newDeadVariablesElimination() *deadVariablesElimination {
	return &deadVariablesElimination{
		refedCount: make(map[mir.Global]uint),
		walkeds:    set.NewHashSet[mir.Global](),
	}
}
func (self *deadVariablesElimination) walk(pkg *mir.Package) {
	// 找出根节点
	roots := self.findRoot(pkg)
	// 从根节点出发计算所有变量/函数被引用的次数
	for cursor := roots.Front(); cursor != nil; cursor = cursor.Next() {
		self.walkGlobal(cursor.Value())
	}
	// 删除零引用的节点
	self.removeZeroRef(pkg)
}
func (self *deadVariablesElimination) findRoot(pkg *mir.Package) *list.List[mir.Global] {
	roots := list.NewList[mir.Global]()
	for cursor := pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		switch global := cursor.Value().(type) {
		case *mir.Function:
			if (global.Extern && global.Blocks.Length() != 0) || global.GetInit() || global.GetFini() {
				roots.PushBack(global)
			}
		case *mir.Variable:
			if global.Extern && global.Value != nil {
				roots.PushBack(global)
			}
		}
	}
	return roots
}
func (self *deadVariablesElimination) walkGlobal(ir mir.Global) {
	if !self.walkeds.Add(ir) {
		return
	}
	self.refedCount[ir]++

	switch global := ir.(type) {
	case *mir.Alias:
		self.walkType(global.Target)
	case *mir.Function:
		for _, p := range global.Params {
			self.walkType(p.Type)
		}
		for cursor := global.Blocks.Front(); cursor != nil; cursor = cursor.Next() {
			self.walkBlock(cursor.Value())
		}
	case *mir.Variable:
		self.walkType(global.Type)
		if global.Value == nil {
			return
		}
		self.walkValue(global.Value)
	}
}
func (self *deadVariablesElimination) walkBlock(ir *mir.Block) {
	for cursor := ir.Insts.Front(); cursor != nil; cursor = cursor.Next() {
		self.walkInst(cursor.Value())
	}
}
func (self *deadVariablesElimination) walkInst(ir mir.Inst) {
	switch inst := ir.(type) {
	case *mir.Alloc:
		self.walkType(inst.Type)
	case *mir.Load:
		self.walkValue(inst.Value)
	case *mir.Store:
		self.walkValue(inst.From)
		self.walkValue(inst.To)
	case *mir.Return:
		if inst.Value == nil {
			return
		}
		self.walkValue(inst.Value)
	case *mir.CondJmp:
		self.walkValue(inst.Cond)
	case *mir.Eq:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Ne:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Lt:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Le:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Gt:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Ge:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.ArrayIndex:
		self.walkValue(inst.From)
		self.walkValue(inst.Index)
	case *mir.StructIndex:
		self.walkValue(inst.From)
	case *mir.PointerIndex:
		self.walkValue(inst.From)
		self.walkValue(inst.Index)
	case *mir.UnwrapUnion:
		self.walkValue(inst.Value)
		self.walkType(inst.To)
	case *mir.Add:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Sub:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Mul:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Div:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Mod:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.And:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Or:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Xor:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Shl:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Shr:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Phi:
		for _, v := range inst.Values {
			self.walkValue(v)
		}
	case *mir.Sint2Sint:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Sint2Uint:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Sint2Float:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Sint2Ptr:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Uint2Uint:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Uint2Sint:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Uint2Float:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Uint2Ptr:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Float2Sint:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Float2Uint:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Float2Float:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Ptr2Ptr:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Ptr2Sint:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Ptr2Uint:
		self.walkValue(inst.From)
		self.walkType(inst.To)
	case *mir.Select:
		self.walkValue(inst.Cond)
		self.walkValue(inst.True)
		self.walkValue(inst.False)
	case *mir.Not:
		self.walkValue(inst.Value)
	case *mir.LogicNot:
		self.walkValue(inst.Value)
	case *mir.LogicAnd:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.LogicOr:
		self.walkValue(inst.Left)
		self.walkValue(inst.Right)
	case *mir.Call:
		self.walkValue(inst.Func)
		for _, a := range inst.Args {
			self.walkValue(a)
		}
	case *mir.WrapUnion:
		self.walkValue(inst.Value)
		self.walkType(inst.To)
	}
}
func (self *deadVariablesElimination) walkValue(ir mir.Value) {
	switch value := ir.(type) {
	case mir.Constant:
		self.walkConstant(value)
	case *mir.Function:
		self.walkGlobal(value)
	}
}
func (self *deadVariablesElimination) walkConstant(ir mir.Constant) {
	switch value := ir.(type) {
	case *mir.Array:
		for _, e := range value.Elems {
			self.walkConstant(e)
		}
	case *mir.Struct:
		for _, e := range value.Elems {
			self.walkConstant(e)
		}
	case *mir.ArrayIndexConst:
		self.walkConstant(value.From)
	case *mir.Variable:
		self.walkGlobal(value)
	}
}
func (self *deadVariablesElimination) walkType(ir mir.Type) {
	switch ir.Kind {
	case mir.TAlias:
		self.walkGlobal(ir.GetAlias())
	case mir.TArray:
		self.walkType(ir.GetArrayElem())
	case mir.TFunc:
		self.walkType(ir.GetFuncRet())
		for _, p := range ir.GetFuncParams() {
			self.walkType(p)
		}
	case mir.TUnion:
		for _, e := range ir.GetUnionElems() {
			self.walkType(e)
		}
	case mir.TPtr:
		self.walkType(ir.GetPtr())
	case mir.TStruct:
		for _, e := range ir.GetStructElems() {
			self.walkType(e)
		}
	}
}
func (self *deadVariablesElimination) removeZeroRef(pkg *mir.Package) {
	for cursor := pkg.Globals.Front(); cursor != nil; {
		next := cursor.Next()
		if self.refedCount[cursor.Value()] == 0 {
			pkg.Globals.RemoveNode(cursor)
		}
		cursor = next
	}
}
