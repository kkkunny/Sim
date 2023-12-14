package module

import (
	"github.com/kkkunny/Sim/mir"
	"github.com/kkkunny/Sim/mir/pass/function"
)

// DeadCodeElimination 死码消除
var DeadCodeElimination = new(_DeadCodeElimination)

type _DeadCodeElimination struct {}

func (self *_DeadCodeElimination) init(_ mir.Module){}

func (self *_DeadCodeElimination) Run(ir *mir.Module){
	for cursor:=ir.Globals().Front(); cursor!=nil; cursor=cursor.Next(){
		self.walkGlobal(cursor.Value)
	}
}

func (self *_DeadCodeElimination) walkGlobal(ir mir.Global){
	switch global := ir.(type) {
	case *mir.Function:
		function.Run(global, function.DeadCodeElimination)
	}
}
