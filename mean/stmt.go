package mean

import "github.com/kkkunny/stl/container/linkedlist"

// Stmt 语句
type Stmt interface {
	stmt()
}

// Block 代码块
type Block struct {
	Stmts linkedlist.LinkedList[Stmt]
}

func (self *Block) stmt() {}

type Return struct{}

func (self *Return) stmt() {}
