package mir

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/container/linkedlist"
)

// Block 代码块
type Block struct {
	f *Function
	index uint
	stmts linkedlist.LinkedList[Stmt]
}

func (self *Function) NewBlock()*Block{
	b := &Block{f: self}
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
		stmtValue, ok := iter.Value().(StmtValue)
		if ok{
			stmtValue.setIndex(i)
			i++
		}
	}
	for iter:=self.stmts.Iterator(); iter.Next(); {
		buf.WriteString("  ")
		buf.WriteString(iter.Value().Define())
		if iter.HasNext(){
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

func (self *Block) Belong()*Function{
	return self.f
}

func (self *Block) Terminated()bool{
	for iter:=self.stmts.Iterator(); iter.Next(); {
		if stlbasic.Is[Terminating](iter.Value()){
			return true
		}
	}
	return false
}

func (self *Block) Stmts()linkedlist.LinkedList[Stmt]{
	return self.stmts
}
