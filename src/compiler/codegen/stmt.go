package codegen

import (
	"github.com/kkkunny/Sim/src/compiler/analyse"
	"github.com/kkkunny/go-llvm"
)

// 代码块
func (self *CodeGenerator) codegenBlock(mean analyse.Block) bool {
	for _, stmt := range mean.Stmts {
		if !self.codegenStmt(stmt) {
			return false
		}
	}
	return true
}

// 语句
func (self *CodeGenerator) codegenStmt(mean analyse.Stmt) bool {
	switch meanStmt := mean.(type) {
	case *analyse.Return:
		self.codegenReturn(*meanStmt)
		return false
	case *analyse.Variable:
		self.codegenVariable(meanStmt)
	case analyse.Expr:
		self.codegenExpr(meanStmt, true)
	case *analyse.Block:
		if !self.codegenBlock(*meanStmt) {
			return false
		}
	case *analyse.IfElse:
		self.codegenIfElse(*meanStmt)
	case *analyse.Loop:
		self.codegenLoop(*meanStmt)
	case *analyse.LoopControl:
		self.codegenLoopControl(*meanStmt)
	default:
		panic("")
	}
	return true
}

// 函数返回
func (self *CodeGenerator) codegenReturn(mean analyse.Return) {
	if mean.Value == nil {
		self.builder.CreateRetVoid()
	} else {
		value := self.codegenExpr(mean.Value, true)
		self.builder.CreateRet(value)
	}
}

// 变量
func (self *CodeGenerator) codegenVariable(mean *analyse.Variable) {
	alloca := self.builder.CreateAlloca(self.codegenType(mean.Type), "")
	value := self.codegenExpr(mean.Value, true)
	self.vars[mean] = alloca
	self.builder.CreateStore(value, alloca)
}

// 条件分支
func (self *CodeGenerator) codegenIfElse(mean analyse.IfElse) {
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
func (self *CodeGenerator) codegenLoop(mean analyse.Loop) {
	cb := llvm.AddBasicBlock(self.function, "")
	self.builder.CreateBr(cb)

	self.builder.SetInsertPointAtEnd(cb)
	lb, eb := llvm.AddBasicBlock(self.function, ""), llvm.AddBasicBlock(self.function, "")
	self.builder.CreateCondBr(self.builder.CreateIntCast(self.codegenExpr(mean.Cond, true), self.ctx.Int1Type(), ""), lb, eb)

	cbBk, ebBk := self.cb, self.eb
	self.cb, self.eb = cb, eb
	self.builder.SetInsertPointAtEnd(lb)
	if self.codegenBlock(*mean.Body) {
		self.builder.CreateBr(cb)
	}

	self.cb, self.eb = cbBk, ebBk

	self.builder.SetInsertPointAtEnd(eb)
}

// 循环控制
func (self *CodeGenerator) codegenLoopControl(mean analyse.LoopControl) {
	if mean.Type == "break" {
		self.builder.CreateBr(self.eb)
	} else {
		self.builder.CreateBr(self.cb)
	}
}
