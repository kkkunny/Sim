package ast

// Type 类型
type Type interface {
	Ast
	typ()
}
