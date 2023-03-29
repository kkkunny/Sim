package mirgen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/Sim/src/compiler/mir"
)

// 代码块
func (self *MirGenerator) genBlock(mean hir.Block) {
	for iter := mean.Stmts.Iterator(); iter.HasValue(); iter.Next() {
		self.genStmt(iter.Value())
	}
}

// 语句
func (self *MirGenerator) genStmt(mean hir.Stmt) {
	switch meanStmt := mean.(type) {
	case *hir.Return:
		self.genReturn(*meanStmt)
	case *hir.Variable:
		self.genVariable(meanStmt)
	case hir.Expr:
		self.genExpr(meanStmt, true)
	case *hir.Block:
		self.genBlock(*meanStmt)
	case *hir.IfElse:
		self.genIfElse(*meanStmt)
	case *hir.Loop:
		self.genLoop(*meanStmt)
	case *hir.Break:
		self.block.NewJmp(self.eb)
	case *hir.Continue:
		self.block.NewJmp(self.cb)
	case *hir.Switch:
		self.genSwitch(*meanStmt)
	case *hir.Match:
		self.genMatch(*meanStmt)
	default:
		panic("")
	}
}

// 函数返回
func (self *MirGenerator) genReturn(mean hir.Return) {
	if mean.Value == nil {
		self.block.NewReturn(nil)
	} else {
		value := self.genExpr(mean.Value, true)
		self.block.NewReturn(value)
	}
}

// 变量
func (self *MirGenerator) genVariable(mean *hir.Variable) {
	alloca := self.block.NewAlloc(self.genType(mean.Typ))
	value := self.genExpr(mean.Value, true)
	self.vars[mean] = alloca
	self.block.NewStore(value, alloca)
}

// 条件分支
func (self *MirGenerator) genIfElse(mean hir.IfElse) {
	_cond := self.genExpr(mean.Cond, true)
	cond := self.block.NewEq(_cond, mir.NewInt(_cond.GetType(), 1))

	tb := self.block.Belong.NewBlock()
	if mean.False == nil {
		eb := self.block.Belong.NewBlock()
		self.block.NewCondJmp(cond, tb, eb)

		self.block = tb
		self.genBlock(*mean.True)
		self.block.NewJmp(eb)

		self.block = eb
	} else {
		fb, eb := self.block.Belong.NewBlock(), self.block.Belong.NewBlock()
		self.block.NewCondJmp(cond, tb, fb)

		self.block = tb
		self.genBlock(*mean.True)
		self.block.NewJmp(eb)

		self.block = fb
		self.genBlock(*mean.False)
		self.block.NewJmp(eb)

		self.block = eb
	}
}

// 循环
func (self *MirGenerator) genLoop(mean hir.Loop) {
	cb := self.block.Belong.NewBlock()
	self.block.NewJmp(cb)

	self.block = cb
	_cond := self.genExpr(mean.Cond, true)
	cond := self.block.NewEq(_cond, mir.NewInt(_cond.GetType(), 1))
	lb, eb := self.block.Belong.NewBlock(), self.block.Belong.NewBlock()
	self.block.NewCondJmp(cond, lb, eb)

	cbBk, ebBk := self.cb, self.eb
	self.cb, self.eb = cb, eb
	self.block = lb
	self.genBlock(*mean.Body)
	self.block.NewJmp(cb)

	self.cb, self.eb = cbBk, ebBk

	self.block = eb
}

// 分支
func (self *MirGenerator) genSwitch(mean hir.Switch) {
	if len(mean.CaseValues) == 0 && mean.Default == nil {
		return
	} else if len(mean.CaseValues) == 0 {
		self.genBlock(*mean.Default)
	} else {
		from := self.genExpr(mean.From, true)

		eb := self.block.Belong.NewBlock()
		var db *mir.Block
		if mean.Default == nil {
			db = eb
		} else {
			db = self.block.Belong.NewBlock()
		}
		for i, c := range mean.CaseValues {
			cv := self.genExpr(c, true)
			_nb, cc := self.block.NewEqual(from, cv)
			self.block = _nb
			ct := self.block.Belong.NewBlock()
			var cf *mir.Block
			if i == len(mean.CaseValues)-1 {
				cf = db
			} else {
				cf = self.block.Belong.NewBlock()
			}
			self.block.NewCondJmp(cc, ct, cf)

			self.block = ct
			self.genBlock(*mean.CaseBodies[i])
			self.block.NewJmp(eb)

			self.block = cf
		}
		if mean.Default != nil {
			self.block = db
			self.genBlock(*mean.Default)
			self.block.NewJmp(eb)
		}

		self.block = eb
	}
}

// 枚举匹配
func (self *MirGenerator) genMatch(mean hir.Match) {
	if len(mean.CaseValues) == 0 && mean.Default == nil {
		return
	} else if len(mean.CaseValues) == 0 {
		self.genBlock(*mean.Default)
	} else {
		FieldHirs := mean.From.Type().GetEnumFields()
		from := self.genExpr(mean.From, true)
		index := self.block.NewStructIndex(from, 0)
		var indexValue mir.Value
		if index.IsPtr() {
			indexValue = self.block.NewLoad(index)
		} else {
			indexValue = index
		}

		eb := self.block.Belong.NewBlock()
		var db *mir.Block
		if mean.Default == nil {
			db = eb
		} else {
			db = self.block.Belong.NewBlock()
		}
		for i, c := range mean.CaseValues {
			var fieldIndex int
			for i, f := range FieldHirs {
				if f.Second == c {
					fieldIndex = i
					break
				}
			}
			cc := self.block.NewEq(indexValue, mir.NewUint(t_usize, uint64(fieldIndex)))
			ct := self.block.Belong.NewBlock()
			var cf *mir.Block
			if i == len(mean.CaseValues)-1 {
				cf = db
			} else {
				cf = self.block.Belong.NewBlock()
			}
			self.block.NewCondJmp(cc, ct, cf)

			self.block = ct
			self.genBlock(*mean.CaseBodies[i])
			self.block.NewJmp(eb)

			self.block = cf
		}
		if mean.Default != nil {
			self.block = db
			self.genBlock(*mean.Default)
			self.block.NewJmp(eb)
		}

		self.block = eb
	}
}
