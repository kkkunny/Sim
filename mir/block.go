package mir

import (
	"fmt"
	"strings"

	stlbasic "github.com/kkkunny/stl/basic"
	"github.com/kkkunny/stl/list"
)

// Block 代码块
type Block struct {
	f *Function
	index uint
	stmts *list.List[Stmt]
}

func (self *Function) NewBlock()*Block{
	b := &Block{
		f: self,
		stmts: list.New[Stmt](),
	}
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
	for cursor:=self.stmts.Front(); cursor!=nil; cursor=cursor.Next(){
		stmtValue, ok := cursor.Value.(StmtValue)
		if ok{
			stmtValue.setIndex(i)
			i++
		}
	}
	for cursor:=self.stmts.Front(); cursor!=nil; cursor=cursor.Next(){
		buf.WriteString("  ")
		buf.WriteString(cursor.Value.Define())
		if cursor.Next() != nil{
			buf.WriteByte('\n')
		}
	}

	return buf.String()
}

func (self *Block) Belong()*Function{
	return self.f
}

func (self *Block) Terminated()bool{
	for cursor:=self.stmts.Front(); cursor!=nil; cursor=cursor.Next(){
		if stlbasic.Is[Terminating](cursor.Value){
			return true
		}
	}
	return false
}

func (self *Block) Stmts()*list.List[Stmt]{
	return self.stmts
}

func (self *Block) IsEntry()bool{
	return self.f.Blocks().Front().Value == self
}
