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

type BinaryType uint8

const (
	BinaryInvalid BinaryType = iota
	BinaryAdd
	BinarySub
	BinaryMul
	BinaryDiv
	BinaryRem
)

// Binary 二元运算
type Binary struct {
	Kind        BinaryType
	Left, Right Expr
}

func (self *Binary) stmt() {}

func (self *Binary) GetType() Type {
	return self.Left.GetType()
}

type UnaryType uint8

const (
	UnaryInvalid UnaryType = iota
	UnaryNegate
)

// Binary 二元运算
type Unary struct {
	Kind  UnaryType
	Value Expr
}

func (self *Unary) stmt() {}

func (self *Unary) GetType() Type {
	return self.Value.GetType()
}
