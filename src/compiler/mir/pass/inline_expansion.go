package pass

import (
	"github.com/kkkunny/containers/list"

	"github.com/kkkunny/Sim/src/compiler/mir"
)

// 内联展开
type inlineExpansion struct{}

func newInlineExpansion() *inlineExpansion {
	return &inlineExpansion{}
}
func (self *inlineExpansion) walk(pkg *mir.Package) {
	for cursor := pkg.Globals.Front(); cursor != nil; cursor = cursor.Next() {
		self.walkGlobal(cursor.Value())
	}
}
func (self *inlineExpansion) walkGlobal(ir mir.Global) {
	switch global := ir.(type) {
	case *mir.Function:
		for cursor := global.Blocks.Front(); cursor != nil; cursor = cursor.Next() {
			vars := make(map[mir.Value]mir.Value)
			self.walkBlock(global, cursor, &vars)
		}
	}
}
func (self *inlineExpansion) walkBlock(
	fn *mir.Function, blockNode *list.ListNode[*mir.Block], vars *map[mir.Value]mir.Value,
) {
	for cursor := blockNode.Value().Insts.Front(); cursor != nil; {
		next := cursor.Next()
		if _, ok := cursor.Value().(*mir.Call); ok || len(*vars) != 0 {
			switch inst := cursor.Value().(type) {
			case *mir.Load:
				inst.Value = self.walkValue(inst.Value, *vars)
			case *mir.Store:
				inst.From = self.walkValue(inst.From, *vars)
				inst.To = self.walkValue(inst.To, *vars)
			case *mir.Return:
				if inst.Value != nil {
					inst.Value = self.walkValue(inst.Value, *vars)
				}
			case *mir.CondJmp:
				inst.Cond = self.walkValue(inst.Cond, *vars)
			case *mir.Phi:
				for i, v := range inst.Values {
					inst.Values[i] = self.walkValue(v, *vars)
				}
			case *mir.Eq:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Ne:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Lt:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Le:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Gt:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Ge:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.ArrayIndex:
				inst.From = self.walkValue(inst.From, *vars)
				inst.Index = self.walkValue(inst.Index, *vars)
			case *mir.StructIndex:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.PointerIndex:
				inst.From = self.walkValue(inst.From, *vars)
				inst.Index = self.walkValue(inst.Index, *vars)
			case *mir.UnwrapUnion:
				inst.Value = self.walkValue(inst.Value, *vars)
			case *mir.Add:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Sub:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Mul:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Div:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Mod:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.And:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Or:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Xor:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Shl:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Shr:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Sint2Sint:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Sint2Uint:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Sint2Float:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Sint2Ptr:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Uint2Uint:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Uint2Sint:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Uint2Float:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Uint2Ptr:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Float2Sint:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Float2Uint:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Float2Float:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Ptr2Ptr:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Ptr2Sint:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Ptr2Uint:
				inst.From = self.walkValue(inst.From, *vars)
			case *mir.Select:
				inst.Cond = self.walkValue(inst.Cond, *vars)
				inst.True = self.walkValue(inst.True, *vars)
				inst.False = self.walkValue(inst.False, *vars)
			case *mir.Not:
				inst.Value = self.walkValue(inst.Value, *vars)
			case *mir.LogicNot:
				inst.Value = self.walkValue(inst.Value, *vars)
			case *mir.LogicAnd:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.LogicOr:
				inst.Left = self.walkValue(inst.Left, *vars)
				inst.Right = self.walkValue(inst.Right, *vars)
			case *mir.Call:
				inst.Func = self.walkValue(inst.Func, *vars)
				for i, a := range inst.Args {
					inst.Args[i] = self.walkValue(a, *vars)
				}

				f, ok := inst.Func.(*mir.Function)
				if !ok || !f.GetInline() {
					break
				}
				if ret, endBlock := self.expandInlineFunction(fn, blockNode, cursor, inst); ret != nil {
					(*vars)[inst] = endBlock.NewLoad(ret)
					endBlock.Insts.MoveToFront(endBlock.Insts.Back())
				}
				// 删除内联函数调用
				blockNode.Value().Insts.RemoveNode(cursor)
			case *mir.WrapUnion:
				inst.Value = self.walkValue(inst.Value, *vars)
			}
		}
		cursor = next
	}
}
func (self *inlineExpansion) walkValue(ir mir.Value, vars map[mir.Value]mir.Value) mir.Value {
	switch value := ir.(type) {
	case mir.Constant, *mir.Function:
		return value
	default:
		if v, ok := vars[value]; ok {
			return v
		}
		return value
	}
}
func (self *inlineExpansion) expandInlineFunction(
	toFn *mir.Function, toBlockNode *list.ListNode[*mir.Block], callNode *list.ListNode[mir.Inst], call *mir.Call,
) (*mir.Alloc, *mir.Block) {
	f := call.Func.(*mir.Function)
	// 参数
	args := make(map[mir.Value]mir.Value, len(call.Args))
	for i, a := range call.Args {
		alloc := toBlockNode.Value().NewAlloc(a.GetType())
		toBlockNode.Value().Insts.MoveToFrontOfNode(toBlockNode.Value().Insts.Back(), callNode)
		toBlockNode.Value().NewStore(a, alloc)
		toBlockNode.Value().Insts.MoveToFrontOfNode(toBlockNode.Value().Insts.Back(), callNode)
		args[f.Params[i]] = alloc
	}
	// 返回值
	var retAlloc *mir.Alloc
	if retType := f.Type.GetFuncRet(); !retType.IsVoid() {
		retAlloc = toBlockNode.Value().NewAlloc(retType)
		toBlockNode.Value().Insts.MoveToFrontOfNode(toBlockNode.Value().Insts.Back(), callNode)
	}
	// 复制内联函数的inst到当前位置
	endBlock := self.copyFunction(f, toFn, toBlockNode, callNode, retAlloc, &args)
	return retAlloc, endBlock
}
func (self *inlineExpansion) copyFunction(
	ir *mir.Function, toFn *mir.Function, toBlockNode *list.ListNode[*mir.Block], callNode *list.ListNode[mir.Inst],
	ret *mir.Alloc,
	vars *map[mir.Value]mir.Value,
) *mir.Block {
	// 将内联函数的代码块插入原函数中
	blocks := make(map[*mir.Block]*mir.Block)
	proBlockNode := toBlockNode
	for cursor := ir.Blocks.Front(); cursor != nil; cursor = cursor.Next() {
		blocks[cursor.Value()] = toFn.NewBlock()
		newBlockNode := toFn.Blocks.Back()
		toFn.Blocks.MoveToFrontOfNode(newBlockNode, proBlockNode)
		proBlockNode = newBlockNode
	}
	// 原函数的代码块插入一个跳转语句，跳转到内联函数的第一个代码块
	toBlockNode.Value().NewJmp(blocks[ir.Blocks.Front().Value()])
	toBlockNode.Value().Insts.MoveToFrontOfNode(toBlockNode.Value().Insts.Back(), callNode)
	// 将调用内联函数后的所有语句放进一个新的代码块，该代码块位于内联函数的代码块之后
	endBlock := toFn.NewBlock()
	toFn.Blocks.MoveToFrontOfNode(toFn.Blocks.Back(), proBlockNode)
	for cursor := callNode.Next(); cursor != nil; cursor = cursor.Next() {
		endBlock.Insts.PushBack(cursor.Value())
	}
	// 将调用内联函数之后的代码块舍弃
	for cursor := callNode.Next(); cursor != nil; {
		next := cursor.Next()
		toBlockNode.Value().Insts.RemoveNode(cursor)
		cursor = next
	}
	// 遍历内联函数的语句，替换值和代码块
	for blockCursor := ir.Blocks.Front(); blockCursor != nil; blockCursor = blockCursor.Next() {
		for instCursor := blockCursor.Value().Insts.Front(); instCursor != nil; instCursor = instCursor.Next() {
			self.copyInst(instCursor.Value(), blocks[blockCursor.Value()], endBlock, blocks, ret, vars)
		}
	}
	return endBlock
}
func (self *inlineExpansion) copyInst(
	ir mir.Inst, toBlock *mir.Block, endBlock *mir.Block, blocks map[*mir.Block]*mir.Block, ret *mir.Alloc,
	vars *map[mir.Value]mir.Value,
) {
	var value mir.Value

	switch inst := ir.(type) {
	case *mir.Alloc:
		value = toBlock.NewAlloc(inst.Type)
	case *mir.Load:
		value = toBlock.NewLoad(self.copyValue(inst.Value, *vars))
	case *mir.Store:
		toBlock.NewStore(self.copyValue(inst.From, *vars), self.copyValue(inst.To, *vars))
	case *mir.Unreachable:
		toBlock.NewUnreachable()
	case *mir.Return:
		if inst.Value != nil {
			toBlock.NewStore(self.copyValue(inst.Value, *vars), ret)
			toBlock.NewJmp(endBlock)
		}
	case *mir.CondJmp:
		toBlock.NewCondJmp(self.copyValue(inst.Cond, *vars), blocks[inst.True], blocks[inst.False])
	case *mir.Jmp:
		toBlock.NewJmp(blocks[inst.Dst])
	case *mir.Phi:
		bs, vs := make([]*mir.Block, len(inst.Froms)), make([]mir.Value, len(inst.Values))
		for i, b := range inst.Froms {
			bs[i] = blocks[b]
			vs[i] = self.copyValue(inst.Values[i], *vars)
		}
		value = toBlock.NewPhi(bs, vs)
	case *mir.Eq:
		value = toBlock.NewEq(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Ne:
		value = toBlock.NewNe(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Lt:
		value = toBlock.NewLt(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Le:
		value = toBlock.NewLe(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Gt:
		value = toBlock.NewGt(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Ge:
		value = toBlock.NewGe(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.ArrayIndex:
		value = toBlock.NewArrayIndex(self.copyValue(inst.From, *vars), self.copyValue(inst.Index, *vars))
	case *mir.StructIndex:
		value = toBlock.NewStructIndex(self.copyValue(inst.From, *vars), inst.Index)
	case *mir.PointerIndex:
		value = toBlock.NewPointerIndex(self.copyValue(inst.From, *vars), self.copyValue(inst.Index, *vars))
	case *mir.UnwrapUnion:
		value = toBlock.NewUnwrapUnion(inst.To, self.copyValue(inst.Value, *vars))
	case *mir.Add:
		value = toBlock.NewAdd(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Sub:
		value = toBlock.NewSub(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Mul:
		value = toBlock.NewMul(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Div:
		value = toBlock.NewDiv(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Mod:
		value = toBlock.NewMod(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.And:
		value = toBlock.NewAnd(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Or:
		value = toBlock.NewOr(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Xor:
		value = toBlock.NewXor(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Shl:
		value = toBlock.NewShl(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Shr:
		value = toBlock.NewShr(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Sint2Sint:
		value = toBlock.NewSint2Sint(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Sint2Uint:
		value = toBlock.NewSint2Uint(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Sint2Float:
		value = toBlock.NewSint2Float(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Sint2Ptr:
		value = toBlock.NewSint2Ptr(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Uint2Uint:
		value = toBlock.NewUint2Uint(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Uint2Sint:
		value = toBlock.NewUint2Sint(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Uint2Float:
		value = toBlock.NewUint2Float(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Uint2Ptr:
		value = toBlock.NewUint2Ptr(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Float2Sint:
		value = toBlock.NewFloat2Sint(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Float2Uint:
		value = toBlock.NewFloat2Uint(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Float2Float:
		value = toBlock.NewFloat2Float(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Ptr2Ptr:
		value = toBlock.NewPtr2Ptr(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Ptr2Sint:
		value = toBlock.NewPtr2Sint(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Ptr2Uint:
		value = toBlock.NewPtr2Uint(self.copyValue(inst.From, *vars), inst.To)
	case *mir.Select:
		value = toBlock.NewSelect(
			self.copyValue(inst.Cond, *vars),
			self.copyValue(inst.True, *vars),
			self.copyValue(inst.False, *vars),
		)
	case *mir.Not:
		value = toBlock.NewNot(self.copyValue(inst.Value, *vars))
	case *mir.LogicNot:
		value = toBlock.NewLogicNot(self.copyValue(inst.Value, *vars))
	case *mir.LogicAnd:
		value = toBlock.NewLogicAnd(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.LogicOr:
		value = toBlock.NewLogicOr(self.copyValue(inst.Left, *vars), self.copyValue(inst.Right, *vars))
	case *mir.Call:
		args := make([]mir.Value, len(inst.Args))
		for i, a := range inst.Args {
			args[i] = self.copyValue(a, *vars)
		}
		value = toBlock.NewCall(self.copyValue(inst.Func, *vars), args...)
	case *mir.WrapUnion:
		value = toBlock.NewWrapUnion(inst.To, self.copyValue(inst.Value, *vars))
	default:
		panic("unreachable")
	}

	if value != nil {
		(*vars)[ir.(mir.Value)] = value
	}
}
func (self *inlineExpansion) copyValue(ir mir.Value, vars map[mir.Value]mir.Value) mir.Value {
	switch value := ir.(type) {
	case mir.Constant, *mir.Function:
		return value
	default:
		if v, ok := vars[value]; ok {
			return v
		}
		return value
	}
}
