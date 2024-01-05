package function

import (
	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/hashmap"
	"github.com/kkkunny/stl/container/hashset"
	"github.com/kkkunny/stl/container/pair"
	"github.com/samber/lo"

	"github.com/kkkunny/Sim/mir"
)

// DeadVariablesElimination 无关变量消除
var DeadVariablesElimination = new(_DeadVariablesElimination)

type _DeadVariablesElimination struct {
	readMap  hashmap.HashMap[mir.Stmt, *hashset.HashSet[mir.Stmt]]
	writeMap  hashmap.HashMap[mir.Stmt, *hashset.HashSet[mir.Stmt]]
	validMap hashset.HashSet[mir.Stmt]
}

func (self *_DeadVariablesElimination) init(ir mir.Function){
	var count uint
	for cursor :=ir.Blocks().Front(); cursor !=nil; cursor = cursor.Next(){
		count += uint(cursor.Value.Stmts().Len())
	}
	self.readMap = hashmap.NewHashMapWithCapacity[mir.Stmt, *hashset.HashSet[mir.Stmt]](count)
	self.writeMap = hashmap.NewHashMapWithCapacity[mir.Stmt, *hashset.HashSet[mir.Stmt]](count)
	self.validMap = hashset.NewHashSet[mir.Stmt]()
}

func (self *_DeadVariablesElimination) Run(ir *mir.Function){
	// 初始化每个语句的读写列表
	for blockCursor :=ir.Blocks().Front(); blockCursor !=nil; blockCursor = blockCursor.Next(){
		for stmtCursor:=blockCursor.Value.Stmts().Front(); stmtCursor!=nil; stmtCursor=stmtCursor.Next(){
			reads := hashset.NewHashSet[mir.Stmt]()
			self.readMap.Set(stmtCursor.Value, &reads)
			writes := hashset.NewHashSet[mir.Stmt]()
			self.writeMap.Set(stmtCursor.Value, &writes)
		}
	}
	// 构建读写关系
	for blockCursor :=ir.Blocks().Front(); blockCursor !=nil; blockCursor = blockCursor.Next(){
		for stmtCursor:=blockCursor.Value.Stmts().Front(); stmtCursor!=nil; stmtCursor=stmtCursor.Next(){
			self.walkStmt(stmtCursor.Value)
		}
	}
	// 检测有效指令
	for {
		var needLoop bool
		for blockCursor :=ir.Blocks().Front(); blockCursor !=nil; blockCursor = blockCursor.Next(){
			for stmtCursor:=blockCursor.Value.Stmts().Front(); stmtCursor!=nil; stmtCursor=stmtCursor.Next(){
				if self.validMap.Contain(stmtCursor.Value){
					// 有效语句的读写依赖全部有效
					for iter:=self.readMap.Get(stmtCursor.Value).Iterator(); iter.Next(); {
						needLoop = needLoop || self.validMap.Add(iter.Value())
					}
					for iter:=self.writeMap.Get(stmtCursor.Value).Iterator(); iter.Next(); {
						needLoop = needLoop || self.validMap.Add(iter.Value())
					}
				}else{
					// 若写的目标是有效语句则该语句有效
					for iter:=self.writeMap.Get(stmtCursor.Value).Iterator(); iter.Next(); {
						if self.validMap.Contain(iter.Value()){
							needLoop = needLoop || self.validMap.Add(stmtCursor.Value)
						}
					}
				}
			}
		}
		if !needLoop{
			break
		}
	}
	// 删除无效指令
	for blockCursor :=ir.Blocks().Front(); blockCursor !=nil; blockCursor = blockCursor.Next(){
		for stmtCursor:=blockCursor.Value.Stmts().Front(); stmtCursor!=nil;{
			stmtNextCursor := stmtCursor.Next()
			if !self.validMap.Contain(stmtCursor.Value){
				blockCursor.Value.Stmts().Remove(stmtCursor)
			}
			stmtCursor = stmtNextCursor
		}
	}
}

func (self *_DeadVariablesElimination) walkStmt(ir mir.Stmt){
	var readValues, writeValues []mir.Value

	switch stmt := ir.(type) {
	case *mir.Store:
		readValues = []mir.Value{stmt.From()}
		writeValues = []mir.Value{stmt.To()}
	case *mir.Return:
		if v, ok := stmt.Value(); ok{
			readValues = []mir.Value{v}
		}
		self.validMap.Add(stmt)
	case *mir.UnCondJump:
		self.validMap.Add(stmt)
	case *mir.CondJump:
		readValues = []mir.Value{stmt.Cond()}
		self.validMap.Add(stmt)
	case *mir.Load:
		readValues = []mir.Value{stmt.From()}
	case *mir.And:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Or:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Xor:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Shl:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Shr:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Not:
		readValues = []mir.Value{stmt.Value()}
	case *mir.Add:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Sub:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Mul:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Div:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Rem:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Cmp:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.PtrEqual:
		readValues = []mir.Value{stmt.Left(), stmt.Right()}
	case *mir.Call:
		readValues = append(stmt.Args(), stmt.Func())
		self.validMap.Add(stmt)
	case *mir.NumberCovert:
		readValues = []mir.Value{stmt.From()}
	case *mir.PtrToUint:
		readValues = []mir.Value{stmt.From()}
	case *mir.UintToPtr:
		readValues = []mir.Value{stmt.From()}
		writeValues = []mir.Value{stmt.From()}
	case *mir.PtrToPtr:
		readValues = []mir.Value{stmt.From()}
		writeValues = []mir.Value{stmt.From()}
	case *mir.ArrayIndex:
		readValues = []mir.Value{stmt.Array(), stmt.Index()}
		writeValues = []mir.Value{stmt.Array()}
	case *mir.StructIndex:
		readValues = []mir.Value{stmt.Struct()}
		writeValues = []mir.Value{stmt.Struct()}
	case *mir.Phi:
		readValues = lo.Map(stmt.Froms(), func(item pair.Pair[*mir.Block, mir.Value], _ int) mir.Value {
			return item.Second
		})
	case *mir.PackArray:
		readValues = stmt.Elems()
	case *mir.PackStruct:
		readValues = stmt.Elems()
	case *mir.Unreachable:
		self.validMap.Add(stmt)
	case *mir.Switch:
		readValues = []mir.Value{stmt.Value()}
		self.validMap.Add(stmt)
	}

	reads := self.readMap.Get(ir)
	for _, v := range readValues {
		if s, ok := v.(mir.StmtValue); ok{
			reads.Add(s)
		}
	}
	writes := self.writeMap.Get(ir)
	for _, v := range writeValues {
		if s, ok := v.(mir.StmtValue); ok{
			writes.Add(s)
		}
	}

	// 如果写的目标是全局变量，则该语句有效
	for _, v := range writeValues {
		if stlbasic.Is[mir.Global](v){
			self.validMap.Add(ir)
			break
		}
	}
}
