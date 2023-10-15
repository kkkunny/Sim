package mean

import "math/big"

// Expr 表达式
type Expr interface {
	Stmt
	GetType() Type
}

// Ident 标识符
type Ident interface {
	Expr
	ident()
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

// Float 浮点数
type Float struct {
	Type  *FloatType
	Value big.Float
}

func (self *Float) stmt() {}

func (self *Float) GetType() Type {
	return self.Type
}
