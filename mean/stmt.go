package mean

// Stmt 语句
type Stmt interface {
	stmt()
}

// Block 代码块
type Block struct{}

func (self *Block) stmt() {}
