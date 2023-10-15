package mean

import "math/big"

// Expr 表达式
type Expr interface {
	Stmt
	GetType() Type
}

// Integer 整数
type Integer struct {
	Type  IntType
	Value big.Int
}

func (self *Integer) stmt() {}

func (self *Integer) GetType() Type {
	return self.Type
}
