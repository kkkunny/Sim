package mir

import (
	"fmt"
	"strings"

	"github.com/kkkunny/stl/container/linkedlist"
)

// Block 代码块
type Block struct {
	index uint
	stmts linkedlist.LinkedList[Stmt]
}

func (self *Function) NewBlock()*Block{
	b := &Block{}
	self.blocks.PushBack(b)
	return b
}

func (self *Block) Name()string{
	return fmt.Sprintf("b%d", self.index)
}

func (self *Block) String()string{
	var buf strings.Builder

	buf.WriteString(self.Name())
	buf.WriteString(":\n")
	var i uint
	for iter:=self.stmts.Iterator(); iter.Next(); {
		buf.WriteString("  ")
		iter.Value().setIndex(i)
		buf.WriteString(iter.Value().Define())
		if iter.HasNext(){
			buf.WriteByte('\n')
		}
		i++
	}

	return buf.String()
}
