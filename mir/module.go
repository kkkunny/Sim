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

	for iter:=ts.Iterator(); iter.Next(); {
		buf.WriteString(iter.Value().Define())
		buf.WriteByte('\n')
	}

	if !ts.Empty(){
		buf.WriteByte('\n')
	}
	for iter:=cs.Iterator(); iter.Next(); {
		buf.WriteString(iter.Value().Define())
		buf.WriteByte('\n')
	}

	if !cs.Empty(){
		buf.WriteByte('\n')
	}
	for iter:=vs.Iterator(); iter.Next(); {
		buf.WriteString(iter.Value().Define())
		buf.WriteByte('\n')
	}

	if !vs.Empty(){
		buf.WriteByte('\n')
	}
	for iter:=fs.Iterator(); iter.Next(); {
		buf.WriteString(iter.Value().Define())
		buf.WriteByte('\n')
	}

	return buf.String()
}
