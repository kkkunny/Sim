package mir

import (
	"strings"

	"github.com/kkkunny/stl/container/linkedlist"
)

// Module 模块
type Module struct {
	ctx *Context

	globals linkedlist.LinkedList[Global]
	structMap map[string]*namedStruct
	valueMap map[string]Value
}

func (self *Context) NewModule()*Module{
	return &Module{
		ctx: self,

		structMap: make(map[string]*namedStruct),
		valueMap: make(map[string]Value),
	}
}

func (self Module) String()string{
	var buf strings.Builder

	var ts linkedlist.LinkedList[*namedStruct]
	var cs linkedlist.LinkedList[*Constant]
	var vs linkedlist.LinkedList[*GlobalVariable]
	var fs linkedlist.LinkedList[*Function]
	for iter:=self.globals.Iterator(); iter.Next(); {
		switch g := iter.Value().(type) {
		case *namedStruct:
			ts.PushBack(g)
		case *GlobalVariable:
			vs.PushBack(g)
		case *Function:
			fs.PushBack(g)
		default:
			panic("unreachable")
		}
	}

	var i uint
	for iter:=ts.Iterator(); iter.Next(); {
		iter.Value().setIndex(i)
		buf.WriteString(iter.Value().Define())
		buf.WriteByte('\n')
		i++
	}

	if !ts.Empty(){
		buf.WriteByte('\n')
	}
	i = 0
	for iter:=cs.Iterator(); iter.Next(); {
		iter.Value().setIndex(i)
		buf.WriteString(iter.Value().Define())
		buf.WriteByte('\n')
		i++
	}

	if !cs.Empty(){
		buf.WriteByte('\n')
	}
	i = 0
for iter:=vs.Iterator(); iter.Next(); {
		iter.Value().setIndex(i)
		buf.WriteString(iter.Value().Define())
		buf.WriteByte('\n')
		i++
	}

	if !vs.Empty(){
		buf.WriteByte('\n')
	}
	i = 0
	for iter:=fs.Iterator(); iter.Next(); {
		iter.Value().setIndex(i)
		buf.WriteString(iter.Value().Define())
		buf.WriteByte('\n')
		i++
	}

	return buf.String()
}
