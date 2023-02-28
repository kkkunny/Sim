package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/hir"
	"github.com/kkkunny/llvm"
	stlutil "github.com/kkkunny/stl/util"
)

// 代码块
func (self *CodeGenerator) codegenBlock(mean hir.Block) bool {
	for iter := mean.Stmts.Iterator(); iter.HasValue(); iter.Next() {
		if !self.codegenStmt(iter.Value()) {
			return false
		}
	}
	return true
}

// 语句
func (self *CodeGenerator) codegenStmt(mean hir.Stmt) bool {
	switch meanStmt := mean.(type) {
	case *hir.Return:
		self.codegenReturn(*meanStmt)
		return false
	case *hir.Variable:
		self.codegenVariable(meanStmt)
	case hir.Expr:
		self.codegenExpr(meanStmt, true)
	case *hir.Block:
		if !self.codegenBlock(*meanStmt) {
			return false
		}
	case *hir.IfElse:
		self.codegenIfElse(*meanStmt)
	case *hir.Loop:
		self.codegenLoop(*meanStmt)
	case *hir.Break:
		self.builder.CreateBr(self.eb)
		return false
	case *hir.Continue:
		self.builder.CreateBr(self.cb)
		return false
	case *hir.Switch:
		self.codegenSwitch(*meanStmt)
	case *hir.Match:
		self.codegenMatch(*meanStmt)
	default:
		panic("")
	}
	return true
}

// 函数返回
func (self *CodeGenerator) codegenReturn(mean hir.Return) {
	if mean.Value == nil {
		self.builder.CreateRetVoid()
	} else {
		value := self.codegenExpr(mean.Value, true)
		self.builder.CreateRet(value)
	}
}

// 变量
func (self *CodeGenerator) codegenVariable(mean *hir.Variable) {
	alloca := self.createAllocaHirType(mean.Type())
	value := self.codegenExpr(mean.Value, true)
	self.vars[mean] = alloca
	store := self.builder.CreateStore(value, alloca)
	store.SetAlignment(int(mean.Type().Align()))
}

// 条件分支
func (self *CodeGenerator) codegenIfElse(mean hir.IfElse) {
	cond := self.builder.CreateIntCast(self.codegenExpr(mean.Cond, true), self.ctx.Int1Type(), "")
	tb := llvm.AddBasicBlock(self.function, "")
	if mean.False == nil {
		eb := llvm.AddBasicBlock(self.function, "")
		self.builder.CreateCondBr(cond, tb, eb)

		self.builder.SetInsertPointAtEnd(tb)
		if self.codegenBlock(*mean.True) {
			self.builder.CreateBr(eb)
		}

		self.builder.SetInsertPointAtEnd(eb)
	} else {
		fb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
		self.builder.CreateCondBr(cond, tb, fb)

		self.builder.SetInsertPointAtEnd(tb)
		if self.codegenBlock(*mean.True) {
			self.builder.CreateBr(eb)
		}

		self.builder.SetInsertPointAtEnd(fb)
		if self.codegenBlock(*mean.False) {
			self.builder.CreateBr(eb)
		}

		self.builder.SetInsertPointAtEnd(eb)
	}
}

// 循环
func (self *CodeGenerator) codegenLoop(mean hir.Loop) {
	cb := llvm.AddBasicBlock(self.function, "")
	self.builder.CreateBr(cb)

	self.builder.SetInsertPointAtEnd(cb)
	lb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
	self.builder.CreateCondBr(
		self.builder.CreateIntCast(self.codegenExpr(mean.Cond, true), self.ctx.Int1Type(), ""),
		lb,
		eb,
	)

	cbBk, ebBk := self.cb, self.eb
	self.cb, self.eb = cb, eb
	self.builder.SetInsertPointAtEnd(lb)
	if self.codegenBlock(*mean.Body) {
		self.builder.CreateBr(cb)
	}

	self.cb, self.eb = cbBk, ebBk

	self.builder.SetInsertPointAtEnd(eb)
}

// 分支
func (self *CodeGenerator) codegenSwitch(mean hir.Switch) {
	if len(mean.CaseValues) == 0 && mean.Default == nil {
		return
	} else if len(mean.CaseValues) == 0 {
		self.codegenBlock(*mean.Default)
	} else {
		from := self.codegenExpr(mean.From, true)

		eb := llvm.AddBasicBlock(self.function, "")
		db := stlutil.Ternary(mean.Default == nil, eb, llvm.AddBasicBlock(self.function, ""))
		for i, c := range mean.CaseValues {
			cv := self.codegenExpr(c, true)
			cc := self.equal(from, cv)
			ct := llvm.AddBasicBlock(self.function, "")
			cf := stlutil.Ternary(i == len(mean.CaseValues)-1, db, llvm.AddBasicBlock(self.function, ""))
			self.builder.CreateCondBr(cc, ct, cf)

			self.builder.SetInsertPointAtEnd(ct)
			if self.codegenBlock(*mean.CaseBodies[i]) {
				self.builder.CreateBr(eb)
			}

			self.builder.SetInsertPointAtEnd(cf)
		}
		if mean.Default != nil {
			self.builder.SetInsertPointAtEnd(db)
			if self.codegenBlock(*mean.Default) {
				self.builder.CreateBr(eb)
			}
		}

		self.builder.SetInsertPointAtEnd(eb)
	}
}

// 枚举匹配
func (self *CodeGenerator) codegenMatch(mean hir.Match) {
	if len(mean.CaseValues) == 0 && mean.Default == nil {
		return
	} else if len(mean.CaseValues) == 0 {
		self.codegenBlock(*mean.Default)
	} else {
		FieldHirs := mean.From.Type().GetEnumFields()
		index := self.createStructIndex(self.codegenExpr(mean.From, true), 0, true, 1)

		eb := llvm.AddBasicBlock(self.function, "")
		db := stlutil.Ternary(mean.Default == nil, eb, llvm.AddBasicBlock(self.function, ""))
		for i, c := range mean.CaseValues {
			var fieldIndex int
			for i, f := range FieldHirs {
				if f.Second == c {
					fieldIndex = i
					break
				}
			}
			cc := self.equal(index, llvm.ConstInt(t_size, uint64(fieldIndex), false))
			ct := llvm.AddBasicBlock(self.function, "")
			cf := stlutil.Ternary(i == len(mean.CaseValues)-1, db, llvm.AddBasicBlock(self.function, ""))
			self.builder.CreateCondBr(cc, ct, cf)

			self.builder.SetInsertPointAtEnd(ct)
			if self.codegenBlock(*mean.CaseBodies[i]) {
				self.builder.CreateBr(eb)
			}

			self.builder.SetInsertPointAtEnd(cf)
		}
		if mean.Default != nil {
			self.builder.SetInsertPointAtEnd(db)
			if self.codegenBlock(*mean.Default) {
				self.builder.CreateBr(eb)
			}
		}

		self.builder.SetInsertPointAtEnd(eb)
	}
}
